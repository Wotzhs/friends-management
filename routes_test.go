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

type testStruct struct {
	arrayRequestBody  url.Values
	stringRequestBody string
	expectedResult
}

type expectedResult struct {
	Success    bool     `json:"success"`
	Friends    []string `json:"friends"`
	Count      int      `json:"count"`
	Recipients []string `json:"recipients"`
}

type user struct {
	Email string `json:"email"`
}

type userActions struct {
	Requestor string `json:"requestor"`
	Target    string `json:"target"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
}

func init() {
	if os.Getenv("GO_ENV") == "test" {
		baseAPI = "http://localhost:3001/api"
	}
}

func TestCreateFriends(t *testing.T) {
	resetDB()
	testSamples := []map[string]interface{}{
		{
			"friends": []string{"andy@example.com", "john@example.com"},
			"success": true,
		},
		{ // duplicate request
			"friends": []string{"andy@example.com", "john@example.com"},
			"success": false,
		},
		{ // same user
			"friends": []string{"andy@example.com", "andy@example.com"},
			"success": false,
		},
		{ // insufficient user
			"friends": []string{"andy@example.com"},
			"success": false,
		},
		{ // invalid user format
			"friends": []string{"andy", "john"},
			"success": false,
		},
	}

	testCases := []testStruct{}
	for _, testSample := range testSamples {
		friends := expectedResult{Friends: testSample["friends"].([]string)}
		jsonTestUser, err := json.Marshal(friends)
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(jsonTestUser),
			expectedResult: expectedResult{
				Success: testSample["success"].(bool),
			},
		})
	}

	for _, testCase := range testCases {
		req, err := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(testCase.stringRequestBody))
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
	addFriends := []map[string]interface{}{
		{"friends": []string{"andy@example.com", "john@example.com"}},
		{"friends": []string{"andy@example.com", "lisa@example.com"}},
		{"friends": []string{"john@example.com", "kate@example.com"}},
	}
	for _, addFriend := range addFriends {
		// errors are not checked as these are tested in TestCreateFriends test
		json, _ := json.Marshal(expectedResult{Friends: addFriend["friends"].([]string)})
		req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(string(json)))
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
	addFriends := []map[string]interface{}{
		{"friends": []string{"andy@example.com", "john@example.com"}},
		{"friends": []string{"andy@example.com", "common@example.com"}},
		{"friends": []string{"andy@example.com", "lisa@example.com"}},
		{"friends": []string{"andy@example.com", "sean@example.com"}},
		{"friends": []string{"john@example.com", "andy@example.com"}},
		{"friends": []string{"john@example.com", "common@example.com"}},
		{"friends": []string{"john@example.com", "lisa@example.com"}},
		{"friends": []string{"lisa@example.com", "sean@example.com"}},
	}
	for _, addFriend := range addFriends {
		// errors are not checked as these are tested in TestCreateFriends test
		json, _ := json.Marshal(expectedResult{Friends: addFriend["friends"].([]string)})
		req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(string(json)))
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
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Friends, actualResult.Friends)
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
	jsonUsers, _ := json.Marshal(expectedResult{Friends: []string{"andy@example.com", "john@example.com"}})
	req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(string(jsonUsers)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)

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
	jsonUsers, _ = json.Marshal(userActions{Requestor: "andy@example.com", Target: "john@example.com"})
	req, _ = http.NewRequest("POST", baseAPI+"/friends/subscribe", strings.NewReader(string(jsonUsers)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ = http.DefaultClient.Do(req)

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
	jsonUsers, _ = json.Marshal(expectedResult{Friends: []string{"sean@example.com", "lisa@example.com"}})
	req, _ = http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(string(jsonUsers)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ = http.DefaultClient.Do(req)

	// ensure new friends have been added successfully
	// errors are skipped as they have been tested in the respective test
	jsonUser, _ := json.Marshal(user{Email: "sean@example.com"})
	req, _ = http.NewRequest("GET", baseAPI+"/friends", strings.NewReader(string(jsonUser)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ = http.DefaultClient.Do(req)

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
	jsonUsers, _ = json.Marshal(userActions{Requestor: "sean@example.com", Target: "lisa@example.com"})
	req, err := http.NewRequest("POST", baseAPI+"/friends/block", strings.NewReader(string(jsonUsers)))
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
	jsonUser, _ = json.Marshal(user{Email: "lisa@example.com"})
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

func TestGetSubscribersList(t *testing.T) {
	resetDB()
	// add connections & subscribers
	// err and result checks are omitted intentionally
	newFriends := []map[string]interface{}{
		{"friends": []string{"andy@example.com", "john@example.com"}},
		{"friends": []string{"lisa@example.com", "john@example.com"}},
	}
	for _, newFriend := range newFriends {
		json, _ := json.Marshal(expectedResult{Friends: newFriend["friends"].([]string)})
		req, _ := http.NewRequest("POST", baseAPI+"/friends", strings.NewReader(string(json)))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		http.DefaultClient.Do(req)
	}

	newSubscriber := userActions{Requestor: "sean@example.com", Target: "john@example.com"}
	jsonSubscriber, _ := json.Marshal(newSubscriber)
	req, _ := http.NewRequest("POST", baseAPI+"/friends/subscribe", strings.NewReader(string(jsonSubscriber)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	http.DefaultClient.Do(req)

	testSamples := []map[string]interface{}{
		{ // with valid data
			"test":       userActions{Sender: "john@example.com", Text: "Hello World! kate@example.com"},
			"success":    true,
			"recipients": []string{"andy@example.com", "lisa@example.com", "sean@example.com", "kate@example.com"},
		},
		{ // with empty message
			"test":       userActions{Sender: "john@example.com"},
			"success":    true,
			"recipients": []string{"andy@example.com", "lisa@example.com", "sean@example.com"},
		},
		{ // with message without mentions
			"test":       userActions{Sender: "john@example.com", Text: "Hello World!"},
			"success":    true,
			"recipients": []string{"andy@example.com", "lisa@example.com", "sean@example.com"},
		},
		{ // with multiple mentions in message
			"test":       userActions{Sender: "john@example.com", Text: "Hello World! kate@example.com, cathy@example.com and someone@example.com"},
			"success":    true,
			"recipients": []string{"andy@example.com", "lisa@example.com", "sean@example.com", "kate@example.com", "cathy@example.com", "someone@example.com"},
		},
		{ // with invalid mention in message
			"test":       userActions{Sender: "john@example.com", Text: "Hello World! kate@exam@ple.com"},
			"success":    true,
			"recipients": []string{"andy@example.com", "lisa@example.com", "sean@example.com"},
		},
		{ // without subscriber but valid mention in message
			"test":       userActions{Sender: "someone@example.com", Text: "Hello World! kate@example.com"},
			"success":    true,
			"recipients": []string{"kate@example.com"},
		},
		{ // without sender
			"test":       userActions{Text: "Hello World! kate@example.com"},
			"success":    false,
			"recipients": []string{},
		},
		{ // blank
			"test":       userActions{},
			"success":    false,
			"recipients": []string{},
		},
	}

	testCases := []testStruct{}
	for _, testSample := range testSamples {
		jsonTest, err := json.Marshal(testSample["test"])
		if err != nil {
			t.Error(err)
		}
		testCases = append(testCases, testStruct{
			stringRequestBody: string(jsonTest),
			expectedResult: expectedResult{
				Success:    testSample["success"].(bool),
				Recipients: testSample["recipients"].([]string),
			},
		})
	}

	for _, testCase := range testCases {
		req, err := http.NewRequest("GET", baseAPI+"/friends/subscribe", strings.NewReader(testCase.stringRequestBody))
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

		sort.Strings(actualResult.Recipients)
		sort.Strings(testCase.expectedResult.Recipients)
		if strings.Join(actualResult.Recipients, ",") != strings.Join(testCase.expectedResult.Recipients, ",") {
			t.Errorf("expecting %v but have %v", testCase.expectedResult.Recipients, actualResult.Recipients)
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
