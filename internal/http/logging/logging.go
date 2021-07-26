package logging

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/object88/shipwright/internal/http/correlation"
)

func ConfigureLoggingMiddleware(logger logr.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		lch := LogContextHandler{
			logger: logger,
			next:   next,
		}
		return handlers.LoggingHandler((&Writer{Log: logger}).Out(), &lch)
	}
}

type LogContextHandler struct {
	logger logr.Logger
	next   http.Handler
}

func (lch *LogContextHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := lch.logger

	corr := correlation.GetCorrelationId(req.Context())
	if corr != "" {
		logger = logger.WithValues("correlationId", corr)
	}

	req = req.WithContext(logr.NewContext(req.Context(), logger))
	lch.next.ServeHTTP(w, req)
}
