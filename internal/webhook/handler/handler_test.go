package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v37/github"
)

const secret = "foo"

func Test_Handler(t *testing.T) {
	checkRunId := int64(13)

	p := processor{
		t: t,
		p: func(t *testing.T, ctx context.Context, evt *github.CheckRunEvent) error {
			if evt == nil {
				return fmt.Errorf("Did not get evt")
			} else if evt.CheckRun == nil {
				return fmt.Errorf("Did not get check run")
			} else if evt.CheckRun.ID == nil {
				return fmt.Errorf("Did not get check run Id")
			} else if *evt.CheckRun.ID != checkRunId {
				return fmt.Errorf("Got unexpected check run Id: %d", *evt.CheckRun.ID)
			}
			return nil
		},
	}
	ts := httptest.NewServer(New(&p, secret).Handle())
	defer ts.Close()

	cre := github.CheckRunEvent{
		CheckRun: &github.CheckRun{
			ID: &checkRunId,
		},
	}
	raw, err := json.Marshal(cre)
	r := createRequest(ts.URL, "check_run", raw)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Errorf("Unexpected error processing request: %w", err)
	}

	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	if string(rawRespBody) != "processed check run" {
		t.Errorf("Unexpected body")
	}
}

func Test_Handler_OtherWebhook(t *testing.T) {
	p := processor{}
	ts := httptest.NewServer(New(&p, secret).Handle())
	defer ts.Close()

	cre := github.PullRequestEvent{}
	raw, err := json.Marshal(cre)
	r := createRequest(ts.URL, "pull_request", raw)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Errorf("Unexpected error processing request: %w", err)
	}

	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	if string(rawRespBody) != "OK" {
		t.Errorf("Unexpected body: '%s'", rawRespBody)
	}
}

func Test_Handler_ProcessingFails(t *testing.T) {
	p := processor{
		t: t,
		p: func(t *testing.T, ctx context.Context, evt *github.CheckRunEvent) error {
			return fmt.Errorf("NOTOK")
		},
	}

	ts := httptest.NewServer(New(&p, secret).Handle())
	defer ts.Close()

	cre := github.PullRequestEvent{}
	raw, err := json.Marshal(cre)
	r := createRequest(ts.URL, "check_run", raw)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Errorf("Unexpected error processing request: %w", err)
	}

	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	if string(rawRespBody) != "processing failed" {
		t.Errorf("Unexpected body: '%s'", rawRespBody)
	}
}

func createRequest(url string, eventType string, raw []byte) *http.Request {
	// Copied from https://github.com/rjz/githubhook/blob/master/githubhook_test.go
	r, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))

	dst := make([]byte, 40)
	computed := hmac.New(sha1.New, []byte(secret))
	computed.Write(raw)
	hex.Encode(dst, computed.Sum(nil))
	signature := "sha1=" + string(dst)

	r.Header.Add("x-hub-signature", signature)
	r.Header.Add("x-github-event", eventType)
	r.Header.Add("x-github-delivery", "bogus id")

	return r
}

type processor struct {
	t *testing.T
	p func(t *testing.T, ctx context.Context, evt *github.CheckRunEvent) error
}

func (p *processor) Process(ctx context.Context, evt *github.CheckRunEvent) error {
	return p.p(p.t, ctx, evt)
}
