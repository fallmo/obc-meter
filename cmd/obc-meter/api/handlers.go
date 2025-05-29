package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fallmo/obc-meter/cmd/obc-meter/db"
	"github.com/gorilla/mux"
)

func getRecords(w http.ResponseWriter, r *http.Request) {
	filters := db.GetRecordsArgs{}

	query := r.URL.Query()

	uids := query.Get("uids")
	run_ids := query.Get("run_ids")
	from_period := query.Get("from_period")
	to_period := query.Get("to_period")

	if uids != "" {
		uidsList := strings.Split(uids, ",")
		filters.Uids = &uidsList
	}

	if run_ids != "" {
		runIds := strings.Split(run_ids, ",")
		filters.RunIds = &runIds
	}

	if from_period != "" {
		t, err := time.Parse(time.RFC3339, from_period) // 2006-01-02T15:04:05Z07:00
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			// format response later
			fmt.Fprintf(w, "Failed to parse query parameter 'from_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.From_period = &t
	}

	if to_period != "" {
		t, err := time.Parse(time.RFC3339, to_period)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.To_period = &t
	}

	records, err := db.GetUsageRecords(filters)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve records")
		return
	}

	json, err := json.Marshal(records)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

}

func getBucketRecords(w http.ResponseWriter, r *http.Request) {
	filters := db.GetBucketRecordsArgs{}
	vars := mux.Vars(r)
	filters.Uid = vars["uid"]

	query := r.URL.Query()

	run_ids := query.Get("run_ids")
	from_period := query.Get("from_period")
	to_period := query.Get("to_period")

	if run_ids != "" {
		runIds := strings.Split(run_ids, ",")
		filters.RunIds = &runIds
	}

	if from_period != "" {
		t, err := time.Parse(time.RFC3339, from_period) // 2006-01-02T15:04:05Z07:00
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			// format response later
			fmt.Fprintf(w, "Failed to parse query parameter 'from_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.From_period = &t
	}

	if to_period != "" {
		t, err := time.Parse(time.RFC3339, to_period)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.To_period = &t
	}

	records, err := db.GetBucketUsageRecords(filters)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve records")
		return
	}

	// w.WriteHeader(200)

	json, err := json.Marshal(records)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

}
