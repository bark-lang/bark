package modules

import (
	"bytes"
	"io"
	"testing"
)

func TestReadResponseBodyWithinLimit(t *testing.T) {
	// Create a body that's within the limit
	data := bytes.Repeat([]byte("a"), 1024) // 1KB
	body := io.NopCloser(bytes.NewReader(data))

	result, err := readResponseBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1024 {
		t.Errorf("expected 1024 bytes, got %d", len(result))
	}
}

func TestReadResponseBodyAtLimit(t *testing.T) {
	// Create a body that's exactly at the limit
	data := bytes.Repeat([]byte("a"), MaxResponseBodySize)
	body := io.NopCloser(bytes.NewReader(data))

	result, err := readResponseBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != MaxResponseBodySize {
		t.Errorf("expected %d bytes, got %d", MaxResponseBodySize, len(result))
	}
}

func TestReadResponseBodyOverLimit(t *testing.T) {
	// Create a body that exceeds the limit
	data := bytes.Repeat([]byte("a"), MaxResponseBodySize+1)
	body := io.NopCloser(bytes.NewReader(data))

	_, err := readResponseBody(body)
	if err == nil {
		t.Fatal("expected error for body exceeding limit")
	}

	if err != ErrResponseTooLarge {
		t.Errorf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestMaxResponseBodySizeValue(t *testing.T) {
	// Verify the constant is 10MB
	expected := 10 * 1024 * 1024
	if MaxResponseBodySize != expected {
		t.Errorf("MaxResponseBodySize should be 10MB (%d), got %d", expected, MaxResponseBodySize)
	}
}
