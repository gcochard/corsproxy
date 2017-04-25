package corsproxy
import (
        "fmt"
        "log"
        "net/http"
        "io/ioutil"
        "appengine"
        "appengine/urlfetch"
	"regexp"
	"os"
)

type myError struct {
	Code int
	Message string
}

func (e myError) Error() string {
	return e.Message
}

type myResp struct {
	Code int
	Body []byte
	Header http.Header
}

func ValidateRequest(r *http.Request) *myError {
	if r.Header.Get("Origin") == "" {
		return &myError{http.StatusBadRequest, "Missing origin header"}
	} else if r.URL.RawQuery == "" {
		return &myError{http.StatusBadRequest, "Missing request query"}
        } else if r.Method != "GET" {
		return &myError{http.StatusBadRequest, "Cross domain request only supports GET"}
	}
	allowedOriginRe := os.Getenv("ALLOWED_ORIGIN_REGEXP")
	matched, err := regexp.MatchString(allowedOriginRe, r.Header.Get("Origin"))
	if err != nil || matched == false {
		return &myError{http.StatusBadRequest, "origin mismatch"}
	}
	return nil
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

func fetchResp (ctx appengine.Context, url string) (*myResp, *myError) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
        if err != nil {
		status := http.StatusInternalServerError
                if resp != nil && resp.StatusCode >= 100 {
                        status = resp.StatusCode
                }
		return nil, &myError{status, fmt.Sprintf("Error: %s", err)}
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
		return nil, &myError{http.StatusInternalServerError, fmt.Sprintf("Error: %s", err)}
        }
	return &myResp{resp.StatusCode, body, resp.Header}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	myerr := ValidateRequest(r)
	if myerr != nil {
		http.Error(w, myerr.Message, myerr.Code)
		return
	}
        log.Printf("URL: %s", r.URL.RawQuery)
        ctx := appengine.NewContext(r)
        //client := urlfetch.Client(ctx)
        //resp, err := client.Get(r.URL.RawQuery)
	resp, myerr := fetchResp(ctx, r.URL.RawQuery)
	if myerr != nil {
		http.Error(w, myerr.Message, myerr.Code)
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
