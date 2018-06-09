package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
)

const (
	blynkScheme = "https"
	blynkHost   = "blynk-cloud.com"
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
	"User-Agent",
	"Content-Type",
}

var client *http.Client

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

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
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handler(w http.ResponseWriter, r *http.Request) {
	newURL := &url.URL{
		Scheme:   blynkScheme,
		Host:     blynkHost,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}
	newBody, err := convertRequestBody(r.Body)
	if checkError(w, err) {
		return
	}
	req, err := http.NewRequest(r.Method, newURL.String(), newBody)
	if checkError(w, err) {
		return
	}
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	copyRequestHeaders(r, req, requestHeadersToCopy)

	resp, err := client.Do(req)
	if checkError(w, err) {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if checkError(w, err) {
		return
	}

	header := w.Header()
	for key, values := range resp.Header {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, value := range values {
			header.Add(canonicalKey, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func checkError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	log.Print(err)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(err.Error()))
	return true
}

func convertRequestBody(body io.ReadCloser) (io.ReadCloser, error) {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, nil
	}
	return ioutil.NopCloser(bytes.NewReader(buf)), nil
}

func copyRequestHeaders(r1, r2 *http.Request, headers []string) {
	for _, header := range headers {
		value := r1.Header.Get(header)
		if value != "" {
			r2.Header.Set(header, value)
		}
	}
}
