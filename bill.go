package main

import (
	"net/http"
	"encoding/json"
	"errors"
)

var LatestSenateBillID = ""
var LatestHouseBillID = ""
const ProPublicaKeyHeader = "X-API-Key"
const ProPublicaOKStatus = "OK"

func getNewSenateBills(apiKey string, session string) ([]Bill, error) {
	req, err := http.NewRequest("GET", "https://api.propublica.org/congress/v1/" + session +"/senate/bills/updated.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(ProPublicaKeyHeader, apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	billsRes := BillsRes{}

	err = decoder.Decode(&billsRes)
	if err != nil {
		return nil, err
	}

	if billsRes.Status != ProPublicaOKStatus {
		return nil, errors.New("Senate bill GET failed: " + billsRes.Status)
	}

	if len(billsRes.Results) < 0 || len(billsRes.Results[0].Bills) < 0 {
		return nil, errors.New("Missing Senate Bills")
	}

	if LatestSenateBillID == "" {
		LatestSenateBillID = billsRes.Results[0].Bills[0].BillID
	}

	recentBills := []Bill{}

	for _, bill := range billsRes.Results[0].Bills {
		if bill.BillID != LatestSenateBillID {
			recentBills = append(recentBills, bill)
		} else {
			break
		}
	}

	return recentBills, nil
}

func getNewHouseBills(apiKey string, session string) ([]Bill, error) {
	req, err := http.NewRequest("GET", "https://api.propublica.org/congress/v1/" + session +"/house/bills/updated.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(ProPublicaKeyHeader, apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	billsRes := BillsRes{}

	err = decoder.Decode(&billsRes)
	if err != nil {
		return nil, err
	}

	if billsRes.Status != ProPublicaOKStatus {
		return nil, errors.New("House bill GET failed: " + billsRes.Status)
	}

	if len(billsRes.Results) < 0 || len(billsRes.Results[0].Bills) < 0 {
		return nil, errors.New("Missing House Bills")
	}

	if LatestHouseBillID == "" {
		LatestHouseBillID = billsRes.Results[0].Bills[0].BillID
	}

	recentBills := []Bill{}

	for _, bill := range billsRes.Results[0].Bills {
		if bill.BillID != LatestHouseBillID {
			recentBills = append(recentBills, bill)
		} else {
			break
		}
	}

	return recentBills, nil
}

type BillsRes struct {
	Status string `json:"status"`
	Copyright string `json:"copyright"`
	Results []struct {
		Congress string `json:"congress"`
		Chamber string `json:"chamber"`
		NumResults string `json:"num_results"`
		Offset string `json:"offset"`
		Bills []Bill `json:"bills"`
	} `json:"results"`
}

type Bill struct {
	BillID string `json:"bill_id"`
	BillType string `json:"bill_type"`
	Number string `json:"number"`
	BillURI string `json:"bill_uri"`
	Title string `json:"title"`
	SponsorID string `json:"sponsor_id"`
	SponsorURI string `json:"sponsor_uri"`
	GpoPdfURI string `json:"gpo_pdf_uri"`
	CongressdotgovURL string `json:"congressdotgov_url"`
	GovtrackURL string `json:"govtrack_url"`
	IntroducedDate string `json:"introduced_date"`
	Active string `json:"active"`
	HousePassage string `json:"house_passage"`
	SenatePassage string `json:"senate_passage"`
	Enacted string `json:"enacted"`
	Vetoed string `json:"vetoed"`
	Cosponsors string `json:"cosponsors"`
	Committees string `json:"committees"`
	PrimarySubject string `json:"primary_subject"`
	Summary string `json:"summary"`
	SummaryShort string `json:"summary_short"`
	LatestMajorActionDate string `json:"latest_major_action_date"`
	LatestMajorAction string `json:"latest_major_action"`
}