package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")

	router.HandleFunc("/records", getRecords).Methods("GET")
	router.HandleFunc("/records/{uid}", getBucketRecords).Methods("GET")

	return router
}
