package test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const (
	baseAPI = "http://localhost:3001/api"
)

func TestCreateFriends(t *testing.T) {
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

	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != `{"success":true}` {
		t.Errorf("expecting "+`{"success":true}`+" but have %v", string(body))
	}
}
