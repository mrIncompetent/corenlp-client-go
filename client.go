package corenlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

type RequestExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	address    string
	httpClient RequestExecutor
}

func New(address string, httpClient RequestExecutor) (*Client, error) {
	u, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address '%s': %w", address, err)
	}

	return &Client{
		address:    u.String(),
		httpClient: httpClient,
	}, nil
}

type requestProperties struct {
	OutputFormat    string `json:"outputFormat,omitempty"`
	Serializer      string `json:"serializer,omitempty"`
	Annotators      string `json:"annotators,omitempty"`
	InputFormat     string `json:"inputFormat,omitempty"`
	InputSerializer string `json:"inputSerializer,omitempty"`
}

const (
	dataFormatSerialized   = "serialized"
	dataSerializerProtobuf = "edu.stanford.nlp.pipeline.ProtobufAnnotationSerializer"
)

func (c *Client) newRequest(ctx context.Context, text string, annotators []string) (*http.Request, error) {
	properties := &requestProperties{
		OutputFormat:    dataFormatSerialized,
		Serializer:      dataSerializerProtobuf,
		Annotators:      strings.Join(annotators, ","),
		InputFormat:     dataFormatSerialized,
		InputSerializer: dataSerializerProtobuf,
	}

	requestPropertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request properties to JSON: %w", err)
	}

	u, err := url.Parse(c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address '%s': %w", c.address, err)
	}

	q := u.Query()
	q.Set("properties", string(requestPropertiesJSON))
	u.RawQuery = q.Encode()

	doc := &Document{
		Text: &text,
	}

	encodedMessage, err := marshalMessage(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to bytes: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(encodedMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to create go request: %w", err)
	}

	req.Header.Set("content-type", "application/x-protobuf")

	return req, nil
}

func (c *Client) Annotate(ctx context.Context, text string, annotators []string) (*Document, error) {
	req, err := c.newRequest(ctx, text, annotators)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode >= 400 {
		return nil, &ServerError{
			statusCode: res.StatusCode,
			body:       string(resBytes),
		}
	}

	doc := &Document{}
	if err := unmarshalMessage(resBytes, doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response into document: %w", err)
	}

	return doc, nil
}

func marshalMessage(message proto.Message) ([]byte, error) {
	var buf []byte
	buf = protowire.AppendVarint(buf, uint64(proto.Size(message)))

	msgBuf, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message to wire format: %w", err)
	}

	buf = append(buf, msgBuf...)

	return buf, nil
}

func unmarshalMessage(b []byte, m proto.Message) error {
	_, varintLen := protowire.ConsumeVarint(b)
	if varintLen < 0 {
		return fmt.Errorf("failed to read message size varint: %w", protowire.ParseError(varintLen))
	}

	out, err := proto.UnmarshalOptions{
		AllowPartial: true,
		Merge:        true,
	}.UnmarshalState(protoiface.UnmarshalInput{
		Buf:     b[varintLen:],
		Message: m.ProtoReflect(),
	})
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	if out.Flags&protoiface.UnmarshalInitialized > 0 {
		return nil
	}

	if err := proto.CheckInitialized(m); err != nil {
		return fmt.Errorf("failed to verify all fields are initialized: %w", err)
	}

	return nil
}
