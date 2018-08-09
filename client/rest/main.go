package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// load cert from disk
	certFilename := "./cert/cert_localhost_5002.pem"
	cert, err := ioutil.ReadFile(certFilename)
	if err != nil {
		log.Fatalf("failed to open %s for reading: %s", certFilename, err)
	}

	// Setup HTTPS client
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert))
	if !ok {
		panic("failed to parse root certificate")
	}
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            roots,
	}
	transport := &http.Transport{TLSClientConfig: tlsConf}

	// construct client message
	baseURL, err := url.Parse("https://localhost:5002")
	rel := &url.URL{Path: "/ping"}
	u := baseURL.ResolveReference(rel)

	var body = []byte(`{ "sender": "John"}`)
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	req.Header.Set("Accept", "application/json")

	// execute request
	httpClient := &http.Client{Transport: transport}
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error %s", err)
	}

	// display response
	pong, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", pong)
}
