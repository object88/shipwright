package correlation

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	correlationIdHeaderKey = "Correlation-ID"
)

type contextKey struct{}

type CorrelationHandler struct {
	next http.Handler
}

func ConfigureMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		ch := CorrelationHandler{
			next: next,
		}
		return ch
	}
}

func (ch CorrelationHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	corr := req.Header.Get(correlationIdHeaderKey)
	if corr == "" {
		c, err := uuid.NewRandom()
		if err != nil {
			// Would hope this never happens, but who knows?
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Add("Retry-After", "1")
			w.Write([]byte("Unexpected error; please retry"))
			return
		}
		corr = c.String()
	}
	w.Header().Add(correlationIdHeaderKey, corr)
	req = req.WithContext(context.WithValue(req.Context(), contextKey{}, corr))
	ch.next.ServeHTTP(w, req)
}

func GetCorrelationId(ctx context.Context) string {
	x := ctx.Value(contextKey{})
	switch x0 := x.(type) {
	case string:
		return x0
	default:
		return ""
	}
}
