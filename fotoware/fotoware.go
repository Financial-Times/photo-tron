package fotoware

import (
	"net/http"
	"strings"
)

type FotowareAPI struct {
	apiKey string
	client *http.Client
}

func NewFotowareAPI(apiKey string) *FotowareAPI {
	return &FotowareAPI{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (f *FotowareAPI) Search(keywords []string) (*http.Response, error) {

	url := "https://fotoware-test.ft.com/fotoweb/archives/5003-Latest%20Images/?q="
	for _, k := range keywords {
		fk := strings.Replace(strings.ToLower(k), " ","%20",-1)
		url += "120%3A" + fk + "+OR+"
	}

	url = strings.TrimSuffix(url,"+OR+")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", `application/vnd.fotoware.assetlist+json`)
	req.Header.Add("fwapitoken", f.apiKey)
	return f.client.Do(req)
}
