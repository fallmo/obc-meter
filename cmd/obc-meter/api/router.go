package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fallmo/obc-meter/cmd/obc-meter/utils"
	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")

	router.HandleFunc("/records/{uid}", getRecords).Methods("GET")

	return router
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	filters := utils.GetBucketRecordsArgs{}
	vars := mux.Vars(r)
	filters.Uid = vars["uid"]

	query := r.URL.Query()

	from_period := query.Get("from_period")
	to_period := query.Get("to_period")

	if from_period != "" {
		t, err := time.Parse(time.RFC3339, from_period) // 2006-01-02T15:04:05Z07:00
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			// format response later
			fmt.Fprintf(w, "Failed to parse query parameter 'from_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
		}

		filters.From_period = &t
	}

	if to_period != "" {
		t, err := time.Parse(time.RFC3339, to_period)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
		}

		filters.To_period = &t
	}

	records, err := utils.GetBucketUsageRecords(filters)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve records")
	}

	w.WriteHeader(200)

	json, err := json.Marshal(records)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

}
