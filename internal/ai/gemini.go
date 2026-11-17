package ai

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Client struct {
	genaiClient *genai.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	return &Client{genaiClient: client}, nil
}

func (c *Client) SummarizeContent(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Truncate text if it's exceedingly long to save tokens
	if len(text) > 30000 {
		text = text[:30000]
	}

	model := "gemini-2.5-flash"
	prompt := fmt.Sprintf("Summarize the following content concisely in one or two paragraphs:\n\n%s", text)

	resp, err := c.genaiClient.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		return part.Text, nil
	}
	
	return "", fmt.Errorf("unexpected part type from gemini")
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, nil
	}

	// Truncate for embedding
	if len(text) > 8000 {
		text = text[:8000]
	}

	model := "text-embedding-004"
	resp, err := c.genaiClient.Models.EmbedContent(ctx, model, genai.Text(text), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Embeddings) > 0 {
		return resp.Embeddings[0].Values, nil
	}

	return nil, fmt.Errorf("no embedding generated")
}

func (c *Client) AnswerQueryWithContext(ctx context.Context, query string, contextText string) (string, error) {
	model := "gemini-2.5-flash"
	// Every context block is labelled with its post id, and the model is required to
	// cite by that id rather than restate the source. internal/verify then replaces
	// each pointer with canonical text from the database, so a citation cannot be
	// wrong — an unresolvable one fails the whole answer.
	prompt := fmt.Sprintf(`You are answering questions about a user's tracked sources.

Rules:
- Use ONLY the context below. Never rely on outside knowledge.
- Every claim MUST cite its source inline using the exact form [post:<id>], copying the id verbatim from the context.
- Never invent an id. Never cite an id that does not appear in the context.
- Do not write out source titles or URLs yourself; cite the id and nothing more.
- If the context does not answer the question, reply exactly: I don't have enough information to answer that.

Context:
%s

Question: %s`, contextText, query)

	resp, err := c.genaiClient.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no answer generated")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		return part.Text, nil
	}
	
	return "", fmt.Errorf("unexpected part type from gemini")
}
