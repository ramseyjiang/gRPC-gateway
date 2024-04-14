package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

type Event struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Visible             bool   `json:"visible"`
	Result              string `json:"result"`
	Location            string `json:"location"`
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
	AdvertisedStartTime string `json:"advertisedStartTime"`
	Status              string `json:"status"`
}

type listEventsResponse struct {
	Events []Event `json:"events"`
}

type listEventsTestCase struct {
	name        string
	url         string
	filter      map[string]interface{}
	expectedLen int
}

const (
	apiHost   = "http://localhost:8000/"
	caseName1 = "No filtered"
	caseName2 = "Filtered visible true"
	caseName3 = "Filtered visible true and advertised_start_time order by asc"
	caseName4 = "Filtered visible true and start_time order by desc"
	caseName5 = "Filtered visible true and end_time order by desc"
	caseName6 = "Filtered advertised_start_time order by desc, status OPEN or CLOSED"
	caseName7 = "Filtered visible true, location exists, and start_time less than end_time"
	caseName8 = "Filtered visible true and name desc, and start_time less than end_time"
	caseName9 = "Filtered visible true and id equals 68"
)

func TestListEvents(t *testing.T) {
	tests := []listEventsTestCase{
		{
			name:        caseName1,
			url:         apiHost + "v1/list-events",
			filter:      map[string]interface{}{},
			expectedLen: 49,
		},
		{
			name:   caseName2,
			url:    apiHost + "v1/list-events",
			filter: map[string]interface{}{"visible": true},
		},
		{
			name: caseName3,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "asc",
			},
		},
		{
			name: caseName4,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "start_time",
				"order_by": "desc",
			},
		},
		{
			name: caseName5,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "end_time",
				"order_by": "desc",
			},
		},
		{
			name: caseName6,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"column":   "advertised_start_time",
				"order_by": "desc",
			},
		},
		{
			name: caseName7,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "location",
				"order_by": "desc",
			},
		},
		{
			name: caseName8,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "name",
				"order_by": "desc",
			},
		},
		{
			name: caseName9,
			url:  apiHost + "v1/list-events",
			filter: map[string]interface{}{
				"visible": true,
				"id":      68,
			},
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

			if tt.name == caseName1 && tt.expectedLen != len(resp.Events) {
				t.Errorf("Unexpected unfiltered response length: %d (expected %d)", len(resp.Events), tt.expectedLen)
				return
			}

			if tt.name == caseName2 {
				for _, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}
				}
			}

			if tt.name == caseName3 {
				for k, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					if k+1 != len(resp.Events) {
						time1, _ := time.Parse(time.RFC3339, v.AdvertisedStartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Events[k+1].AdvertisedStartTime)
						if time1.After(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Events[k+1].AdvertisedStartTime)
							return
						}
					}
				}
			}

			if tt.name == caseName4 {
				for k, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					if k+1 != len(resp.Events) {
						time1, _ := time.Parse(time.RFC3339, v.StartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Events[k+1].StartTime)
						if time1.Before(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Events[k+1].AdvertisedStartTime)
							return
						}
					}
				}
			}

			if tt.name == caseName5 {
				for k, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					if k+1 != len(resp.Events) {
						time1, _ := time.Parse(time.RFC3339, v.EndTime)
						time2, _ := time.Parse(time.RFC3339, resp.Events[k+1].EndTime)
						if time1.Before(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Events[k+1].AdvertisedStartTime)
							return
						}
					}
				}
			}

			if tt.name == caseName6 {
				for k, v := range resp.Events {
					if k+1 != len(resp.Events) {
						time1, _ := time.Parse(time.RFC3339, v.AdvertisedStartTime)
						time2, _ := time.Parse(time.RFC3339, resp.Events[k+1].AdvertisedStartTime)
						if time1.Before(time2) {
							t.Errorf("Unexpected filtered response advertised_start_time: %v (expected %v)", v.AdvertisedStartTime, resp.Events[k+1].AdvertisedStartTime)
							return
						}
					}

					if !(v.Status == "OPEN" || v.Status == "CLOSED") {
						t.Errorf("Unexpected filtered response status: %v (expected %v)", v.Status, "OPEN or CLOSED")
						return
					}
				}
			}

			if tt.name == caseName7 {
				for _, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// check the event special columns
					if v.Location == "" {
						t.Error("Unexpected filtered response columns")
						return
					}

					startTime, _ := time.Parse(time.RFC3339, v.StartTime)
					endTime, _ := time.Parse(time.RFC3339, v.EndTime)
					if startTime.After(endTime) {
						t.Errorf("Unexpected filtered response start_time: %v not less than end_time %v", startTime, endTime)
						return
					}
				}
			}

			if tt.name == caseName8 {
				for k, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// check the event special columns
					if v.Name == "" {
						t.Error("Unexpected filtered response columns")
						return
					}

					if k+1 != len(resp.Events) {
						name1 := v.Name
						name2 := resp.Events[k+1].Name
						if strings.ToLower(name1) < strings.ToLower(name2) {
							t.Errorf("Unexpected filtered response name1: %v not before name2 %v", name1, name2)
							return
						}
					}

					startTime, _ := time.Parse(time.RFC3339, v.StartTime)
					endTime, _ := time.Parse(time.RFC3339, v.EndTime)
					if startTime.After(endTime) {
						t.Errorf("Unexpected filtered response start_time: %v not less than end_time %v", startTime, endTime)
						return
					}
				}
			}

			if tt.name == caseName9 {
				for _, v := range resp.Events {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}

					// check the event special columns
					if v.ID != "68" {
						t.Errorf("Unexpected sports event id response columns, expected ID is %v, got ID is %v", 68, v.ID)
						return
					}
				}
			}
		})
	}
}

func makePostRequest(url string, requestBody interface{}) (*listEventsResponse, error) {
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

	var listResp *listEventsResponse
	if err = json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, errors.New("failed to decode JSON response")
	}

	return listResp, nil
}
