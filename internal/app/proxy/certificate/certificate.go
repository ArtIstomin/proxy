package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	mrand "math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	rsaBits    = 2048
	certFolder = "certs"
)

var mu sync.Mutex

// Generate creates self-signed certificate for provided hostname
func Generate(r *http.Request) *tls.Config {
	hostname := strings.Split(r.Host, ":")[0]
	certFile, keyFile := getCertPath(hostname)

	if !pathExists(certFile) {
		log.Println("creating cert for", hostname)
		createCerts(hostname)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	checkError(err)

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}

func createCerts(hostName string) {
	mu.Lock()
	defer mu.Unlock()

	caCertFile := "certs/myCA.cer"
	caKeyFile := "certs/myCA.key"

	certFile, keyFile := getCertPath(hostName)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(mrand.Int63n(time.Now().Unix())),
		Subject: pkix.Name{
			Country:            []string{"DE"},
			Organization:       []string{"PPP"},
			OrganizationalUnit: []string{"MMM"},
			CommonName:         hostName,
		},

		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().AddDate(10, 0, 0).UTC(),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames:     []string{hostName},
	}

	if ip := net.ParseIP(hostName); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, hostName)
	}

	var err error
	var rootCA tls.Certificate

	rootCA, err = tls.LoadX509KeyPair(caCertFile, caKeyFile)
	checkError(err)

	rootCA.Leaf, err = x509.ParseCertificate(rootCA.Certificate[0])
	checkError(err)

	var priv *rsa.PrivateKey

	priv, err = rsa.GenerateKey(rand.Reader, rsaBits)
	checkError(err)

	var derBytes []byte

	derBytes, err = x509.CreateCertificate(rand.Reader, template, rootCA.Leaf, &priv.PublicKey, rootCA.PrivateKey)
	checkError(err)

	if !pathExists(certFolder) {
		os.Mkdir(certFolder, 0777)
	}

	certOut, err := os.Create(certFile)
	checkError(err)

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	checkError(err)

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func getCertPath(hostname string) (string, string) {
	cert := path.Join(certFolder, hostname+".pem")
	key := path.Join(certFolder, hostname+".key")
	return cert, key
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
