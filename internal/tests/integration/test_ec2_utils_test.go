//go:build !windows
// +build !windows

package integration

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/fullsailor/pkcs7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
)

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

func instanceDocumentHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(instanceDocument))
	if err != nil {
		w.WriteHeader(500)
	}
}

func certificateGenerate() (priv *rsa.PrivateKey, derBytes []byte, err error) {
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
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	return priv, derBytes, err
}

func pkcsHandler(priv *rsa.PrivateKey, derBytes []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
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
		if err = signedData.AddSigner(cert, priv, pkcs7.SignerInfoConfig{}); err != nil {
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
		_, err = w.Write(encoded)
		if err != nil {
			w.WriteHeader(500)
		}
	}
}

func stsHandler(t *testing.T, accountID, user string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		form, _ := url.ParseQuery(string(body))

		// action must be GetCallerIdentity
		assert.Equal(t, "GetCallerIdentity", form.Get("Action"))

		w.Header().Set("Content-Type", "text/xml")
		_, err := w.Write([]byte(fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
        <GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
            <GetCallerIdentityResult>
                <Arn>arn:aws:iam::%[1]s:user/%[2]s</Arn>
                <UserId>AKIAI44QH8DHBEXAMPLE</UserId>
                <Account>%[1]s</Account>
            </GetCallerIdentityResult>
            <ResponseMetadata>
                <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
            </ResponseMetadata>
        </GetCallerIdentityResponse>`, accountID, user)))
		if err != nil {
			t.Errorf("failed to write response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		assert.NoError(t, err)
	})
}

func ec2Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	_, err := w.Write([]byte(`<DescribeInstancesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>8f7724cf-496f-496e-8fe3-example</requestId>
    <reservationSet>
        <item>
            <reservationId>r-1234567890abcdef0</reservationId>
            <ownerId>123456789012</ownerId>
            <groupSet/>
            <instancesSet>
                <item>
                    <instanceId>i-00000000000000000</instanceId>
                    <imageId>ami-00000000</imageId>
                    <instanceState>
                        <code>16</code>
                        <name>running</name>
                    </instanceState>
                    <privateDnsName>ip-192-168-1-88.eu-west-1.compute.internal</privateDnsName>
                    <dnsName>ec2-54-194-252-215.eu-west-1.compute.amazonaws.com</dnsName>
                    <reason/>
                    <keyName>my_keypair</keyName>
                    <amiLaunchIndex>0</amiLaunchIndex>
                    <productCodes/>
                    <instanceType>t2.micro</instanceType>
                    <launchTime>2015-12-22T10:44:05.000Z</launchTime>
                    <placement>
                        <availabilityZone>eu-west-1c</availabilityZone>
                        <groupName/>
                        <tenancy>default</tenancy>
                    </placement>
                    <monitoring>
                        <state>disabled</state>
                    </monitoring>
                    <subnetId>subnet-56f5f633</subnetId>
                    <vpcId>vpc-11112222</vpcId>
                    <privateIpAddress>192.168.1.88</privateIpAddress>
                    <ipAddress>54.194.252.215</ipAddress>
                    <sourceDestCheck>true</sourceDestCheck>
                    <groupSet>
                        <item>
                            <groupId>sg-e4076980</groupId>
                            <groupName>SecurityGroup1</groupName>
                        </item>
                    </groupSet>
                    <architecture>x86_64</architecture>
                    <rootDeviceType>ebs</rootDeviceType>
                    <rootDeviceName>/dev/xvda</rootDeviceName>
                    <blockDeviceMapping>
                        <item>
                            <deviceName>/dev/xvda</deviceName>
                            <ebs>
                                <volumeId>vol-1234567890abcdef0</volumeId>
                                <status>attached</status>
                                <attachTime>2015-12-22T10:44:09.000Z</attachTime>
                                <deleteOnTermination>true</deleteOnTermination>
                            </ebs>
                        </item>
                    </blockDeviceMapping>
                    <virtualizationType>hvm</virtualizationType>
                    <clientToken>xMcwG14507example</clientToken>
                    <tagSet>
                        <item>
                            <key>Name</key>
                            <value>Server_1</value>
                        </item>
                    </tagSet>
                    <hypervisor>xen</hypervisor>
                    <networkInterfaceSet>
                        <item>
                            <networkInterfaceId>eni-551ba033</networkInterfaceId>
                            <subnetId>subnet-56f5f633</subnetId>
                            <vpcId>vpc-11112222</vpcId>
                            <description>Primary network interface</description>
                            <ownerId>123456789012</ownerId>
                            <status>in-use</status>
                            <macAddress>02:dd:2c:5e:01:69</macAddress>
                            <privateIpAddress>192.168.1.88</privateIpAddress>
                            <privateDnsName>ip-192-168-1-88.eu-west-1.compute.internal</privateDnsName>
                            <sourceDestCheck>true</sourceDestCheck>
                            <groupSet>
                                <item>
                                    <groupId>sg-e4076980</groupId>
                                    <groupName>SecurityGroup1</groupName>
                                </item>
                            </groupSet>
                            <attachment>
                                <attachmentId>eni-attach-39697adc</attachmentId>
                                <deviceIndex>0</deviceIndex>
                                <status>attached</status>
                                <attachTime>2015-12-22T10:44:05.000Z</attachTime>
                                <deleteOnTermination>true</deleteOnTermination>
                            </attachment>
                            <association>
                                <publicIp>54.194.252.215</publicIp>
                                <publicDnsName>ec2-54-194-252-215.eu-west-1.compute.amazonaws.com</publicDnsName>
                                <ipOwnerId>amazon</ipOwnerId>
                            </association>
                            <privateIpAddressesSet>
                                <item>
                                    <privateIpAddress>192.168.1.88</privateIpAddress>
                                    <privateDnsName>ip-192-168-1-88.eu-west-1.compute.internal</privateDnsName>
                                    <primary>true</primary>
                                    <association>
                                    <publicIp>54.194.252.215</publicIp>
                                    <publicDnsName>ec2-54-194-252-215.eu-west-1.compute.amazonaws.com</publicDnsName>
                                    <ipOwnerId>amazon</ipOwnerId>
                                    </association>
                                </item>
                            </privateIpAddressesSet>
                            <ipv6AddressesSet>
                               <item>
                                   <ipv6Address>2001:db8:1234:1a2b::123</ipv6Address>
                               </item>
                           </ipv6AddressesSet>
                        </item>
                    </networkInterfaceSet>
                    <ebsOptimized>false</ebsOptimized>
                </item>
            </instancesSet>
        </item>
    </reservationSet>
</DescribeInstancesResponse>`))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func iamGetUserHandler(t *testing.T, accountID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))

		// action must be GetUser
		assert.Equal(t, "GetUser", form.Get("Action"))

		w.Header().Set("Content-Type", "text/xml")
		_, err := w.Write([]byte(fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
        <GetUserResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
            <GetUserResult>
                <User>
                    <Path>/</Path>
                    <UserName>%[1]s</UserName>
                    <UserId>m3o9qmhhl9dnjlh2fflg</UserId>
                    <Arn>arn:aws:iam::%[2]s:user/%[1]s</Arn>
                    <CreateDate>2024-07-21T17:21:27.259000Z</CreateDate>
                </User>
            </GetUserResult>
            <ResponseMetadata>
                <RequestId>3d0e2445-64ea-4bfb-9244-30d810773f9e</RequestId>
            </ResponseMetadata>
        </GetUserResponse>`, form.Get("UserName"), accountID)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		assert.NoError(t, err)
	})
}

