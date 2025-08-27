package main

import (
	"context"
	"net/http"
	"strconv"

	"com.github/jrovieri/golang/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type UserKey string

const userCtx UserKey = "user"

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, &user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	followerUser := getUserFromContext(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	err := app.store.Users.Follow(r.Context(), followerUser.ID, payload.UserID)
	if err != nil {
		switch err {
		case store.ErrConflict:
			app.conflict(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	unfollowedUser := getUserFromContext(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	err := app.store.Users.UnFollow(r.Context(), unfollowedUser.ID, payload.UserID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idParam := chi.URLParam(r, "userID")
		userID, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.Get(ctx, userID)
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
