package corsproxy
import(
	"testing"
	"net/http"
	"net/http/httptest"
	"reflect"
	"os"
)

func TestValidateRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "http:/test.com/?https://what.the.what", nil)
	os.Setenv("ALLOWED_ORIGIN_REGEXP", "^https?:\\/\\/.*you\\.com")
	myErr := validateRequest(r)
	if myErr == nil {
		t.Errorf("Error == nil, expected error")
	}
	if myErr.Error() != "Missing origin header" {
		t.Errorf("Error == %s, expected error %s", myErr.Error(), "Missing origin header")
	}
	r = httptest.NewRequest("GET", "http:/test.com/", nil)
	r.Header.Set("Origin", "https://you.com")
	myErr = validateRequest(r)
	if myErr == nil {
		t.Errorf("Error == nil, expected error")
	}
	if myErr.Error() != "Missing request query" {
		t.Errorf("Error == %s, expected error %s", myErr.Error(), "Missing origin header")
	}
	r = httptest.NewRequest("POST", "http:/test.com/?https://what.the.what", nil)
	r.Header.Set("Origin", "https://you.com")
	myErr = validateRequest(r)
	if myErr == nil {
		t.Errorf("Error == nil, expected error")
	}
	if myErr.Error() != "Cross domain request only supports GET" {
		t.Errorf("Error == %s, expected error %s", myErr.Error(), "Missing origin header")
	}
	r = httptest.NewRequest("GET", "http:/test.com/?https://what.the.what", nil)
	r.Header.Set("Origin", "https://me.com")
	myErr = validateRequest(r)
	if myErr == nil {
		t.Errorf("Error == nil, expected error")
	}
	if myErr.Error() != "origin mismatch" {
		t.Errorf("Error == %s, expected error %s", myErr.Error(), "Missing origin header")
	}
	r = httptest.NewRequest("GET", "http:/test.com/?https://what.the.what", nil)
	r.Header.Set("Origin", "https://you.com")
	myErr = validateRequest(r)
	if myErr != nil {
		t.Errorf("Error: %s, expected nil", myErr.Error())
	}
}

func TestWriteCorsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http:/test.com/?https://what.the.what", nil)
	rsp := http.Response{ "200 OK", 200, "HTTP/1.1", 1, 1, http.Header{}, nil, -1, nil, true, true, nil, nil, nil, }
	rsp.Header.Set("a", "b")
	r.Header.Set("Origin", "you")
	expectedHeaders := http.Header{}
	expectedHeaders.Set("A", "b")
	expectedHeaders.Set("Access-Control-Allow-Origin", "you")
	expectedHeaders.Set("Access-Control-Allow-Methods", "GET")
	expectedHeaders.Set("Access-Control-Max-Age", "86400")
	writeCorsHeaders(w, r, &rsp)
	eq := reflect.DeepEqual(w.Header(), expectedHeaders)
	if !eq {
		t.Errorf("expected headers %q != actual headers %q", expectedHeaders, w.Header())
	}
}
