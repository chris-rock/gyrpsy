// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style

// Generate a self-signed X.509 certificate for a TLS server. Outputs to
// 'cert.pem' and 'key.pem' and will overwrite existing files.

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/chris-rock/gyrpsy/cert"
)

var (
	host       = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	validFrom  = flag.String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	validFor   = flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	isCA       = flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
	rsaBits    = flag.Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	ecdsaCurve = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521")
)

func main() {
	flag.Parse()

	config := &cert.CertConfig{
		Host:         *host,
		Organization: "Acme Co",
		ValidFrom:    *validFrom,
		ValidFor:     *validFor,
		IsCA:         *isCA,
		RsaBits:      *rsaBits,
		EcdsaCurve:   *ecdsaCurve,
	}

	co, ko, err := cert.Generate(config)

	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	log.Printf("Cert:\n%s\n%s", string(co.Bytes()), string(ko.Bytes()))

	certOut, err := os.Create("cert.pem")
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}

	// write buffer to file
	certOut.Write(co.Bytes())

	certOut.Close()
	log.Print("written cert.pem\n")

	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print("failed to open key.pem for writing:", err)
		return
	}

	// write buffer to file
	keyOut.Write(ko.Bytes())
	keyOut.Close()
	log.Print("written key.pem\n")
}
