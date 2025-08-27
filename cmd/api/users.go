package main

import (
	"net/http"
	"strconv"

	"com.github/jrovieri/golang/social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, err := app.store.Users.Get(r.Context(), userID)
	if err != nil {
		switch err {
		case store.ErrResourceNotFound:
			app.notFound(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, &user); err != nil {
		app.internalServerError(w, r, err)
	}
}
