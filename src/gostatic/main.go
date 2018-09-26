package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, data)
	})

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
