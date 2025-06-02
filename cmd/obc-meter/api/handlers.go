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

		filters.FromPeriod = &t
	}

	if to_period != "" {
		t, err := time.Parse(time.RFC3339, to_period)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.ToPeriod = &t
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

		filters.FromPeriod = &t
	}

	if to_period != "" {
		t, err := time.Parse(time.RFC3339, to_period)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_period'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.ToPeriod = &t
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

func getRuns(w http.ResponseWriter, r *http.Request) {
	filters := db.GetRunsArgs{}

	query := r.URL.Query()

	ids := query.Get("ids")
	trigger := query.Get("trigger")
	from_time := query.Get("from_time")
	to_time := query.Get("to_time")

	if ids != "" {
		idsList := strings.Split(ids, ",")
		filters.Ids = &idsList
	}

	if trigger != "" {
		filters.Trigger = &trigger
	}

	if from_time != "" {
		t, err := time.Parse(time.RFC3339, from_time) // 2006-01-02T15:04:05Z07:00
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			// format response later
			fmt.Fprintf(w, "Failed to parse query parameter 'from_time'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.FromTime = &t
	}

	if to_time != "" {
		t, err := time.Parse(time.RFC3339, to_time)
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to parse query parameter 'to_time'. It must be in RFC3339 format (%v)\n", time.Now().Format(time.RFC3339))
			return
		}

		filters.ToTime = &t
	}

	runs, err := db.GetRuns(filters)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve runs")
		return
	}

	json, err := json.Marshal(runs)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve runs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func getRun(w http.ResponseWriter, r *http.Request) {
	filters := db.GetRunsArgs{}
	vars := mux.Vars(r)
	filters.Ids = &[]string{vars["id"]}

	runs, err := db.GetRuns(filters)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to retrieve run")
		return
	}

	if runs == nil || len(*runs) < 1 {
		w.WriteHeader(400)
		fmt.Println(err)
		fmt.Fprintf(w, "Run not found")
		return
	}

	runsVal := *runs
	run := runsVal[0]

	json, err := json.Marshal(run)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to parse run")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
