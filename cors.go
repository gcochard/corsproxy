package cors
import (
        "fmt"
        "log"
        "net/http"
        "io/ioutil"
        "google.golang.org/appengine"
        "google.golang.org/appengine/urlfetch"
	"regexp"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") == "" {
		http.Error(w, "Missing origin header", http.StatusBadRequest)
		return
	}
        if r.URL.RawQuery == "" {
		http.Error(w, "Missing request query", http.StatusBadRequest)
                return
        }
	if r.Method != "GET" {
		http.Error(w, "Cross domain request only supports GET", http.StatusBadRequest)
		return
	}
	allowedOriginRe := os.Getenv("ALLOWED_ORIGIN_REGEXP")
	matched, err := regexp.MatchString(allowedOriginRe, r.Header.Get("Origin"))
	if err != nil || matched == false {
		http.Error(w, "origin mismatch", http.StatusBadRequest)
		return
	}
        w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
        w.Header().Add("Access-Control-Allow-Methods", "GET")
        w.Header().Add("Access-Control-Max-Age", "86400")
        log.Printf("URL: %s", r.URL.RawQuery)
        ctx := appengine.NewContext(r)
        client := urlfetch.Client(ctx)
        resp, err := client.Get(r.URL.RawQuery)
        if err != nil {
		status := http.StatusInternalServerError
                if resp != nil && resp.StatusCode >= 100 {
                        status = resp.StatusCode
                }
		http.Error(w, fmt.Sprintf("Error: %s", err), status)
                return
        }
        for k, v := range resp.Header {
                for _, s := range v {
                        w.Header().Add(k, s)
                }
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
                return
        }
        w.WriteHeader(resp.StatusCode)
        w.Write(body)
}

func init() {
        http.HandleFunc("/", handler)
}
