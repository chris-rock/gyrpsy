package cert

import (
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
	"time"

	"github.com/Sirupsen/logrus"
)

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
		logrus.Errorf("Unrecognized elliptic curve: %q", config.EcdsaCurve)
	}
	if err != nil {
		logrus.Errorf("failed to generate private key: %s", err)
	}

	var notBefore time.Time
	if len(config.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", config.ValidFrom)
		if err != nil {
			logrus.Errorf("Failed to parse creation date: %s\n", err)
			return nil, nil, err
		}
	}

	notAfter := notBefore.Add(config.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		logrus.Errorf("failed to generate serial number: %s", err)
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
		logrus.Errorf("Failed to create certificate: %s", err)
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
			logrus.Errorf("Unable to marshal ECDSA private key: %v", err)
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, nil
	}
}
