package suggest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SuggestionBody struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Thing struct {
		ID        string   `json:"id"`
		PrefLabel string   `json:"prefLabel"`
		Types     []string `json:"types"`
	} `json:"thing"`
	Provenances []struct {
		Scores []struct {
			ScoringSystem string  `json:"scoringSystem"`
			Value         float64 `json:"value"`
		} `json:"scores"`
	} `json:"provenances"`
}

type SuggestAPI struct {
	apiKey string
	client *http.Client
}

func NewSuggestAPI(apiKey string) *SuggestAPI {
	return &SuggestAPI{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (f *SuggestAPI) Search(body []byte) (*SuggestionBody, error) {

	url := "https://api.ft.com/suggest?apiKey=" + f.apiKey

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", `application/json`)
	req.Header.Add("accept", `application/json`)

	res, err := f.client.Do(req)

	if err != nil {
		return &SuggestionBody{}, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return &SuggestionBody{}, err
	}

	var sug SuggestionBody
	err = json.Unmarshal(b, &sug)

	if err != nil {
		return &SuggestionBody{}, err
	}

	fmt.Printf("%+v", sug)

	return &sug, nil
}
