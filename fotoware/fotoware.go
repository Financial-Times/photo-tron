package fotoware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (f *FotowareAPI) Search(keywords []string) (Payload, error) {

	url := "https://fotoware-test.ft.com/fotoweb/archives/5003-Latest%20Images/?q="
	for _, k := range keywords {
		fk := strings.Replace(strings.ToLower(k), " ", "%20", -1)
		url += "120%3A" + fk + "+OR+"
	}

	url = strings.TrimSuffix(url, "+OR+")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Payload{}, err
	}

	req.Header.Add("accept", `application/vnd.fotoware.assetlist+json`)
	req.Header.Add("fwapitoken", f.apiKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return Payload{}, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return Payload{}, err
	}

	var p Payload
	err = json.Unmarshal(b, &p)

	if err != nil {
		return Payload{}, err
	}

	fmt.Printf("%+v", p)

	return p, nil
}

type Payload struct {
	Data []struct {
		Previews []struct {
			Size   int    `json:"size"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			Href   string `json:"href"`
			Square bool   `json:"square"`
		} `json:"previews"`
		BuiltinFields []struct {
			Field    string      `json:"field"`
			Required bool        `json:"required"`
			Value    interface{} `json:"value"`
		} `json:"builtinFields"`
	} `json:"data"`
}
