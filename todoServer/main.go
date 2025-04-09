package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Println("Starting server")
	host := flag.String("h", "localhost", "Server host")
	port := flag.Int("p", 8080, "Server port")
	todoFile := flag.String("f", "todoServer.json", "Todo JSON file")
	flag.Parse()
	log.Printf("host: %s, port: %d, todoFile: %s\n", *host, *port, *todoFile)

	s := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *host, *port),
		Handler:      newMux(*todoFile),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
