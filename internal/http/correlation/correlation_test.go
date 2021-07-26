package correlation

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// Test_Correlation_New ensures that a correlation ID is generated if missing,
// and added to the response.
func Test_Correlation_New(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})
	outerHandler := ConfigureMiddleware()(innerHandler)
	ts := httptest.NewServer(outerHandler)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Unexpected error: %w", err)
	}
	respCorr := res.Header.Get(correlationIdHeaderKey)
	if respCorr == "" {
		t.Errorf("Returned correlation Id is empty string")
	}
	if respCorr == uuid.Nil.String() {
		t.Errorf("Returned all-zero correlation ID")
	}
}

// Test_Correlation_New_OnContext ensures that a newly generated correlation ID
// is put on the context for inner handlers
func Test_Correlation_New_OnContext(t *testing.T) {
	contextCorr := ""

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextCorr = GetCorrelationId(r.Context())
	})
	outerHandler := ConfigureMiddleware()(innerHandler)
	ts := httptest.NewServer(outerHandler)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Unexpected error: %w", err)
	}

	if contextCorr == "" {
		t.Errorf("Returned correlation Id is empty string")
	}
	if contextCorr == uuid.Nil.String() {
		t.Errorf("Returned all-zero correlation ID")
	}
}

// Test_Correlation_AlreadyPresent ensures that a correlation ID is not
// replaced when the request has one.
func Test_Correlation_AlreadyPresent(t *testing.T) {
	const (
		corr = "1234"
	)
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})
	outerHandler := ConfigureMiddleware()(innerHandler)
	ts := httptest.NewServer(outerHandler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Test failurre: failed to create http request: %w", err)
	}
	req.Header.Add(correlationIdHeaderKey, corr)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Unexpected error: %w", err)
	}
	respCorr := res.Header.Get(correlationIdHeaderKey)
	if respCorr != corr {
		t.Errorf("Returned correlation Id is incorrect: '%s'", respCorr)
	}
}

// Test_Correlation_AlreadyPresent_OnContext ensures that if a correlation ID
// is not on the incoming request, it is put on the context for inner handlers.
func Test_Correlation_AlreadyPresent_OnContext(t *testing.T) {
	const (
		corr = "1234"
	)
	contextCorr := ""

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextCorr = GetCorrelationId(r.Context())
	})
	outerHandler := ConfigureMiddleware()(innerHandler)
	ts := httptest.NewServer(outerHandler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Test failure: failed to create http request: %w", err)
	}
	req.Header.Add(correlationIdHeaderKey, corr)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Test failure: failed to perform GET request: %w", err)
	}

	if contextCorr != corr {
		t.Errorf("Correlation Id on inner context is wrong: '%s'", contextCorr)
	}
}

func Test_Correlation_NoCorrelationOnContext(t *testing.T) {
	ctx := context.Background()
	contextCorr := GetCorrelationId(ctx)
	if contextCorr != "" {
		t.Errorf("Got unexpected correlation Id from fresh context: %s", contextCorr)
	}
}
