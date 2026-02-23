package middleware

import (
	"context"
	"net/http"
	"testing"
)

func TestWithRequestIDAndContextExtraction(t *testing.T) {
	ctx := WithRequestID(context.Background(), "abc-123")
	requestID, ok := RequestIDFromContext(ctx)
	if !ok {
		t.Fatal("expected request id in context")
	}
	if requestID != "abc-123" {
		t.Fatalf("unexpected request id: %s", requestID)
	}
}

func TestSetRequestIDHeaderFromContext(t *testing.T) {
	ctx := WithRequestID(context.Background(), "rid-42")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	SetRequestIDHeaderFromContext(req)
	if got := req.Header.Get(RequestIDHeader); got != "rid-42" {
		t.Fatalf("expected %s header to be rid-42, got %q", RequestIDHeader, got)
	}
}

func TestNewRequestWithRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "rid-new-request")
	req, err := NewRequestWithRequestID(ctx, http.MethodPost, "https://example.com/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get(RequestIDHeader); got != "rid-new-request" {
		t.Fatalf("expected propagated request id, got %q", got)
	}
}
