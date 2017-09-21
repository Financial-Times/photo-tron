package annotations

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"bytes"
	"io/ioutil"

	"github.com/Financial-Times/photo-tron/fotoware"
	"github.com/Financial-Times/photo-tron/suggest"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/husobee/vestigo"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	annotationsAPI AnnotationsAPI
	fotowareAPI    *fotoware.FotowareAPI
}

func NewHandler(api AnnotationsAPI, fwAPI *fotoware.FotowareAPI) *Handler {
	return &Handler{api, fwAPI}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uuid := vestigo.Param(r, "uuid")
	tID := tidutils.GetTransactionIDFromRequest(r)
	ctx := tidutils.TransactionAwareContext(context.Background(), tID)
	ann, err := h.annotationsAPI.Get(ctx, uuid)
	if err != nil {
		log.WithError(err).WithField(tidutils.TransactionIDKey, tID).WithField("uuid", uuid).Error("Error in calling UPP Public Annotations API")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keywords := []string{}

	for _, a := range ann {
		if a.Predicate == "http://www.ft.com/ontology/annotation/majorMentions" ||
			a.Predicate == "http://www.ft.com/ontology/annotation/mentions" ||
			a.Predicate == "http://www.ft.com/ontology/annotation/about" {
			keywords = append(keywords, a.PrefLabel)
		}
	}

	resp, err := h.fotowareAPI.Search(keywords)

	defer resp.Body.Close()

	w.Header().Add("Content-Type", "application/json")
	if resp.StatusCode == http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		reader := bytes.NewReader(respBody)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, reader)
		return
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else {
		writeMessage(w, "Service unavailable", http.StatusServiceUnavailable)
	}
}

func writeMessage(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)

	message := make(map[string]interface{})
	message["message"] = msg
	j, err := json.Marshal(&message)

	if err != nil {
		log.WithError(err).Warn("Failed to parse provided message to json, this is a bug.")
		return
	}

	w.Write(j)
}

type SuggestHandler struct {
	suggestAPI  *suggest.SuggestAPI
	fotowareAPI *fotoware.FotowareAPI
}

func NewSuggestHandler(suggestAPI *suggest.SuggestAPI, fwAPI *fotoware.FotowareAPI) *SuggestHandler {
	return &SuggestHandler{suggestAPI, fwAPI}
}

func (h *SuggestHandler) SuggestServeHTTP(w http.ResponseWriter, r *http.Request) {
	tID := tidutils.GetTransactionIDFromRequest(r)

	body, err := ioutil.ReadAll(r.Body)

	suggestions, err := h.suggestAPI.Search(body)
	if err != nil {
		log.WithError(err).WithField(tidutils.TransactionIDKey, tID).Error("Error in calling UPP Public Annotations API")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keywords := []string{}

	for _, s := range suggestions.Suggestions {
		if s.Thing.PrefLabel != "" {
			keywords = append(keywords, s.Thing.PrefLabel)
		}
	}

	resp, err := h.fotowareAPI.Search(keywords)

	defer resp.Body.Close()

	w.Header().Add("Content-Type", "application/json")
	if resp.StatusCode == http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		reader := bytes.NewReader(respBody)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, reader)
		return
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else {
		writeMessage(w, "Service unavailable", http.StatusServiceUnavailable)
	}
}
