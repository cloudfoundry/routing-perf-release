package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/c9s/goprocinfo/linux"
)

func getCPU(w http.ResponseWriter, r *http.Request) {
	stat, _ := linux.ReadStat("/proc/stat")
	body, _ := json.Marshal(stat.CPUStatAll)
	w.Write(body)
}

func main() {
	http.HandleFunc("/", getCPU)             // set router
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
