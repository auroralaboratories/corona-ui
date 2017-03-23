package util

import (
	"encoding/json"
	"net/http"
	"time"
)

const ApplicationName = `corona-ui`
const ApplicationSummary = `A Webkit-based GTK window for running desktop web applications`
const ApplicationVersion = `0.2.0`

var StartedAt = time.Now()

func Respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set(`Content-Type`, `application/json`)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
