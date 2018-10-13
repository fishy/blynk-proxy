package httpsproxy

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

const (
	targetScheme = "https"
)

var requestHeadersToCopy = []string{
	"Content-Type",
	"User-Agent",
}

var client *http.Client

// ProxyMux creates an http serve mux to do the proxy job.
func ProxyMux(
	targetHost string,
	certPool *x509.CertPool,
	selfURL *url.URL,
	timeout time.Duration,
) *http.ServeMux {
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if selfURL != nil {
				// Don't follow any redirects, rewrite Location header later.
				return http.ErrUseLastResponse
			}
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		Timeout: timeout,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerWrapper(client, targetHost, selfURL))

	return mux
}

func handlerWrapper(
	client *http.Client,
	targetHost string,
	selfURL *url.URL,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newURL := &url.URL{
			Scheme: targetScheme,
			Host:   targetHost,
			// In incoming r.URL only these 2 fields are set:
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		req, err := http.NewRequest(r.Method, newURL.String(), r.Body)
		if CheckError(w, err) {
			return
		}
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)
		CopyRequestHeaders(r, req, requestHeadersToCopy)

		resp, err := client.Do(req)
		if CheckError(w, err) {
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if CheckError(w, err) {
			return
		}

		header := w.Header()
		for key, values := range resp.Header {
			canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
			for _, value := range values {
				if canonicalKey == "Location" {
					value = RewriteURL(value, targetHost, selfURL)
				}
				header.Add(canonicalKey, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

// CheckError checks error. If error is non-nil, it writes HTTP status code 502
// (bad gateway) and the error message to the response and returns true.
func CheckError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	log.Print(err)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(err.Error()))
	return true
}

// CopyRequestHeaders copies specified headers from one http.Request to another.
func CopyRequestHeaders(from, to *http.Request, headers []string) {
	for _, header := range headers {
		value := from.Header.Get(header)
		if value != "" {
			to.Header.Set(header, value)
		}
	}
}

// RewriteURL rewrites all targetHost URLs to us.
func RewriteURL(origURL, targetHost string, selfURL *url.URL) string {
	if selfURL == nil {
		return origURL
	}

	u, err := url.Parse(origURL)
	if err != nil {
		log.Print(err)
		return origURL
	}
	if u.Host == targetHost {
		u.Scheme = selfURL.Scheme
		u.Host = selfURL.Host
		return u.String()
	}
	return origURL
}
