package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func newMux(todoFile string) http.Handler {
	log.Printf("newMux: %s\n", todoFile)
	m := http.NewServeMux()
	mutex := &sync.Mutex{}

	m.HandleFunc("/", rootHandler)

	log.Printf("Create todoRouter serving %s\n", todoFile)
	tr := todoRouter(todoFile, mutex)

	log.Println("Setting up hadles for /todo")
	m.Handle("/todo", http.StripPrefix("/todo", tr))
	m.Handle("/todo/", http.StripPrefix("/todo/", tr))

	return m
}

func replyTextContent(w http.ResponseWriter, r *http.Request, status int, content string) {
	log.Printf("replyTextContent: %d\n", status)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func replyJSONContent(w http.ResponseWriter, r *http.Request, status int, resp *todoResponse) {
	log.Printf("replyJSONContent: %d\n", status)
	body, err := json.Marshal(resp)
	if err != nil {
		replyError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func replyError(w http.ResponseWriter, r *http.Request, status int, message string) {
	log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, message)
	http.Error(w, http.StatusText(status), status)
}
