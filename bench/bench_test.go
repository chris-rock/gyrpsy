package bench

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/chris-rock/gyrpsy/api/pingpong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var res *pingpong.PongReply

func GetCertificates(keyFilename string, certFilename string) (ko []byte, co []byte) {
	// load key from disk
	var err error
	ko, err = ioutil.ReadFile(keyFilename)
	if err != nil {
		log.Fatalf("failed to open %s for reading: %s", keyFilename, err)
		panic("could not load tls key")
	}
	co, err = ioutil.ReadFile(certFilename)
	if err != nil {
		log.Fatalf("failed to open %s for reading: %s", certFilename, err)
		panic("could not load tls cert")
	}
	return ko, co
}

func BenchmarkGoGatewayGRPC(b *testing.B) {
	serverAddr := "localhost:5000"
	creds, err := credentials.NewClientTLSFromFile("../cert/cert_localhost_5000.pem", "")
	if err != nil {
		panic(err)
	}
	creds.OverrideServerName(serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		b.Fatalf("could not dial %v", err)
	}
	defer conn.Close()

	b.Run("pingpong proto ingestion", func(b *testing.B) {
		client := pingpong.NewPingPongClient(conn)
		for n := 0; n < b.N; n++ {
			output, err := client.Ping(context.Background(), &pingpong.PingRequest{Sender: "John"})
			if err != nil {
				b.Fatalf("could not call %v", err)
			}
			res = output
		}
	})
}

func BenchmarkGoGatewayRest(b *testing.B) {
	_, cert := GetCertificates("../cert/key_localhost_5000.pem", "../cert/cert_localhost_5000.pem")

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
	baseURL, _ := url.Parse("https://localhost:5000")
	rel := &url.URL{Path: "/pingpong/ping"}
	u := baseURL.ResolveReference(rel)

	b.Run("pingpong json ingestion ", func(b *testing.B) {
		var body = []byte(`{ "sender": "John"}`)
		req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
		if err != nil {
			b.Fatalf("Error %s", err)
		}
		req.Header.Set("Accept", "application/json")
		httpClient := &http.Client{Transport: transport}
		for n := 0; n < b.N; n++ {
			// execute request
			res, err := httpClient.Do(req)
			if err != nil {
				b.Fatalf("Error %s", err)
			}
			defer res.Body.Close()
			reqdata, _ := ioutil.ReadAll(res.Body)
			if res.StatusCode != 200 {
				b.Fatalf("could not send message %d\n%v\n%v", res.StatusCode, res.Status, string(reqdata))
			}
		}
	})
}

func BenchmarkDirectGrpc(b *testing.B) {
	serverAddr := "localhost:5001"
	creds, err := credentials.NewClientTLSFromFile("../cert/cert_localhost_5001.pem", "")
	if err != nil {
		panic(err)
	}
	creds.OverrideServerName(serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		b.Fatalf("could not dial %v", err)
	}
	defer conn.Close()

	b.Run("pingpong proto ingestion", func(b *testing.B) {
		client := pingpong.NewPingPongClient(conn)
		for n := 0; n < b.N; n++ {
			output, err := client.Ping(context.Background(), &pingpong.PingRequest{Sender: "John"})
			if err != nil {
				b.Fatalf("could not call %v", err)
			}
			res = output
		}
	})
}

func BenchmarkDirectRest(b *testing.B) {
	_, cert := GetCertificates("../cert/key_localhost_5002.pem", "../cert/cert_localhost_5002.pem")

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
	baseURL, _ := url.Parse("https://localhost:5002")
	rel := &url.URL{Path: "/ping"}
	u := baseURL.ResolveReference(rel)

	b.Run("pingpong json ingestion ", func(b *testing.B) {
		var body = []byte(`{ "sender": "John"}`)
		req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
		if err != nil {
			b.Fatalf("Error %s", err)
		}
		req.Header.Set("Accept", "application/json")
		httpClient := &http.Client{Transport: transport}

		for n := 0; n < b.N; n++ {
			// execute request
			res, err := httpClient.Do(req)
			if err != nil {
				b.Fatalf("could not do http call %v", err)
			}
			defer res.Body.Close()
			reqdata, _ := ioutil.ReadAll(res.Body)
			if res.StatusCode != 200 {
				b.Fatalf("could not send message %d\n%v\n%v", res.StatusCode, res.Status, string(reqdata))
			}
		}
	})
}

