package router

import (
	"net/http"

	"github.com/cloudretic/router/pkg/middleware"
	"github.com/cloudretic/router/pkg/route"
)

type Router interface {
	Attach(mw middleware.Middleware)
	AddRoute(r route.Route, h http.Handler)
	AddNotFound(h http.Handler)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

func New(with Router, cfs ...ConfigFunc) (Router, error) {
	for _, cf := range cfs {
		err := cf(with)
		if err != nil {
			return nil, err
		}
	}
	return with, nil
}

func Declare(with Router, cfs ...ConfigFunc) Router {
	for _, cf := range cfs {
		err := cf(with)
		if err != nil {
			panic(err)
		}
	}
	return with
}
