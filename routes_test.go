package test

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
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

type userActions struct {
	Requestor string `json:"requestor"`
	Target    string `json:"target"`
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

func TestGetCommonFriendsList(t *testing.T) {
	resetDB()
	// add friends
	addFriends := []url.Values{
		url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}},
		url.Values{"friends": []string{`["andy@example.com", "common@example.com"]`}},
		url.Values{"friends": []string{`["andy@example.com", "lisa@example.com"]`}},
		url.Values{"friends": []string{`["andy@example.com", "sean@example.com"]`}},
		url.Values{"friends": []string{`["john@example.com", "andy@example.com"]`}},
		url.Values{"friends": []string{`["john@example.com", "common@example.com"]`}},
		url.Values{"friends": []string{`["john@example.com", "lisa@example.com"]`}},
		url.Values{"friends": []string{`["lisa@example.com", "sean@example.com"]`}},
	}
	for _, addFriend := range addFriends {
		// errors are not checked as these are tested in TestCreateFriends test
		reqBody := addFriend.Encode()
		req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		http.DefaultClient.Do(req)
	}

	// get common friends
	testCases := []testStruct{}
	testUsers := []map[string]interface{}{
		// andy and others
		{ // andy - john common friends
			"friends":       []string{"andy@example.com", "john@example.com"},
			"commonFriends": []string{"common@example.com", "lisa@example.com"},
			"count":         2,
		},
		{ // andy - common common friends
			"friends":       []string{"andy@example.com", "common@example.com"},
			"commonFriends": []string{"john@example.com"},
			"count":         1,
		},
		{ // andy - lisa common friends
			"friends":       []string{"andy@example.com", "lisa@example.com"},
			"commonFriends": []string{"john@example.com", "sean@example.com"},
			"count":         2,
		},
		{ // andy - sean common friends
			"friends":       []string{"andy@example.com", "sean@example.com"},
			"commonFriends": []string{"lisa@example.com"},
			"count":         1,
		},
		// john and others
		{ // john - andy common friends - should be the same as andy - john
			"friends":       []string{"john@example.com", "andy@example.com"},
			"commonFriends": []string{"common@example.com", "lisa@example.com"},
			"count":         2,
		},
		{ // john - common common friends
			"friends":       []string{"john@example.com", "common@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		{ // john - lisa common friends
			"friends":       []string{"john@example.com", "lisa@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		{ // john - sean common friends
			"friends":       []string{"john@example.com", "sean@example.com"},
			"commonFriends": []string{"lisa@example.com", "andy@example.com"},
			"count":         2,
		},
		// common and others
		{ // common - andy common friends - should be the same as andy - common
			"friends":       []string{"common@example.com", "andy@example.com"},
			"commonFriends": []string{"john@example.com"},
			"count":         1,
		},
		{ // common - john common friends - should be the same as john - common
			"friends":       []string{"common@example.com", "john@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		{ // common - lisa common friends
			"friends":       []string{"common@example.com", "lisa@example.com"},
			"commonFriends": []string{"andy@example.com", "john@example.com"},
			"count":         2,
		},
		{ // common - sean common friends
			"friends":       []string{"common@example.com", "sean@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		// lisa and others
		{ // lisa - andy common friends - should be the same as andy - lisa
			"friends":       []string{"lisa@example.com", "andy@example.com"},
			"commonFriends": []string{"john@example.com", "sean@example.com"},
			"count":         2,
		},
		{ // lisa - john common friends - should be the same as john - common
			"friends":       []string{"lisa@example.com", "john@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		{ // lisa - common common friends - should be the same as common - lisa
			"friends":       []string{"lisa@example.com", "common@example.com"},
			"commonFriends": []string{"andy@example.com", "john@example.com"},
			"count":         2,
		},
		{ // lisa - sean common friends
			"friends":       []string{"lisa@example.com", "sean@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		// sean and others
		{ // sean - andy common friends - should be the same as sean - andy
			"friends":       []string{"sean@example.com", "andy@example.com"},
			"commonFriends": []string{"lisa@example.com"},
			"count":         1,
		},
		{ // sean - john common friends - should be the same as john - sean
			"friends":       []string{"sean@example.com", "john@example.com"},
			"commonFriends": []string{"lisa@example.com", "andy@example.com"},
			"count":         2,
		},
		{ // sean - common common friends - should be the same as common - sean
			"friends":       []string{"common@example.com", "sean@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
		{ // sean - lisa common friends - should be the same as lisa - sean
			"friends":       []string{"lisa@example.com", "sean@example.com"},
			"commonFriends": []string{"andy@example.com"},
			"count":         1,
		},
	}
	for _, testUser := range testUsers {
		friends := expectedResult{Friends: testUser["friends"].([]string)}
		jsonTestUser, err := json.Marshal(friends)
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(jsonTestUser),
			expectedResult: expectedResult{
				Success: testUser["count"].(int) > 0,
				Friends: testUser["commonFriends"].([]string),
				Count:   testUser["count"].(int),
			},
		})
	}

	for _, testCase := range testCases {
		req, err := http.NewRequest("GET", baseAPI+"/friends/common", strings.NewReader(testCase.stringRequestBody))
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

		sort.Strings(actualResult.Friends)
		sort.Strings(testCase.expectedResult.Friends)
		if strings.Join(actualResult.Friends, ",") != strings.Join(testCase.expectedResult.Friends, ",") {
			t.Errorf("expecting %v but have %v %v", testCase.expectedResult.Friends, actualResult.Friends, string(bodyBytes))
		}
		if actualResult.Count != testCase.expectedResult.Count {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Count, actualResult.Count)
		}
	}
}

func TestSubScribeUpdates(t *testing.T) {
	resetDB()
	testSubscribeSamples := []map[string]interface{}{
		{"json": userActions{Requestor: "lisa@example.com", Target: "john@example.com"}, "expectedResult": true},
		{"json": userActions{Requestor: "lisa@example.com", Target: "john@example.com"}, "expectedResult": false},
		{"json": userActions{Requestor: "lisa@example.com"}, "expectedResult": false},
		{"json": userActions{Target: "john@example.com"}, "expectedResult": false},
		{"json": userActions{}, "expectedResult": false},
	}
	testCases := []testStruct{}
	for _, testSubscribeSample := range testSubscribeSamples {
		json, err := json.Marshal(testSubscribeSample["json"])
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(json),
			expectedResult: expectedResult{
				Success: testSubscribeSample["expectedResult"].(bool),
			},
		})
	}

	// subscribe updates
	for _, testCase := range testCases {
		req, err := http.NewRequest("POST", baseAPI+"/friends/subscribe", strings.NewReader(testCase.stringRequestBody))
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

	// ensure no friends are created
	testUsers := []map[string]interface{}{
		{"email": "lisa@example.com", "friends": []string{}, "count": 0},
		{"email": "john@example.com", "friends": []string{}, "count": 0},
	}

	testCases = []testStruct{}
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

func TestBlockUpdates(t *testing.T) {
	resetDB()
	// block not connected users
	testSubscribeSamples := []map[string]interface{}{
		{"json": userActions{Requestor: "andy@example.com", Target: "john@example.com"}, "expectedResult": true},
		{"json": userActions{Requestor: "andy@example.com", Target: "john@example.com"}, "expectedResult": false},
		{"json": userActions{Requestor: "andy@example.com"}, "expectedResult": false},
		{"json": userActions{Target: "john@example.com"}, "expectedResult": false},
		{"json": userActions{}, "expectedResult": false},
	}

	testCases := []testStruct{}
	for _, testSubscribeSample := range testSubscribeSamples {
		json, err := json.Marshal(testSubscribeSample["json"])
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(json),
			expectedResult: expectedResult{
				Success: testSubscribeSample["expectedResult"].(bool),
			},
		})
	}

	for _, testCase := range testCases {
		req, err := http.NewRequest("POST", baseAPI+"/friends/block", strings.NewReader(testCase.stringRequestBody))
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

	// ensure new friends conenction cannot be made
	// errors are skipped as they have been tested in the respecive tests
	friends := url.Values{"friends": []string{`["andy@example.com", "john@example.com"]`}}.Encode()
	req, err := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(friends))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	actualResult := expectedResult{}
	if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
		t.Errorf("failed to unmarshal test result %v", err)
	}
	if actualResult.Success != false {
		t.Errorf("expecting %v but have %v", false, actualResult.Success)
	}

	// ensure blocked target cannot subscribe to block requestor
	// errors are skipped as they have been tested in the respecive tests
	users := userActions{Requestor: "andy@example.com", Target: "john@example.com"}
	jsonUsers, err := json.Marshal(users)
	req, err = http.NewRequest("POST", baseAPI+"/friends/subscribe", strings.NewReader(string(jsonUsers)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = http.DefaultClient.Do(req)

	bodyBytes, _ = ioutil.ReadAll(res.Body)
	actualResult = expectedResult{}
	if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
		t.Errorf("failed to unmarshal test result %v", err)
	}
	if actualResult.Success != false {
		t.Errorf("expecting %v but have %v", false, actualResult.Success)
	}

	// add new friends to test blocking connected users
	// errors are skipped as they have been tested in the respective test
	newFriends := url.Values{"friends": []string{`["sean@example.com", "lisa@example.com"]`}}.Encode()
	req, err = http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(newFriends))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = http.DefaultClient.Do(req)

	// ensure new friends have been added successfully
	// errors are skipped as they have been tested in the respective test
	user := user{Email: "sean@example.com"}
	jsonUser, _ := json.Marshal(user)
	req, err = http.NewRequest("GET", baseAPI+"/friends", strings.NewReader(string(jsonUser)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = http.DefaultClient.Do(req)

	bodyBytes, _ = ioutil.ReadAll(res.Body)
	actualResult = expectedResult{}
	if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
		t.Errorf("failed to unmarshal test result %v", err)
	}
	if actualResult.Success != true {
		t.Errorf("expecting %v but have %v", true, actualResult.Success)
	}
	if strings.Join(actualResult.Friends, ",") != strings.Join([]string{"lisa@example.com"}, ",") {
		t.Errorf("expecting %v but have %v", []string{"lisa@example.com"}, actualResult.Friends)
	}
	if actualResult.Count != 1 {
		t.Errorf("expecting %v but have %v", 1, actualResult.Count)
	}

	// test block connected users
	users = userActions{Requestor: "sean@example.com", Target: "lisa@example.com"}
	jsonUsers, _ = json.Marshal(users)
	req, err = http.NewRequest("POST", baseAPI+"/friends/block", strings.NewReader(string(jsonUsers)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("expecting status code of 200 but have %v", res.StatusCode)
	}

	bodyBytes, _ = ioutil.ReadAll(res.Body)
	actualResult = expectedResult{}
	if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
		t.Errorf("failed to unmarshal test result %v", err)
	}
	if actualResult.Success != true {
		t.Errorf("expecting %v but have %v %v", true, actualResult.Success, string(bodyBytes))
	}

	// ensure blocked target is no longer a friend of the block requestor
	blockedUser := struct { // had to do it this way, go kept complaining user is not a type
		Email string `json:"email"`
	}{Email: "lisa@example.com"}
	jsonUser, _ = json.Marshal(blockedUser)
	req, err = http.NewRequest("GET", baseAPI+"/friends", strings.NewReader(string(jsonUser)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = http.DefaultClient.Do(req)

	bodyBytes, _ = ioutil.ReadAll(res.Body)
	actualResult = expectedResult{}
	if err := json.Unmarshal(bodyBytes, &actualResult); err != nil {
		t.Errorf("failed to unmarshal test result %v", err)
	}
	if actualResult.Success != false {
		t.Errorf("expecting %v but have %v", false, actualResult.Success)
	}
	if strings.Join(actualResult.Friends, ",") != strings.Join([]string{}, ",") {
		t.Errorf("expecting %v but have %v", []string{}, actualResult.Friends)
	}
	if actualResult.Count != 0 {
		t.Errorf("expecting %v but have %v", 0, actualResult.Count)
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
