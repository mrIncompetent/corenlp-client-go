# corenlp-client-go

A [CoreNLP](https://stanfordnlp.github.io/CoreNLP/) client written in Go which uses [Protobuf](https://developers.google.com/protocol-buffers) for message serialization.

## Usage 

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mrincompetent/corenlp-client-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := corenlp.New("http://127.0.0.1:9000", http.DefaultClient)
	if err != nil {
		log.Fatalf("failed to create a client: %v", err)
	}

	doc, err := client.Annotate(ctx, "this is a wonderful library", []string{"sentiment"})
	if err != nil {
		log.Fatalf("failed to annotate text: %v", err)
	}

	log.Println(*doc.Sentence[0].Sentiment) // Prints: Positive
}
```
