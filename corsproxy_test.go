package corsproxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestValidateRequest(t *testing.T) {
	const (
		goodUrl     = "http:/test.com/?https://what.the.what"
		badUrl      = "http:/test.com/"
		goodOrigin  = "https://you.com"
		badOrigin   = "https://me.com"
		emptyOrigin = ""
	)
	os.Setenv("ALLOWED_ORIGIN_REGEXP", "^https?:\\/\\/.*you\\.com")
	tests := []struct {
		title  string
		method string
		url    string
		origin string
		err    error
		status int
	}{
		{
			"Test missing origin header",
			"GET",
			goodUrl,
			emptyOrigin,
			errMissingOrigin,
			http.StatusBadRequest,
		},
		{
			"Test missing query string",
			"GET",
			badUrl,
			goodOrigin,
			errMissingQuery,
			http.StatusBadRequest,
		},
		{
			"Test wrong VERB",
			"POST",
			goodUrl,
			goodOrigin,
			errUnsupportedMethod,
			http.StatusMethodNotAllowed,
		},
		{
			"Test origin mismatch",
			"GET",
			goodUrl,
			badOrigin,
			errOriginMismatch,
			http.StatusBadRequest,
		},
		{
			"Test valid case",
			"GET",
			goodUrl,
			goodOrigin,
			nil,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, tt.url, nil)
		if tt.origin != "" {
			r.Header.Set("Origin", tt.origin)
		}
		wantErr, wantStatus := tt.err, tt.status
		if gotStatus, gotErr := validateRequest(r); gotErr != wantErr || gotStatus != wantStatus {
			t.Errorf("%s:  got err %v, want err %v\ngot status %d, want status %d", tt.title, gotErr, wantErr, gotStatus, wantStatus)
		}
	}
}

func TestWriteCorsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http:/test.com/?https://what.the.what", nil)
	rsp := http.Response{"200 OK", 200, "HTTP/1.1", 1, 1, http.Header{}, nil, -1, nil, true, true, nil, nil, nil}
	rsp.Header.Set("a", "b")
	r.Header.Set("Origin", "http://you")
	expectedHeaders := http.Header{}
	expectedHeaders.Set("A", "b")
	expectedHeaders.Set("Access-Control-Allow-Origin", "http://you")
	expectedHeaders.Set("Access-Control-Allow-Methods", "GET")
	expectedHeaders.Set("Access-Control-Max-Age", "86400")
	writeCorsHeaders(w, r, rsp.Header)
	eq := reflect.DeepEqual(w.Header(), expectedHeaders)
	if !eq {
		t.Errorf("expected headers %q != actual headers %q", expectedHeaders, w.Header())
	}
}
