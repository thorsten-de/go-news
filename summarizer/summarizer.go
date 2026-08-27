package summarizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/thorsten-de/go-news/domain"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Config struct {
	OllamaURL string
	Model     string
}

func DefaultConfig() Config {
	return Config{
		OllamaURL: "http://localhost:11434",
		Model:     "llama3.1",
	}
}

type Summarizer struct {
	llm llms.Model
}

// Creates a new Summarizer using the given configuration.
func New(config Config) (*Summarizer, error) {
	llm, err := ollama.New(
		ollama.WithServerURL(config.OllamaURL),
		ollama.WithModel(config.Model),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ollama client: %w", err)
	}
	return &Summarizer{llm: llm}, nil
}

// Summarize generates a natural language summary of the given articles using the LLM.
func (s *Summarizer) Summarize(ctx context.Context, articles []*domain.Article, focusTopic string) (string, error) {
	if len(articles) == 0 {
		return "", fmt.Errorf("no articles provided to summarize")
	}

	result, err := llms.GenerateFromSinglePrompt(
		ctx,
		s.llm,
		s.buildPrompt(articles, focusTopic),
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(300),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	return strings.TrimSpace(result), nil
}

const topicPrompt = `You are summarizing news articles related to: %[1]s

Focus on information relevant to "%[1]s". Create a cohesive summary that answers
what's new or important about this topic. If articles cover different aspects,
organize your summary logically.`

const generalPrompt = `You are summarizing recent news articles.

Create a brief news digest covering the key stories. If articles are related,
connect them. If they cover different topics, mention each briefly.
Keep it concise and informative.`

func (s *Summarizer) buildPrompt(articles []*domain.Article, focusTopic string) string {
	var sb strings.Builder

	if focusTopic != "" {
		fmt.Fprintf(&sb, topicPrompt, focusTopic)
	} else {
		sb.WriteString(generalPrompt)
	}

	sb.WriteString("\n\nArticles to summarize:\n\n")
	for i, article := range articles {
		fmt.Fprintf(&sb, "%d. Title: %s\n", i+1, article.Title)
		if article.Description != "" {
			fmt.Fprintf(&sb, "   Summary: %s\n", article.Description)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Your summary (2-3 paragraphs):\n")
	return sb.String()
}
