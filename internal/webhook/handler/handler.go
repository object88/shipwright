package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v37/github"
	"github.com/rjz/githubhook"
)

type Processor interface {
	Process(context.Context, *github.CheckRunEvent) error
}

type Handler struct {
	processor Processor
	secret    []byte
}

func New(processor Processor, secret string) *Handler {
	return &Handler{
		processor: processor,
		secret:    []byte(secret),
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

		switch hook.Event {
		case "check_run":
			evt := github.CheckRunEvent{}
			if err := json.Unmarshal(hook.Payload, &evt); err != nil {
				logger.Error(err, "Invalid JSON")
				w.Write([]byte("invalid json"))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := h.processor.Process(r.Context(), &evt); err != nil {
				logger.Error(err, "Failed to process webhook: %w", err)
				w.Write([]byte("processing failed"))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Write([]byte("processed check run"))

		default:
			// Ignore this hook.
			w.Write([]byte("OK"))
		}
	}
}
