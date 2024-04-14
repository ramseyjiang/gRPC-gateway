package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
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

type getRaceResponse struct {
	Race Race `json:"race"` // Adjust tag if the JSON field name is different
}

type getRaceTestCase struct {
	name       string
	url        string
	statusCode int
	expected   *racing.Race
	errMessage string
}

const (
	apiHost              = "http://localhost:8000/"
	listTestCaseName1    = "No filtered"
	listTestCaseName2    = "Filtered visible true"
	listTestCaseName3    = "Filtered visible and meeting_ids"
	listTestCaseName4    = "Filtered visible true and advertised_start_time order by asc"
	listTestCaseName5    = "Filtered visible true and advertised_start_time order by desc"
	listTestCaseName6    = "Filtered visible true, advertised_start_time order by desc, all status is CLOSED"
	getRaceTestCaseName1 = "Success: Valid ID"
	getRaceTestCaseName2 = "Empty: Non-existent ID"
)

var meetingIDs = []int{3, 8}

func TestListRaces(t *testing.T) {
	tests := []listRacesTestCase{
		{
			name:        listTestCaseName1,
			url:         apiHost + "v1/list-races",
			filter:      map[string]interface{}{},
			expectedLen: 100,
		},
		{
			name:        listTestCaseName2,
			url:         apiHost + "v1/list-races",
			filter:      map[string]interface{}{"visible": true},
			expectedLen: 54,
		},
		{
			name: listTestCaseName3,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":     true,
				"meeting_ids": meetingIDs,
			},
			expectedLen: 14,
		},
		{
			name: listTestCaseName4,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "asc",
			},
			expectedLen: 54,
		},
		{
			name: listTestCaseName5,
			url:  apiHost + "v1/list-races",
			filter: map[string]interface{}{
				"visible":  true,
				"column":   "advertised_start_time",
				"order_by": "desc",
			},
			expectedLen: 54,
		},
		{
			name: listTestCaseName6,
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

			if tt.name == listTestCaseName1 && tt.expectedLen != len(resp.Races) {
				t.Errorf("Unexpected unfiltered response length: %d (expected %d)", len(resp.Races), tt.expectedLen)
				return
			}

			if tt.name == listTestCaseName2 && tt.expectedLen == len(resp.Races) {
				for _, v := range resp.Races {
					if v.Visible == false {
						t.Errorf("Unexpected filtered response visible: %v (expected %v)", v.Visible, true)
						return
					}
				}
			}

			// fmt.Println(tt.name, len(resp.Races))
			if tt.name == listTestCaseName3 && tt.expectedLen == len(resp.Races) {
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

			if tt.name == listTestCaseName4 && tt.expectedLen == len(resp.Races) {
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

			if tt.name == listTestCaseName5 && tt.expectedLen == len(resp.Races) {
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

			if tt.name == listTestCaseName6 {
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

func TestGetRace(t *testing.T) {
	tests := []getRaceTestCase{
		{
			name:       getRaceTestCaseName1,
			url:        apiHost + "v1/race?id=57",
			statusCode: http.StatusOK,
			expected: &racing.Race{
				Id:        57,
				MeetingId: 8,
				Name:      "Virginia wolves",
				Number:    11,
				Visible:   true,
				Status:    "CLOSED",
			},
			errMessage: "",
		},
		{
			name:       getRaceTestCaseName2,
			url:        apiHost + "v1/race?id=999",
			statusCode: http.StatusOK, // Adjust based on actual error code
			expected: &racing.Race{
				Id:        0,
				MeetingId: 0,
				Name:      "",
				Number:    0,
				Visible:   false,
				Status:    "",
			},
			errMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Make the HTTP GET request
			resp, err := http.Get(tc.url)

			// Verify error message if expected
			if tc.errMessage != "" {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			// Check for unexpected errors
			if err != nil {
				t.Fatalf("Failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			// Verify response status code
			if resp.StatusCode != tc.statusCode {
				t.Errorf("Unexpected status code: %d (expected %d)", resp.StatusCode, tc.statusCode)
				return
			}

			// Decode the response
			var raceResp getRaceResponse
			if err := json.NewDecoder(resp.Body).Decode(&raceResp); err != nil {
				t.Fatalf("Failed to decode JSON response: %v", err)
			}

			// Extract and convert the Race object
			raceID, err := strconv.ParseInt(raceResp.Race.ID, 10, 64)
			if err != nil {
				t.Errorf("Error converting ID: %v", err)
				return
			}

			meetingID, err := strconv.ParseInt(raceResp.Race.MeetingID, 10, 64)
			if err != nil {
				t.Errorf("Error converting MeetingID: %v", err)
				return
			}

			number, err := strconv.ParseInt(raceResp.Race.Number, 10, 64)
			if err != nil {
				t.Errorf("Error converting Number: %v", err)
				return
			}

			race := racing.Race{
				Id:        raceID,
				MeetingId: meetingID,
				Name:      raceResp.Race.Name,
				Number:    number,
				Visible:   raceResp.Race.Visible,
				Status:    raceResp.Race.Status,
			}

			// Compare actual and expected races using the compareRaces function
			if !compareRaces(t, &race, tc.expected) {
				t.Errorf("Races do not match")
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

func compareRaces(t *testing.T, actual *racing.Race, expected *racing.Race) bool {
	if actual == nil && expected == nil {
		return true // Both are nil, so considered equal
	}

	if actual == nil || expected == nil {
		t.Errorf("Race objects mismatch: Actual: %v, Expected: %v", actual, expected)
		return false
	}

	if actual.Id != expected.Id {
		t.Errorf("Race ID mismatch: Actual: %d, Expected: %d", actual.Id, expected.Id)
		return false
	}

	if actual.MeetingId != expected.MeetingId {
		t.Errorf("Race MeetingID mismatch: Actual: %d, Expected: %d", actual.MeetingId, expected.MeetingId)
		return false
	}

	if actual.Name != expected.Name {
		t.Errorf("Race Name mismatch: Actual: %s, Expected: %s", actual.Name, expected.Name)
		return false
	}

	if actual.Number != expected.Number {
		t.Errorf("Race Number mismatch: Actual: %d, Expected: %d", actual.Number, expected.Number)
		return false
	}

	if actual.Visible != expected.Visible {
		t.Errorf("Race Visible mismatch: Actual: %v, Expected: %v", actual.Visible, expected.Visible)
		return false
	}

	if actual.Status != expected.Status {
		t.Errorf("Race Status mismatch: Actual: %s, Expected: %s", actual.Status, expected.Status)
		return false
	}

	return true // All fields match
}
