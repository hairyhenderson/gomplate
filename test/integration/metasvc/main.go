package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"net/http"

	"github.com/fullsailor/pkcs7"
)

var port string
var priv *rsa.PrivateKey
var derBytes []byte

const instanceDocument = `{
    "devpayProductCodes" : null,
    "availabilityZone" : "xx-test-1b",
    "privateIp" : "10.1.2.3",
    "version" : "2010-08-31",
    "instanceId" : "i-00000000000000000",
    "billingProducts" : null,
    "instanceType" : "t2.micro",
    "accountId" : "1",
    "imageId" : "ami-00000000",
    "pendingTime" : "2000-00-01T0:00:00Z",
    "architecture" : "x86_64",
    "kernelId" : null,
    "ramdiskId" : null,
    "region" : "xx-test-1"
}`

func main() {
	flag.StringVar(&port, "p", "8081", "Port to listen to")
	flag.Parse()

	certificateGenerate()

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	// defer l.Close()
	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/latest/dynamic/instance-identity/pkcs7", pkcsHandler)
	http.HandleFunc("/latest/dynamic/instance-identity/document", documentHandler)
	http.HandleFunc("/certificate", certificateHandler)

	http.HandleFunc("/quit", quitHandler(l))

	http.Serve(l, nil)
}

func certificateGenerate() {
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func pkcsHandler(w http.ResponseWriter, r *http.Request) {
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		log.Fatalf("Cannot decode certificate: %s", err)
	}

	// Initialize a SignedData struct with content to be signed
	signedData, err := pkcs7.NewSignedData([]byte(instanceDocument))
	if err != nil {
		log.Fatalf("Cannot initialize signed data: %s", err)
	}

	// Add the signing cert and private key
	if err := signedData.AddSigner(cert, priv, pkcs7.SignerInfoConfig{}); err != nil {
		log.Fatalf("Cannot add signer: %s", err)
	}

	// Finish() to obtain the signature bytes
	detachedSignature, err := signedData.Finish()
	if err != nil {
		log.Fatalf("Cannot finish signing data: %s", err)
	}

	encoded := pem.EncodeToMemory(&pem.Block{Type: "PKCS7", Bytes: detachedSignature})

	encoded = bytes.TrimPrefix(encoded, []byte("-----BEGIN PKCS7-----\n"))
	encoded = bytes.TrimSuffix(encoded, []byte("\n-----END PKCS7-----\n"))

	w.Header().Set("Content-Type", "text/plain")
	w.Write(encoded)
}

func documentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(instanceDocument))
}

func certificateHandler(w http.ResponseWriter, r *http.Request) {
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	w.Header().Set("Content-Type", "text/plain")
	w.Write(encoded)
}

func quitHandler(l net.Listener) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Close()
		w.WriteHeader(http.StatusNoContent)
	}
}
