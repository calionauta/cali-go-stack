package llm

import (
	"encoding/json"
	"io"

	ds "github.com/starfederation/datastar-go/datastar"
)

// StreamToSSE sends LLM tokens to a Datastar SSE connection.
func StreamToSSE(sse *ds.ServerSentEventGenerator, client *Client, prompt string) error {
	if err := sse.Send(ds.EventTypePatchSignals, []string{`{"streaming":true}`}); err != nil {
		return err
	}

	return client.ChatStream(sse.Context(), prompt, func(chunk string) error {
		data, _ := json.Marshal(map[string]string{"token": chunk})
		return sse.Send(ds.EventTypePatchSignals, []string{string(data)})
	})
}

// ConsumeStream reads an io.ReadCloser and sends chunks to a callback.
func ConsumeStream(stream io.ReadCloser, fn func(string) error) error {
	defer stream.Close()
	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			if err := fn(string(buf[:n])); err != nil {
				return err
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
