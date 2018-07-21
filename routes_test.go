package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const (
	baseAPI = "http://localhost:3001/api"
)

type expectedResult struct {
	Success bool `json:"success"`
}

type testStruct struct {
	requestBody url.Values
	expectedResult
}

func TestCreateFriends(t *testing.T) {
	testCases := []testStruct{
		{
			url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
			expectedResult{Success: true},
		},
		{ // duplicate request
			url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
			expectedResult{Success: false},
		},
		{ // same user
			url.Values{"friends": []string{`["andy@example.com", "andy@example.com"]`}},
			expectedResult{Success: false},
		},
		{ // insufficient user
			url.Values{"friends": []string{`["andy@example.com"]`}},
			expectedResult{Success: false},
		},
		{ // invalid user format
			url.Values{"friends": []string{`["andy", "john"]`}},
			expectedResult{Success: false},
		},
	}

	for _, testCase := range testCases {
		reqBody := url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}}.Encode()
		req, err := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}

		if res.StatusCode != 200 {
			t.Errorf("expecting status code of 200 but have %v", res.StatusCode)
		}

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		actualResult := expectedResult{}
		if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
			t.Errorf("failed to unmarshal test result %v", err)
		}
		if actualResult != testCase.expectedResult {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Success, actualResult.Success)
		}
	}
}
