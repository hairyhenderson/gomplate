package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"time"
)

// Req -
type Req struct {
	Headers http.Header `json:"headers"`
}

var port int

func main() {
	flag.IntVar(&port, "p", 8080, "Port to listen to")
	flag.Parse()

	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		log.Fatal(err)
	}
	// defer l.Close()
	http.HandleFunc("/", rootHandler)

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

func quitHandler(l net.Listener) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		go func() {
			time.Sleep(500 * time.Millisecond)
			l.Close()
		}()
	}
}
