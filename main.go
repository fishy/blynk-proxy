package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"time"
)

const (
	blynkScheme = "https"
	blynkHost   = "blynk-cloud.com"
)

const (
	selfScheme = "https"
	selfHost   = "blynk-proxy.appspot.com"
)

var blynkCert = []byte(`-----BEGIN CERTIFICATE-----
MIID5TCCAs2gAwIBAgIJAIHSnb+cv4ECMA0GCSqGSIb3DQEBCwUAMIGIMQswCQYD
VQQGEwJVQTENMAsGA1UECAwES3lpdjENMAsGA1UEBwwES3lpdjELMAkGA1UECgwC
SVQxEzARBgNVBAsMCkJseW5rIEluYy4xGDAWBgNVBAMMD2JseW5rLWNsb3VkLmNv
bTEfMB0GCSqGSIb3DQEJARYQZG1pdHJpeUBibHluay5jYzAeFw0xNjAzMTcxMTU4
MDdaFw0yMTAzMTYxMTU4MDdaMIGIMQswCQYDVQQGEwJVQTENMAsGA1UECAwES3lp
djENMAsGA1UEBwwES3lpdjELMAkGA1UECgwCSVQxEzARBgNVBAsMCkJseW5rIElu
Yy4xGDAWBgNVBAMMD2JseW5rLWNsb3VkLmNvbTEfMB0GCSqGSIb3DQEJARYQZG1p
dHJpeUBibHluay5jYzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALso
bhbXQuNlzYBFa9h9pd69n43yrGTL4Ba6k5Q1zDwY9HQbMdfC5ZfnCkqT7Zf+R5MO
RW0Q9nLsFNLJkwKnluRCYGyUES8NAmDLQBbZoVc8mv9K3mIgAQvGyY2LmKak5GSI
V0PC3x+iN03xU2774+Zi7DaQd7vTl/9RGk8McyHe/s5Ikbe14bzWcY9ZV4PKgCck
p1chbmLhSfGbT3v64sL8ZbIppQk57/JgsZMrVpjExvxQPZuJfWbtoypPfpYO+O8l
1szaMlTEPIZVMoYi9uE+DnOlhzJFn6Ac4FMrDzJXzMmCweSX3IxguvXALeKhUHQJ
+VP3G6Q3pkZRVKz+5XsCAwEAAaNQME4wHQYDVR0OBBYEFJtqtI62Io66cZgiTR5L
A5Tl5m+xMB8GA1UdIwQYMBaAFJtqtI62Io66cZgiTR5LA5Tl5m+xMAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQELBQADggEBAKphjtEOGs7oC3S87+AUgIw4gFNOuv+L
C98/l47OD6WtsqJKvCZ1lmKxY5aIro9FBPk8ktCOsbwEjE+nyr5wul+6CLFr+rnv
7OHYGwLpjoz+rZgYJiQ61E1m0AZ4y9Fyd+D90HW6247vrBXyEiUXOhN/oDDVfDQA
eqmNBx1OqWel81D3tA7zPMA7vUItyWcFIXNjOCP+POy7TMxZuhuPMh5bVu+/cthl
/Q9u/Z2lKl4CWV0Ivt2BtlN6iefva0e2AP/As+gfwjxrb0t11zSILLNJ+nxRIwg+
k4MGb1zihKbIXUzsjslONK4FY5rlQUSwKJgEAVF0ClxB4g6dECm0ckc=
-----END CERTIFICATE-----`)

var requestHeadersToCopy = []string{
	"Content-Type",
	"User-Agent",
}

var client *http.Client

func main() {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("Cannot get system cert pool: %v", err)
		certPool = x509.NewCertPool()
	}
	certPool.AppendCertsFromPEM(blynkCert)

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow any redirects
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		Timeout: 30 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handler(w http.ResponseWriter, r *http.Request) {
	newURL := &url.URL{
		Scheme: blynkScheme,
		Host:   blynkHost,
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
				value = RewriteURL(value)
			}
			header.Add(canonicalKey, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
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

// RewriteURL rewrites all blynk-cloud.com URLs to us.
func RewriteURL(origURL string) string {
	u, err := url.Parse(origURL)
	if err != nil {
		log.Print(err)
		return origURL
	}
	if u.Host == blynkHost {
		u.Scheme = selfScheme
		u.Host = selfHost
		return u.String()
	}
	return origURL
}
