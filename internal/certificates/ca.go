package certificates

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
)

var (
	localhostname, _ = os.Hostname()
)

const (
	keyFile  = "ca-key.pem"
	certFile = "ca-cert.crt"
)

// LoadCA loads the ca from "HOME/dir"
func LoadCA() (cert tls.Certificate, err error) {
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if os.IsNotExist(err) {
		cert, err = genCA()
	}
	if err == nil {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	return
}

func genCA() (cert tls.Certificate, err error) {
	certPEM, keyPEM, err := GenerateCA(localhostname)
	if err != nil {
		return
	}
	cert, _ = tls.X509KeyPair(certPEM, keyPEM)
	err = ioutil.WriteFile(certFile, certPEM, 0644)
	if err == nil {
		err = ioutil.WriteFile(keyFile, keyPEM, 0644)
	}
	return cert, err
}
