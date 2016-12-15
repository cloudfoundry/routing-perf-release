package stats

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	ErrMultipleCollector   = errors.New("CPU collector already running")
	ErrCollectorNotRunning = errors.New("CPU collector is not started")
)

type Handler struct {
	collector    Collector
	startCounter int32
}

func NewStatHandler(c Collector) *Handler {
	return &Handler{
		collector: c,
	}
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	count := atomic.LoadInt32(&h.startCounter)
	if count != 0 {
		http.Error(w, ErrMultipleCollector.Error(), http.StatusBadRequest)
		return
	}
	atomic.AddInt32(&h.startCounter, 1)
	//	TODO: need to handle the err
	go h.collector.Run()
	log.Printf("started cpu collector at %s", time.Now().Format(time.RFC3339))
	w.Write([]byte("Collecting CPU stats\n"))
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	count := atomic.LoadInt32(&h.startCounter)
	if count != 1 {
		http.Error(w, ErrCollectorNotRunning.Error(), http.StatusBadRequest)
		return
	}

	cpuStats := h.collector.Result()
	log.Printf("stoped cpu collector at %s", time.Now().Format(time.RFC3339))
	json, err := json.Marshal(cpuStats)
	if err != nil {
		log.Printf("Failed to marshal %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	atomic.AddInt32(&h.startCounter, -1)
}
