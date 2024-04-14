package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type Race struct {
	ID                  string `json:"id"`
	MeetingID           string `json:"meetingId"`
	Name                string `json:"name"`
	Number              string `json:"number"`
	Visible             bool   `json:"visible"`
	AdvertisedStartTime string `json:"advertisedStartTime"`
	Status              string `json:"status"`
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
	caseName2 = "Filtered visible true"
	caseName3 = "Filtered visible and meeting_ids"
	caseName4 = "Filtered visible true and advertised_start_time order by asc"
	caseName5 = "Filtered visible true and advertised_start_time order by desc"
	caseName6 = "Filtered visible true, advertised_start_time order by desc, all status is CLOSED"
)

var meetingIDs = []int{3, 8}

func TestListRaces(t *testing.T) {
	tests := []listRacesTestCase{
		{
			name:        caseName1,
			url:         apiHost + "v1/list-races",
			filter:      map[string]interface{}{},
			expectedLen: 100,
		},
		{
			name:        caseName2,
			url:         apiHost + "v1/list-races",
			filter:      map[string]interface{}{"visible": true},
			expectedLen: 54,
		},
		{
			name: caseName3,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":     true,
				"meeting_ids": meetingIDs,
			},
			expectedLen: 14,
		},
		{
			name: caseName4,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "asc",
			},
			expectedLen: 54,
		},
		{
			name: caseName5,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "desc",
			},
			expectedLen: 54,
		},
		{
			name: caseName6,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "desc",
			},
			expectedLen: 54,
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

			if tt.name == caseName2 && tt.expectedLen == len(resp.Races) {
				for _, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}
				}
			}

			// fmt.Println(tt.name, len(resp.Races))
			if tt.name == caseName3 && tt.expectedLen == len(resp.Races) {
				for _, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// "slices.Contains" requires go1.18 or later (-lang was set to go1.16, so I create contains function below.
					mID, _ := strconv.Atoi(v.MeetingID)
					if !contains(meetingIDs, mID) {
						t.Errorf("Unexpected filtered response meeting ID is %v in : expected %v", mID, meetingIDs)
						return
					}
				}
			}

			if tt.name == caseName4 && tt.expectedLen == len(resp.Races) {
				for k, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// check all data is asc, compare current and next data's AdvertisedStartTime
					if k+1 != len(resp.Races) {
						time1, _ := time.Parse(time.RFC3339, v.AdvertisedStartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Races[k+1].AdvertisedStartTime)
						if time1.After(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Races[k+1].AdvertisedStartTime)
							return
						}
					}
				}
			}

			if tt.name == caseName5 && tt.expectedLen == len(resp.Races) {
				for k, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// check all data is desc, compare current and next data's AdvertisedStartTime
					if k+1 != len(resp.Races) {
						time1, _ := time.Parse(time.RFC3339, v.AdvertisedStartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Races[k+1].AdvertisedStartTime)
						if time1.Before(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Races[k+1].AdvertisedStartTime)
							return
						}
					}
				}
			}

			if tt.name == caseName6 {
				for k, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					if k+1 != len(resp.Races) {
						time1, _ := time.Parse(time.RFC3339, v.AdvertisedStartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Races[k+1].AdvertisedStartTime)
						if time1.Before(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Races[k+1].AdvertisedStartTime)
							return
						}
					}

					if v.Status != "CLOSED" {
						t.Errorf("Unexpected filtered response status: %v (expected %v)", v.Status, "CLOSED")
						return
					}
				}
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

func contains(meetingIDs []int, meetingID int) bool {
	for _, id := range meetingIDs {
		if id == meetingID {
			return true
		}
	}
	return false
}
