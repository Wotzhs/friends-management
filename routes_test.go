package test

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

var (
	baseAPI = "http://localhost:3000/api"
)

type user struct {
	Email string `json:"email"`
}

type expectedResult struct {
	Success bool     `json:"success"`
	Friends []string `json:"friends"`
	Count   int      `json:"count"`
}

type testStruct struct {
	arrayRequestBody  url.Values
	stringRequestBody string
	expectedResult
}

func init() {
	if os.Getenv("GO_ENV") == "test" {
		baseAPI = "http://localhost:3001/api"
	}
}

func TestCreateFriends(t *testing.T) {
	resetDB()
	testCases := []testStruct{
		{
			arrayRequestBody: url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
			expectedResult:   expectedResult{Success: true},
		},
		{ // duplicate request
			arrayRequestBody: url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
			expectedResult:   expectedResult{Success: false},
		},
		{ // same user
			arrayRequestBody: url.Values{"friends": []string{`["andy@example.com", "andy@example.com"]`}},
			expectedResult:   expectedResult{Success: false},
		},
		{ // insufficient user
			arrayRequestBody: url.Values{"friends": []string{`["andy@example.com"]`}},
			expectedResult:   expectedResult{Success: false},
		},
		{ // invalid user format
			arrayRequestBody: url.Values{"friends": []string{`["andy", "john"]`}},
			expectedResult:   expectedResult{Success: false},
		},
	}

	for _, testCase := range testCases {
		reqBody := testCase.arrayRequestBody.Encode()
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
		if actualResult.Success != testCase.expectedResult.Success {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Success, actualResult.Success)
		}
	}
}

func TestGetFriendsList(t *testing.T) {
	resetDB()
	// add friends
	addFriends := []url.Values{
		url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
		url.Values{"friends": []string{`["andy@example.com", "lisa@example.com"]`}},
		url.Values{"friends": []string{`["john@example.com", "kate@example.com"]`}},
	}
	for _, addFriend := range addFriends {
		// errors are not checked as these are tested in TestCreateFriends test
		reqBody := addFriend.Encode()
		req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		http.DefaultClient.Do(req)
	}

	// get friends
	testCases := []testStruct{}
	testUsers := []map[string]interface{}{
		{"email": "andy@example.com", "friends": []string{"john@example.com", "lisa@example.com"}, "count": 2},
		{"email": "john@example.com", "friends": []string{"andy@example.com", "kate@example.com"}, "count": 2},
		{"email": "lisa@example.com", "friends": []string{"andy@example.com"}, "count": 1},
		{"email": "sean@example.com", "friends": []string{}, "count": 0},
	}
	for _, testUser := range testUsers {
		user := user{testUser["email"].(string)}
		jsonTestUser, err := json.Marshal(user)
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(jsonTestUser),
			expectedResult: expectedResult{
				Success: testUser["count"].(int) > 0,
				Friends: testUser["friends"].([]string),
				Count:   testUser["count"].(int),
			},
		})
	}

	for _, testCase := range testCases {

		req, err := http.NewRequest("GET", baseAPI+"/friends", strings.NewReader(testCase.stringRequestBody))
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
		if actualResult.Success != testCase.expectedResult.Success {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Success, actualResult.Success)
		}
		if strings.Join(actualResult.Friends, ",") != strings.Join(testCase.expectedResult.Friends, ",") {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Friends, actualResult.Friends)
		}
		if actualResult.Count != testCase.expectedResult.Count {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Count, actualResult.Count)
		}
	}
}

func resetDB() {
	conninfo := "user=postgres host=db sslmode=disable dbname=friends_management_test"
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatalf("error in db connection info %+v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("error in pinging db %v", err)
	}
	db.Exec("DELETE FROM relationships")
}
