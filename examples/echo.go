package examples

import (
	"net/http"

	"github.com/cloudretic/matcha/pkg/rctx"
	"github.com/cloudretic/matcha/pkg/router"
)

func echoAdmin(w http.ResponseWriter, req *http.Request) {
	name := rctx.GetParam(req.Context(), "name")
	w.Write([]byte("Hello, admin " + name + "!"))
}

func echo(w http.ResponseWriter, req *http.Request) {
	name := rctx.GetParam(req.Context(), "name")
	w.Write([]byte("Hello, " + name + "!"))
}

func EchoExample() {
	rt := router.Default()
	rt.HandleFunc(http.MethodGet, "/hello/[name]{admin:.+}", echoAdmin)
	rt.HandleFunc(http.MethodGet, "/hello/[name]", echo)
	http.ListenAndServe(":3000", rt)
}