package corsproxy

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

var (
	errOriginMismatch    = errors.New("Origin mismatch")
	errMissingOrigin     = errors.New("Missing origin header")
	errMissingQuery      = errors.New("Missing request query")
	errUnsupportedMethod = errors.New("Method not allowed")
	errParsingRegex      = errors.New("Error parsing origin")
)

type myError struct {
	Code    int
	Message string
}

func (e myError) Error() string {
	return e.Message
}

type myResp struct {
	Code   int
	Body   []byte
	Header http.Header
}

func validateRequest(r *http.Request) (int, error) {
	switch {
	case r.Header.Get("Origin") == "":
		return http.StatusBadRequest, errMissingOrigin
	case r.URL.RawQuery == "":
		return http.StatusBadRequest, errMissingQuery
	case r.Method != "GET":
		return http.StatusMethodNotAllowed, errUnsupportedMethod
	}
	allowedOriginRe := os.Getenv("ALLOWED_ORIGIN_REGEXP")
	matched, err := regexp.MatchString(allowedOriginRe, r.Header.Get("Origin"))
	if err != nil {
		return http.StatusInternalServerError, errParsingRegex
	}
	if matched == false {
		return http.StatusBadRequest, errOriginMismatch
	}
	return 200, nil
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request, resp *http.Response) {
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Max-Age", "86400")
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}
	return
}

func fetchResp(ctx context.Context, url string) (*myResp, int, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		status := http.StatusInternalServerError
		if resp != nil && resp.StatusCode >= 100 {
			status = resp.StatusCode
		}
		return nil, status, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &myResp{resp.StatusCode, body, resp.Header}, resp.StatusCode, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	status, myerr := validateRequest(r)
	if myerr != nil {
		http.Error(w, myerr.Error(), status)
		return
	}
	log.Printf("URL: %s", r.URL.RawQuery)
	ctx := appengine.NewContext(r)
	resp, status, myerr := fetchResp(ctx, r.URL.RawQuery)
	if myerr != nil {
		http.Error(w, myerr.Error(), status)
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Max-Age", "86400")
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}
	w.WriteHeader(resp.Code)
	w.Write(resp.Body)
}

func init() {
	http.HandleFunc("/", handler)
}
