package server

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/chris-rock/gyrpsy/cert"
)

func GetCertificates(hostname string) (ko []byte, co []byte) {

	certFilename := "./cert.pem"
	keyFilename := "./key.pem"

	// only generate key and cert
	// if ./cert.pem or ./key.pem do not exit
	_, errCert := os.Stat(certFilename)
	_, errKey := os.Stat(keyFilename)
	if os.IsNotExist(errCert) || os.IsNotExist(errKey) {
		config := &cert.CertConfig{
			Host:         hostname,
			Organization: "Acme Co",
			ValidFrom:    "",
			ValidFor:     365 * 24 * time.Hour,
			IsCA:         false,
			RsaBits:      2048,
			EcdsaCurve:   "",
		}

		ko, co = GenerateCertificates(config, certFilename, keyFilename)
	} else {
		// load key from disk
		var err error
		ko, err = ioutil.ReadFile(keyFilename)
		if err != nil {
			log.Fatalf("failed to open %s for reading: %s", keyFilename, err)
		}
		co, err = ioutil.ReadFile(certFilename)
		if err != nil {
			log.Fatalf("failed to open %s for reading: %s", certFilename, err)
		}
	}

	return ko, co
}

func GenerateCertificates(config *cert.CertConfig, certFilename string, keyFilename string) ([]byte, []byte) {
	co, ko, err := cert.Generate(config)
	if err != nil {
		logrus.Error(err)
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
