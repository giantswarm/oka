// Package llm provides an interface for interacting with Large Language Models
// (LLMs). It includes a factory for creating LLM clients and methods for
// generating text completions.
package llm

import (
	"github.com/tmc/langchaingo/llms"

	"github.com/giantswarm/oka/pkg/config"
)

// New creates a new llms.Model based on the provided configuration. It uses a
// factory to build the appropriate LLM client (e.g., OpenAI, Anthropic) and
// returns a generic `llms.Model` interface.
func New(conf config.LLM) (llms.Model, error) {
	factory, err := NewFactory(conf)
	if err != nil {
		return nil, err
	}

	model, err := factory.Build(conf)
	if err != nil {
		return nil, err
	}

	return model, nil
}
