package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v37/github"
	"github.com/rjz/githubhook"
)

type Handler struct {
	secret []byte
}

func New(secret string) *Handler {
	return &Handler{
		secret: []byte(secret),
	}
}

func (h *Handler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logr.FromContextOrDiscard(r.Context())
		hook, err := githubhook.Parse(h.secret, r)
		if err != nil {
			logger.Error(err, "Failed to parse webhook request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		evt := github.PullRequestEvent{}
		if err := json.Unmarshal(hook.Payload, &evt); err != nil {
			logger.Error(err, "Invalid JSON")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte("OK"))
	}
}