func setupDatasourcesVaultAWSTest(t *testing.T, accountID, user string) (*fs.Dir, *vaultClient, *httptest.Server, []byte) {
	t.Helper()

	priv, der, _ := certificateGenerate()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	mux := http.NewServeMux()
	mux.HandleFunc("/latest/dynamic/instance-identity/pkcs7", pkcsHandler(priv, der))
	mux.HandleFunc("/latest/dynamic/instance-identity/document", instanceDocumentHandler)
	mux.HandleFunc("/latest/api/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		if r.Body != nil {
			var err error
			b, err = io.ReadAll(r.Body)
			if !assert.NoError(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
		}
		t.Logf("IMDS Token request: %s %s: %s", r.Method, r.URL, b)

		w.Write([]byte("testtoken"))
	}))
	mux.HandleFunc("/latest/meta-data/instance-id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("IMDS request: %s %s", r.Method, r.URL)
		w.Write([]byte("i-00000000"))
	}))
	mux.Handle("/sts/", stsHandler(t, accountID, user))
	mux.Handle("/iam/", iamGetUserHandler(t, accountID))
	mux.HandleFunc("/ec2/", ec2Handler)
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("unhandled request: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))

	// Vault sends requests to "/sts///" for some reason, and the ServeMux
	// responds by redirecting to "/sts/" which Vault rejects. So we need to
	// handle the extra slashes in a middleware first.
	stripSlashes := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for strings.HasSuffix(r.URL.Path, "//") {
			r.URL.Path = r.URL.Path[:len(r.URL.Path)-1]
		}
		mux.ServeHTTP(w, r)
	})

	srv := httptest.NewServer(stripSlashes)
	t.Cleanup(srv.Close)

	tmpDir, v := startVault(t)

	err := v.vc.Sys().PutPolicy("writepol", `path "*" {
  policy = "write"
}`)
	require.NoError(t, err)
	err = v.vc.Sys().PutPolicy("readpol", `path "*" {
  policy = "read"
}`)
	require.NoError(t, err)

	return tmpDir, v, srv, cert
}
