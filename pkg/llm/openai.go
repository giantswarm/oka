package llm

import "github.com/tmc/langchaingo/llms/openai"

// newOpenAIFactory returns a new LLMFactory for the OpenAI provider.
// It uses a genericFactory to create a factory that can build an
// openai.LLM client.
func newOpenAIFactory() LLMFactory {
	return &genericFactory[openai.Option, *openai.LLM]{
		newFunc: openai.New,
		optsFunc: genericFactoryOptions[openai.Option]{
			Token: openai.WithToken,
			Model: openai.WithModel,
		},
	}
}
