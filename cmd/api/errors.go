package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error -  %s %s: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request -  %s %s: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found -  %s %s: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *application) conflict(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict -  %s %s: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusConflict, err.Error())
}
