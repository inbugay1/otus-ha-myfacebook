package httphandler

import (
	"encoding/json"
	"net/http"
)

type Health struct {
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Health) Handle(responseWriter http.ResponseWriter, _ *http.Request) error {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	responseWriter.WriteHeader(http.StatusOK)

	err := json.NewEncoder(responseWriter).Encode(healthResponse{
		Status: "OK",
	})
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return nil
}
