package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fishy/httpsproxy"
)

const (
	blynkURL   = "https://blynk-cloud.com"
	selfURLEnv = "SELF_URL"
)

// Get it by the following command:
// openssl s_client -showcerts -connect blynk-cloud.com:443 </dev/null
const blynkCert = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`

// AppEngine log will auto add date and time, so there's no need to double log
// them in our own loggers.
var (
	infoLog  = log.New(os.Stderr, "I ", log.Lshortfile)
	warnLog  = log.New(os.Stderr, "W ", log.Lshortfile)
	errorLog = log.New(os.Stderr, "E ", log.Lshortfile)
)

func main() {

	targetURL, err := url.Parse(blynkURL)
	if err != nil {
		errorLog.Fatal(err)
	}

	certPool, failed, err := httpsproxy.NewCertPool(blynkCert)
	if err != nil {
		warnLog.Printf("Cannot get system cert pool: %v", err)
	}
	if len(failed) > 0 {
		warnLog.Printf("Failed to add cert(s) to pool: %v", failed)
	}

	selfURL, err := url.Parse(os.Getenv(selfURLEnv))
	if err != nil {
		warnLog.Printf("Cannot get parse self URL: %v", err)
	}

	mux := httpsproxy.Mux(
		httpsproxy.DefaultHTTPClient(
			certPool,
			30*time.Second,
			httpsproxy.NoRedirCheckRedirectFunc,
		),
		targetURL,
		selfURL,
		errorLog,
	)
	// AppEngine health check
	mux.HandleFunc(
		"/_ah/health",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "healthy")
		},
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		infoLog.Printf("Defaulting to port %s", port)
	}
	infoLog.Printf("Listening on port %s", port)
	infoLog.Fatal(http.ListenAndServe(":"+port, mux))
}
