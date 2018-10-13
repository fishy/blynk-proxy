package httpsproxy

import (
	"crypto/x509"
)

// NewCertPool creates a new cert pool.
//
// It tries to get the system cert pool first, then append new pemCerts into the
// pool.
//
// Any new certs failed to append to the pool will be returned via failedCerts.
//
// If for any reason it's unable to get the system cert pool, the error will be
// returned by sysCertPoolErr and the returned certPool will only have
// successfully added new certs.
func NewCertPool(pemCerts ...string) (
	certPool *x509.CertPool,
	failedCerts []string,
	sysCertPoolErr error,
) {
	certPool, sysCertPoolErr = x509.SystemCertPool()
	if sysCertPoolErr != nil {
		certPool = x509.NewCertPool()
	}
	for _, cert := range pemCerts {
		if !certPool.AppendCertsFromPEM([]byte(cert)) {
			failedCerts = append(failedCerts, cert)
		}
	}
	return
}
