package httpsproxy_test

import (
	"crypto/x509"
	"fmt"
	"strings"
	"testing"

	"github.com/fishy/blynk-proxy/lib/httpsproxy"
)

var validCert = `-----BEGIN CERTIFICATE-----
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

var invalidCert = "asdf"

func TestCertPool(t *testing.T) {
	basePool, baseErr := getBaselinePool(t)

	t.Run(
		"no new certs",
		func(t *testing.T) {
			newPool, failedCerts, newErr := httpsproxy.NewCertPool()
			if len(failedCerts) != 0 {
				t.Errorf("failedCerts expected to be empty, got: %v", failedCerts)
			}
			if newErr != baseErr {
				t.Errorf("Expected error %v, got %v", baseErr, newErr)
			}
			comparePools(t, basePool, newPool, 0)
		},
	)

	t.Run(
		"valid cert",
		func(t *testing.T) {
			newPool, failedCerts, newErr := httpsproxy.NewCertPool(validCert)
			if len(failedCerts) != 0 {
				t.Errorf("failedCerts expected to be empty, got: %v", failedCerts)
			}
			if newErr != baseErr {
				t.Errorf("Expected error %v, got %v", baseErr, newErr)
			}
			comparePools(t, basePool, newPool, 1)
		},
	)

	t.Run(
		"invalid cert",
		func(t *testing.T) {
			newPool, failedCerts, newErr := httpsproxy.NewCertPool(invalidCert)
			if len(failedCerts) != 1 || failedCerts[0] != invalidCert {
				t.Errorf("failedCerts expected 1, got: %v", failedCerts)
			}
			if newErr != baseErr {
				t.Errorf("Expected error %v, got %v", baseErr, newErr)
			}
			comparePools(t, basePool, newPool, 0)
		},
	)

	t.Run(
		"mixed certs",
		func(t *testing.T) {
			newPool, failedCerts, newErr := httpsproxy.NewCertPool(
				validCert,
				invalidCert,
			)
			if len(failedCerts) != 1 || failedCerts[0] != invalidCert {
				t.Errorf("failedCerts expected 1, got: %v", failedCerts)
			}
			if newErr != baseErr {
				t.Errorf("Expected error %v, got %v", baseErr, newErr)
			}
			comparePools(t, basePool, newPool, 1)
		},
	)
}

func getBaselinePool(t *testing.T) (*x509.CertPool, error) {
	t.Helper()

	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	return pool, err
}

func comparePools(t *testing.T, base, newPool *x509.CertPool, diff uint) {
	t.Helper()

	inBase, inNew := subjectDiff(base.Subjects(), newPool.Subjects())

	if len(inBase) > 0 {
		t.Errorf(
			"The following certs are not in new pool: %s",
			subjectsToString(inBase),
		)
	}

	if len(inNew) != int(diff) {
		t.Errorf(
			"Exepcted %d new certs, got %s",
			len(inNew),
			subjectsToString(inNew),
		)
	}

	t.Logf("New certs: %s", subjectsToString(inNew))
}

func subjectDiff(a, b [][]byte) (inA, inB [][]byte) {
	mapA := subjectsToMap(a)
	mapB := subjectsToMap(b)

	for _, sub := range a {
		if !mapB[string(sub)] {
			inA = append(inA, sub)
		}
	}

	for _, sub := range b {
		if !mapA[string(sub)] {
			inB = append(inB, sub)
		}
	}

	return
}

func subjectsToMap(subs [][]byte) map[string]bool {
	ret := make(map[string]bool)
	for _, sub := range subs {
		ret[string(sub)] = true
	}
	return ret
}

func subjectsToString(subs [][]byte) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d:[", len(subs)))
	for i, sub := range subs {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", sub))
	}
	builder.WriteString("]")
	return builder.String()
}
