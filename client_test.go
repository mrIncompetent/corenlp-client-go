package corenlp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestNew(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		_, err := New("invalid-address", http.DefaultClient)
		var uErr *url.Error

		t.Logf("Returned error: (Type: %T) %v", err, err)
		if !errors.As(err, &uErr) {
			t.Error("Returned error is not a URL error")
		}
	})

	t.Run("successful creation", func(t *testing.T) {
		client, err := New("http://127.0.0.1:9000", http.DefaultClient)
		if err != nil {
			t.Errorf("Expected err to be nil, got: %v", err)
		}
		if client == nil {
			t.Error("Returned client is nil")
		}
	})
}

func MustMarshalMessage(t testing.TB, msg proto.Message) []byte {
	b, err := marshalMessage(msg)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	return b
}

func TestClient_Annotate(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		text := "the quick brown fox jumps over the lazy dog"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			doc := &Document{Text: &text}
			encodedMessage := MustMarshalMessage(t, doc)
			if _, err := w.Write(encodedMessage); err != nil {
				t.Fatalf("failed to write response: %v", err)
			}
		}))
		defer ts.Close()

		client, err := New(ts.URL, ts.Client())
		if err != nil {
			t.Fatalf("Failed to create the client: %v", err)
		}

		if _, err := client.Annotate(context.Background(), text, []string{"tokenize", "ssplit", "pos"}); err != nil {
			t.Fatalf("Failed to annotate text: %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		t.Run("validation error", func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte("something failed")); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}))
			defer ts.Close()

			client, err := New(ts.URL, ts.Client())
			if err != nil {
				t.Fatalf("Failed to create the client: %v", err)
			}

			_, err = client.Annotate(context.Background(), "", []string{"tokenize", "ssplit", "pos"})
			var serr *ServerError
			if !errors.As(err, &serr) {
				t.Fatalf("Expected a ServerError to be returned, got: (Type: %T) %v", err, err)
			}

			if serr.statusCode != http.StatusInternalServerError {
				t.Errorf("Expected error to have statusCode %d, got %d", http.StatusInternalServerError, serr.statusCode)
			}

			if serr.body != "something failed" {
				t.Errorf("Expected error to have body '%s', got '%s'", "something failed", serr.body)
			}

			t.Log(serr.Error())
		})
	})
}
