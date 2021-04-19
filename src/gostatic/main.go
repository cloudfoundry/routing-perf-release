package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var data string

func main() {
	s := os.Getenv("RESPONSE_SIZE")
	if s == "" {
		s = "1"
	}
	responseSize, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Error parsing response size: %s", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("The PORT environment variable is empty")
	}
	data = strings.Repeat("Z", responseSize*1024)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, data)
	})

	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: h2c.NewHandler(handler, h2s),
	}

	if err := h1s.ListenAndServe(); err != nil {
		panic(err)
	}
}
