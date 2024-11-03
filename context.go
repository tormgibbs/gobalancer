package main

import (
	"context"
	"net/http"
)

type contextKey string

type Retries int

const retryContextKey = contextKey("retries")


func (app *application) contextSetRetries(r *http.Request, retries Retries) *http.Request {
	ctx := context.WithValue(r.Context(), retryContextKey, retries)
	return r.WithContext(ctx)
}

func (app *application) contextGetRetries(r *http.Request) Retries {
	retries, ok := r.Context().Value(retryContextKey).(Retries)
	if !ok {
		panic("missing retries value in request context")
	}
	return retries
}