package routes

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/object88/shipwright/internal/http/correlation"
	"github.com/object88/shipwright/internal/http/logging"
	"github.com/object88/shipwright/internal/http/router/route"
	"github.com/object88/shipwright/internal/webhook"
)

func Defaults(logger logr.Logger, secret string) []*route.Route {
	return []*route.Route{
		{
			Path: "/v1/api",
			Middleware: []mux.MiddlewareFunc{
				correlation.ConfigureMiddleware(),
				logging.ConfigureLoggingMiddleware(logger),
			},
			Subroutes: []*route.Route{
				{
					Path:    "/webhook",
					Handler: webhook.New(secret).Handle(),
					Methods: []string{http.MethodPost},
				},
			},
		},
	}
}
