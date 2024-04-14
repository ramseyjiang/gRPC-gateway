package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

type Race struct {
	ID                  string `json:"id"`
	MeetingID           string `json:"meetingId"`
	Name                string `json:"name"`
	Number              string `json:"number"`
	Visible             bool   `json:"visible"`
	AdvertisedStartTime string `json:"advertisedStartTime"`
}

type listRacesResponse struct {
	Races []Race `json:"races"`
}

type listRacesTestCase struct {
	name        string
	url         string
	filter      map[string]interface{}
	expectedLen int
}

const (
	apiHost   = "http://localhost:8000/"
	caseName1 = "No filtered"
)

func TestListRaces(t *testing.T) {
	tests := []listRacesTestCase{
		{
			name:        caseName1,
			url:         apiHost + "v1/list-races",
			filter:      map[string]interface{}{},
			expectedLen: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"filter": tt.filter}
			resp, err := makePostRequest(tt.url, data)
			if err != nil {
				t.Error(err)
				return
			}

			if tt.name == caseName1 && tt.expectedLen != len(resp.Races) {
				t.Errorf("Unexpected unfiltered response length: %d (expected %d)", len(resp.Races), tt.expectedLen)
				return
			}
		})
	}
}

func makePostRequest(url string, requestBody interface{}) (*listRacesResponse, error) {
	// Marshal the request body to JSON bytes
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Create the request body reader
	requestBodyReader := bytes.NewReader(requestBodyJSON)

	// Execute the POST request
	resp, err := http.Post(url, "application/json", requestBodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Assert on the response status code (adjust based on expected behavior)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code")
	}

	var listResp *listRacesResponse
	if err = json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, errors.New("failed to decode JSON response")
	}

	return listResp, nil
}
