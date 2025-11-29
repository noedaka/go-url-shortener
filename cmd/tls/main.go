package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"url-shortener"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}

	exeDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	certsDir := filepath.Join(exeDir, "certs")

	certPath := filepath.Join(certsDir, "cert.pem")
	keyPath := filepath.Join(certsDir, "private.pem")

	if err = os.WriteFile(certPath, certPEM.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}

	if err = os.WriteFile(keyPath, privateKeyPEM.Bytes(), 0600); err != nil {
		log.Fatal(err)
	}
}
