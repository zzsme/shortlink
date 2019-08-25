package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	validator "gopkg.in/validator.v2"
)

// App encapsulates Env Router and middleware
type App struct {
	Router *mux.Router
}

type shortenReq struct {
	URL                 string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes"`
}

type shorlinkResp struct {
	Shortlink string `json:"shorlink"`
}

// Initialize is initialization of app
func (a *App) Initialize() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.initializeRouters()
}

func (a *App) initializeRouters() {
	a.Router.HandleFunc("/api/shorten", a.createShortlink).Methods("POST")
	a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("GET")
	a.Router.HandleFunc("/{showlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")
}

func (a *App) createShortlink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}
	if err := validator.Validate(req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("validate parameters failed %v", r.Body)})
		return
	}
	defer r.Body.Close()

	fmt.Printf("%v\n", req)
}
func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {

	vals := r.URL.Query()
	s := vals.Get("shortlink")

	fmt.Printf("%s\n", s)
}
func (a *App) redirect(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	fmt.Printf("%s\n", vars["shortlink"])
}

// Run App
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError))

	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
