package llm

import (
	"context"
	"os"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/openai"
)

type Client struct {
	apiKey  string
	modelID string
}

func New(apiKey string) *Client {
	modelID := "gpt-4o"
	if v := os.Getenv("GOAI_MODEL"); v != "" {
		modelID = v
	}
	return &Client{apiKey: apiKey, modelID: modelID}
}

func (c *Client) model() provider.LanguageModel {
	if c.apiKey == "" {
		return openai.Chat(c.modelID)
	}
	return openai.Chat(c.modelID, openai.WithAPIKey(c.apiKey))
}

func (c *Client) Chat(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "(no API key configured)", nil
	}
	result, err := goai.GenerateText(ctx, c.model(), goai.WithPrompt(prompt))
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

type StreamFunc func(chunk string) error

func (c *Client) ChatStream(ctx context.Context, prompt string, fn StreamFunc) error {
	if c.apiKey == "" {
		return fn("(no API key configured)")
	}
	stream, err := goai.StreamText(ctx, c.model(), goai.WithPrompt(prompt))
	if err != nil {
		return err
	}
	for chunk := range stream.TextStream() {
		if err := fn(chunk); err != nil {
			return err
		}
	}
	return stream.Err()
}
