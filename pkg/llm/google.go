package llm

import (
	"context"

	"github.com/tmc/langchaingo/llms/googleai"
)

// newGoogleFactory returns a new LLMFactory for the Google AI provider.
// It uses a genericFactory to create a factory that can build a
// googleai.GoogleAI client.
func newGoogleFactory() LLMFactory {
	return &genericFactory[googleai.Option, *googleai.GoogleAI]{
		newFunc: googleAINew,
		optsFunc: genericFactoryOptions[googleai.Option]{
			Token: googleai.WithAPIKey,
			Model: googleai.WithDefaultModel,
		},
	}
}

// googleAINew is a wrapper around the googleai.New function that provides a
// consistent interface for the genericFactory.
func googleAINew(opts ...googleai.Option) (*googleai.GoogleAI, error) {
	return googleai.New(context.Background(), opts...)
}