func BenchmarkDirectRestHTTP(b *testing.B) {
	// construct client message
	baseURL, _ := url.Parse("http://localhost:5003")
	rel := &url.URL{Path: "/ping"}
	u := baseURL.ResolveReference(rel)

	b.Run("pingpong json ingestion ", func(b *testing.B) {
		var body = []byte(`{ "sender": "John"}`)
		req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
		if err != nil {
			b.Fatalf("Error %s", err)
		}
		req.Header.Set("Accept", "application/json")
		httpClient := &http.Client{}

		for n := 0; n < b.N; n++ {
			// execute request
			res, err := httpClient.Do(req)
			if err != nil {
				b.Fatalf("could not do http call %v", err)
			}
			defer res.Body.Close()
			reqdata, _ := ioutil.ReadAll(res.Body)
			if res.StatusCode != 200 {
				b.Fatalf("could not send message %d\n%v\n%v", res.StatusCode, res.Status, string(reqdata))
			}
		}
	})
}

func BenchmarkNginxRestHTTPS(b *testing.B) {
	_, cert := GetCertificates("../cert/key_localhost_8443.pem", "../cert/cert_localhost_8443.pem")

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
	baseURL, _ := url.Parse("https://localhost:8443")
	rel := &url.URL{Path: "/ping"}
	u := baseURL.ResolveReference(rel)

	b.Run("pingpong json ingestion ", func(b *testing.B) {
		var body = []byte(`{ "sender": "John"}`)
		req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
		if err != nil {
			b.Fatalf("Error %s", err)
		}
		req.Header.Set("Accept", "application/json")
		httpClient := &http.Client{Transport: transport}

		for n := 0; n < b.N; n++ {
			// execute request
			res, err := httpClient.Do(req)
			if err != nil {
				b.Fatalf("could not do http call %v", err)
			}
			defer res.Body.Close()
			reqdata, _ := ioutil.ReadAll(res.Body)
			if res.StatusCode != 200 {
				b.Fatalf("could not send message %d\n%v\n%v", res.StatusCode, res.Status, string(reqdata))
			}
		}
	})
}

func BenchmarkNginxTLSGrpc(b *testing.B) {
	serverAddr := "localhost:8443"
	creds, err := credentials.NewClientTLSFromFile("../cert/cert_localhost_8443.pem", "")
	if err != nil {
		panic(err)
	}
	creds.OverrideServerName(serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		b.Fatalf("could not dial %v", err)
	}
	defer conn.Close()

	b.Run("pingpong proto ingestion", func(b *testing.B) {
		client := pingpong.NewPingPongClient(conn)
		for n := 0; n < b.N; n++ {
			output, err := client.Ping(context.Background(), &pingpong.PingRequest{Sender: "John"})
			if err != nil {
				b.Fatalf("could not call %v", err)
			}
			res = output
		}
	})
}

func BenchmarkNginxGrpc(b *testing.B) {
	serverAddr := "localhost:8444"
	creds, err := credentials.NewClientTLSFromFile("../cert/cert_localhost_8444.pem", "")
	if err != nil {
		panic(err)
	}
	creds.OverrideServerName(serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		b.Fatalf("could not dial %v", err)
	}
	defer conn.Close()

	b.Run("pingpong proto ingestion", func(b *testing.B) {
		client := pingpong.NewPingPongClient(conn)

		for n := 0; n < b.N; n++ {
			output, err := client.Ping(context.Background(), &pingpong.PingRequest{Sender: "John"})
			if err != nil {
				b.Fatalf("could not call %v", err)
			}
			res = output
		}
	})
}
