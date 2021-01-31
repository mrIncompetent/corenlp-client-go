// +build integration

package corenlp_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mrincompetent/corenlp-client-go"
)

func TestClient_Annotate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := corenlp.New("http://127.0.0.1:9000", http.DefaultClient)
	if err != nil {
		t.Fatalf("failed to create a client: %v", err)
	}

	const testText = "the quick brown fox jumps over the lazy dog"
	doc, err := client.Annotate(ctx, testText, []string{"tokenize", "ssplit", "pos"})
	if err != nil {
		t.Fatalf("failed to annotate text: %v", err)
	}

	if doc == nil {
		t.Fatal("returned document is nil")
	}

	if doc.Text == nil {
		t.Fatal("returned document does not contain text")
	}

	if *doc.Text != testText {
		t.Logf("Input:    %s", testText)
		t.Logf("Doc.Text: %s", *doc.Text)

		t.Error("The text in the returned document does not match with our input text")
	}
}
