package annotations

import (
	"context"
	"encoding/json"

	"net/http"

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

	w.Header().Add("Content-Type", "application/json")
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

	p, err := h.fotowareAPI.Search(keywords)

	if err != nil {
		writeMessage(w, err.Error(), 500)
	}

	outBody, err := buildBody(p)

	if err != nil {
		writeMessage(w, err.Error(), 500)
	}
	w.Write(outBody)
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

type Result struct {
	Url         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
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
	w.Header().Add("Content-Type", "application/json")
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

	p, err := h.fotowareAPI.Search(keywords)

	if err != nil {
		writeMessage(w, err.Error(), 500)
	}

	outBody, err := buildBody(p)

	if err != nil {
		writeMessage(w, err.Error(), 500)
	}
	w.Write(outBody)
}

func buildBody(p fotoware.Payload) ([]byte, error) {
	results := []Result{}

	for _, d := range p.Data {
		var href string
		res := Result{}
		for _, p := range d.Previews {
			if p.Size == 2400 {
				href = p.Href
			}

		}
		res.Url = "https://fotoware-test.ft.com" + href
		for _, f := range d.BuiltinFields {
			if f.Field == "title" {
				res.Title = f.Value.(string)
			}
			if f.Field == "description" {
				res.Description = f.Value.(string)
			}
			if f.Field == "tags" {
				tags := []string{}
				for _, t := range f.Value.([]interface{}) {
					tags = append(tags, t.(string))
				}
				res.Tags = tags
			}
		}

		results = append(results, res)
	}

	return json.Marshal(results)
}
