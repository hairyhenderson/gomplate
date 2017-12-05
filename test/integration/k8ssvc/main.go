package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

var port string
var dir string

func main() {
	flag.StringVar(&port, "p", "8083", "Port to listen to")
	flag.StringVar(&dir, "d", "/tmp", "certs directory")
	flag.Parse()

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/v1/namespaces/default/configmaps/test/", testConfigmapHandler)
	http.HandleFunc("/quit", quitHandler(l))
	log.Fatal(http.ServeTLS(l, nil, dir+"/server.crt", dir+"/server.key"))

}

func testConfigmapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{                                                                                
        "kind": "ConfigMap",                                                           
        "apiVersion": "v1",                                                            
        "metadata": {                                                                  
          "name": "test",                                                           
          "namespace": "default",                                                      
          "selfLink": "/api/v1/namespaces/default/configmaps/test",                 
          "uid": "fae4ccf0-d695-11e7-861c-080027ec6cd8",                               
          "resourceVersion": "134",                                                    
          "creationTimestamp": "2017-12-01T12:48:50Z"                                  
        },                                                                             
        "data": {                                                                      
          "test.property.1": "foo",                                               
          "test.property.2": "bar",                                               
          "test.property.file": "property.1=value-1\nproperty.2=value-2\nproperty.3=value-3"
        }                                                                              
      }                                                                                
      `))
}

func quitHandler(l net.Listener) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Close()
		w.WriteHeader(http.StatusNoContent)
	}
}
