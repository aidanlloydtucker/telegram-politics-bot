package main

import (
	"net/http"
	"encoding/json"
	"errors"
)

var LatestExecutiveOrder = 0

func getNewExecutiveOrders() ([]ExecutiveOrder, error) {
	req, err := http.NewRequest("GET", "https://www.federalregister.gov/api/v1/documents.json", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("conditions[correction]", "0")
	q.Add("conditions[president]", "donald-trump")
	q.Add("conditions[presidential_document_type_id]", "2")
	q.Add("conditions[type]", "PRESDOCU")
	q.Add("fields[]", "executive_order_notes")
	q.Add("fields[]", "executive_order_number")
	q.Add("fields[]", "html_url")
	q.Add("fields[]", "publication_date")
	q.Add("fields[]", "signing_date")
	q.Add("fields[]", "title")
	q.Add("per_page", "10")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	eoRes := ExecutiveOrderRes{}

	err = decoder.Decode(&eoRes)
	if err != nil {
		return nil, err
	}

	if len(eoRes.Results) < 0 {
		return nil, errors.New("Missing Executive Orders")
	}

	if LatestExecutiveOrder == 0 {
		LatestExecutiveOrder = eoRes.Results[0].ExecutiveOrderNumber
	}

	recentEOs := []ExecutiveOrder{}

	for _, eo := range eoRes.Results {
		if eo.ExecutiveOrderNumber > LatestExecutiveOrder {
			recentEOs = append(recentEOs, eo)
		}
	}

	return recentEOs, nil
}

type ExecutiveOrderRes struct {
	Count int `json:"count"`
	Description string `json:"description"`
	TotalPages int `json:"total_pages"`
	NextPageUrl string `json:"next_page_url"`
	Results []ExecutiveOrder `json:"results"`
}

type ExecutiveOrder struct {
	Title string `json:"title"`
	ExecutiveOrderNumber int `json:"executive_order_number"`
	SigningDate string `json:"signing_date"`
	PublicationDate string `json:"publication_date"`
	ExecutiveOrderNotes string `json:"executive_order_notes"`
	HTMLUrl string `json:"html_url"`
}