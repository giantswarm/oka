package llm

import "github.com/tmc/langchaingo/llms/anthropic"

// newAnthropicFactory creates a new LLMFactory for the Anthropic provider.
// It uses the genericFactory to create a factory that can build an
// anthropic.LLM client.
func newAnthropicFactory() LLMFactory {
	return &genericFactory[anthropic.Option, *anthropic.LLM]{
		newFunc: anthropic.New,
		optsFunc: genericFactoryOptions[anthropic.Option]{
			Token: anthropic.WithToken,
			Model: anthropic.WithModel,
		},
	}
}
