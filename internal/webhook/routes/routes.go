package routes

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/object88/shipwright/internal/http/correlation"
	"github.com/object88/shipwright/internal/http/logging"
	"github.com/object88/shipwright/internal/http/router/route"
	"github.com/object88/shipwright/internal/webhook/handler"
)

func Defaults(logger logr.Logger, processor handler.Processor, secret string) []*route.Route {
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
					Handler: handler.New(processor, secret).Handle(),
					Methods: []string{http.MethodPost},
				},
			},
		},
	}
}
