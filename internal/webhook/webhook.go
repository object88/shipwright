package webhook

import (
	"context"

	"github.com/google/go-github/v37/github"
)

type Processor struct{}

func New() *Processor {
	return &Processor{}
}

func (p *Processor) Process(ctx context.Context, evt *github.CheckRunEvent) error {
	return nil
}
