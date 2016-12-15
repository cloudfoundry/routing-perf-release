package main

import (
	"cpumonitor/stats"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// TODO
// security??

var (
	port        = flag.Int("port", 8080, "port for the server")
	runInterval = flag.Int("runInterval", 1000, "Sampling interval, how ofter to call CPU stats in millisecond")
	cpuInterval = flag.Int("cpuInterval", 0, "Time between each CPU stat in second, if 0 is given it will compare the current cpu times against the last call")
	perCPU      = flag.Bool("perCpu", false, "Percent calculates the percentage of cpu used either per CPU or combined")
)

func main() {
	flag.Parse()
	if *port == 0 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "-port must be within range 1024-65535")
		os.Exit(1)

	}
	if *runInterval < 1 {
		fmt.Fprintf(os.Stderr, "-runInterval must be above 1")
		os.Exit(1)

	}

	startServer()
}

func startServer() {
	log.Printf("CPU monitor is running on %d \n", *port)
	cpuCalculator := new(stats.CPUOps)
	runInterval := time.Duration(int32(*runInterval)) * time.Millisecond
	cpuInterval := time.Duration(int32(*cpuInterval)) * time.Second
	cpuCollector := stats.NewCPUCollector(cpuCalculator, stats.DefaultTicker(), runInterval, cpuInterval, *perCPU)
	statsHandler := stats.NewStatHandler(cpuCollector)
	http.HandleFunc("/start", statsHandler.Start)
	http.HandleFunc("/stop", statsHandler.Stop)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
