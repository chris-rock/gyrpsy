package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"math/big"
	"net"

	"strings"
)

func CreateCertificates(hostname string, port string) {
	certFilename := "cert_" + hostname + "_" + port + ".pem"
	keyFilename := "key_" + hostname + "_" + port + ".pem"

	_, errCert := os.Stat(certFilename)
	_, errKey := os.Stat(keyFilename)

	if os.IsNotExist(errCert) || os.IsNotExist(errKey) {
		GenerateCertificates(hostname+":"+port, certFilename, keyFilename)
	}
}

func GenerateCertificates(hostname string, certFilename string, keyFilename string) ([]byte, []byte) {

	config := &CertConfig{
		Host:         hostname,
		Organization: "Acme Co",
		ValidFrom:    "",
		ValidFor:     365 * 24 * time.Hour,
		IsCA:         false,
		RsaBits:      2048,
		EcdsaCurve:   "",
	}

	fmt.Printf("%v", config)

	co, ko, err := Generate(config)
	if err != nil {
		panic(err)
	}

	certOut, err := os.Create(certFilename)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}

	// write buffer to file
	certOut.Write(co)

	certOut.Close()
	log.Printf("written %s\n", certFilename)

	keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("failed to open %s for writing:%s", keyFilename, err)
		return nil, nil
	}

	// write buffer to file
	keyOut.Write(ko)
	keyOut.Close()
	log.Printf("written %s\n", keyFilename)

	return ko, co
}

type CertConfig struct {
	Host         string
	Organization string
	ValidFrom    string
	ValidFor     time.Duration
	IsCA         bool
	RsaBits      int
	EcdsaCurve   string
}

func Generate(config *CertConfig) (co []byte, ko []byte, err error) {
	var priv interface{}
	switch config.EcdsaCurve {
	case "":
		priv, err = rsa.GenerateKey(rand.Reader, config.RsaBits)
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		fmt.Printf("Unrecognized elliptic curve: %q", config.EcdsaCurve)
	}
	if err != nil {
		fmt.Printf("failed to generate private key: %s", err)
	}

	var notBefore time.Time
	if len(config.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", config.ValidFrom)
		if err != nil {
			fmt.Printf("Failed to parse creation date: %s\n", err)
			return nil, nil, err
		}
	}

	notAfter := notBefore.Add(config.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("failed to generate serial number: %s", err)
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(config.Host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if config.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s", err)
		return nil, nil, err
	}

	var coBuf bytes.Buffer
	coWriter := bufio.NewWriter(&coBuf)
	pem.Encode(coWriter, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	coWriter.Flush()

	var koBuf bytes.Buffer
	koWriter := bufio.NewWriter(&koBuf)

	block, err := pemBlockForKey(priv)
	if err != nil {
		return nil, nil, err
	}

	pem.Encode(koWriter, block)
	koWriter.Flush()

	return coBuf.Bytes(), koBuf.Bytes(), nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Printf("Unable to marshal ECDSA private key: %v", err)
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, nil
	}
}

func main() {
	CreateCertificates("localhost", "5000")
	CreateCertificates("localhost", "5001")
	CreateCertificates("localhost", "5002")
	CreateCertificates("localhost", "8443")
	CreateCertificates("localhost", "8444")
}
