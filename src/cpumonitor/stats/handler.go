package stats

import (
	"net/http"
)

type Handler struct {
	collector Collector
}

func NewStatHandler(c Collector) *Handler {
	return &Handler{
		collector: c,
	}
}

func (c *Handler) Start(w http.ResponseWriter, r *http.Request) {
	// TODO: will return an error so we should handle that case
	//	err := c.cpuStats.Run()
	//	if err != nil {
	//		w.WriteHeader(http.StatusBadRequest)
	//	}
}
