package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
)

// Req -
type Req struct {
	Headers http.Header `json:"headers"`
}

var port string

func main() {
	flag.StringVar(&port, "p", "8082", "Port to listen to")
	flag.Parse()

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	// defer l.Close()
	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/sts/", stsHandler)
	http.HandleFunc("/ec2/", ec2Handler)
	http.HandleFunc("/quit", quitHandler(l))

	http.Serve(l, nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	req := Req{r.Header}
	b, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func stsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(`<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
   <Arn>arn:aws:iam::1:user/Test</Arn>
    <UserId>AKIAI44QH8DHBEXAMPLE</UserId>
    <Account>1</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`))
}

func ec2Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(`<DescribeInstancesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
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
}

func quitHandler(l net.Listener) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Close()
		w.WriteHeader(http.StatusNoContent)
	}
}
