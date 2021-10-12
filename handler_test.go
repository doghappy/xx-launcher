package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func init() {
	readConfig()
	startProcess = func(dir string, name string) (*os.Process, error) {
		return nil, nil
	}
}

func TestStart(t *testing.T) {
	var tests = []struct {
		AssertContent bool
		ReqContent    string
		ResCode       int
		ResContent    string
	}{
		{
			AssertContent: true,
			ReqContent:    "",
			ResCode:       http.StatusBadRequest,
			ResContent:    "EOF\n",
		},
		{
			ReqContent: "test",
			ResCode:    http.StatusBadRequest,
		},
		{
			AssertContent: true,
			ReqContent:    `{"id":2}`,
			ResCode:       http.StatusBadRequest,
			ResContent:    "invalid regionId: 0\n",
		},
		{
			AssertContent: true,
			ReqContent:    `{"regionId":2}`,
			ResCode:       http.StatusBadRequest,
			ResContent:    "invalid regionId: 2\n",
		},
		{
			AssertContent: true,
			ReqContent:    `{"regionId":1}`,
			ResCode:       http.StatusOK,
			ResContent:    "ok\n",
		},
	}

	for _, test := range tests {
		router := httprouter.New()
		router.POST("/start", startHandler)
		body := strings.NewReader(test.ReqContent)
		req, err := http.NewRequest(http.MethodPost, "/start", body)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		bytes, err := json.Marshal(test)
		if err != nil {
			t.Fatal(err)
		}
		testcase := string(bytes)

		if test.ResCode != rr.Code {
			t.Errorf("expected resCode is '%d', actual resCode is '%d'.\ntest case: %s\n", test.ResCode, rr.Code, testcase)
		}

		if test.AssertContent {
			resContent := rr.Body.String()
			if test.ResContent != resContent {
				t.Errorf("expected resContent is '%s', actual resContent is '%s'.\ntest case: %s\n", test.ResContent, resContent, testcase)
			}
		}
	}
}
