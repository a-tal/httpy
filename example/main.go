// example of using httpy

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/a-tal/httpy"
)

// lazily left here to compare the effect of calling httpy.Request
func goHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %+v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if len(body) < 1 {
		body = []byte("ok")
	}
	status := 200
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write: %+v", err)
	}
}

func pyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %+v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// if/how you use params or path or both is up to you, httpy doesn't care
	// shove that work off onto python or do it in golang, doesn't matter
	status, body, headers, err := httpy.Request(
		r.Method,
		r.URL.Path,
		string(body),
		nil, // params, map[string][]string
		r.URL.Query(),
		r.Header,
	)

	if err != nil {
		log.Printf("python error: %+v", err)
		http.Error(w, "failed to call python", http.StatusInternalServerError)
	}

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write: %+v", err)
	}
}

func main() {
	_, err := httpy.Init("worker", "go_init", "worker", "go_request")
	if err != nil {
		panic(fmt.Sprintf("failed to initialize python: %+v", err))
	}

	http.HandleFunc("/python", pyHandler)
	http.HandleFunc("/golang", goHandler)

	server := &http.Server{Addr: ":8080"}
	log.Fatal(server.ListenAndServe())
}
