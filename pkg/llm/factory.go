package llm

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"

	"github.com/giantswarm/oka/pkg/config"
)

// LLMFactory is an interface for building LLM models. It abstracts the creation
// of specific LLM clients, allowing for a common interface across different
// providers.
type LLMFactory interface {
	Build(llmConfig config.LLM) (llms.Model, error)
}

// NewFactory returns a new LLMFactory based on the provider specified in the
// configuration. It supports "anthropic", "google", and "openai" providers.
func NewFactory(conf *config.Config) (LLMFactory, error) {
	switch conf.LLM.Provider {
	case "anthropic":
		return newAnthropicFactory(), nil
	case "google":
		return newGoogleFactory(), nil
	case "openai":
		return newOpenAIFactory(), nil
	}

	return nil, fmt.Errorf("unknown LLM provider: %s", conf.LLM.Provider)
}

// genericFactory is a generic implementation of the LLMFactory interface.
// It is used to create LLM clients for different providers using a common
// set of options.
type genericFactory[O any, M llms.Model] struct {
	newFunc  func(...O) (M, error)
	optsFunc genericFactoryOptions[O]
}

// genericFactoryOptions holds the functions for creating provider-specific
// options, such as setting the API token or model name.
type genericFactoryOptions[O any] struct {
	Token func(string) O
	Model func(string) O
}

// Build creates a new LLM model using the provided configuration.
// It constructs the necessary options and passes them to the newFunc.
func (f *genericFactory[O, M]) Build(llmConfig config.LLM) (llms.Model, error) {
	opts := f.buildOpts(llmConfig)

	model, err := f.newFunc(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM model: %w", err)
	}

	return model, nil
}

// buildOpts constructs the list of options for the LLM client based on the
// provided configuration.
func (f *genericFactory[O, M]) buildOpts(llmConfig config.LLM) []O {
	var opts []O

	f.buildOpt(&opts, llmConfig.Token, f.optsFunc.Token)
	f.buildOpt(&opts, llmConfig.Model, f.optsFunc.Model)

	return opts
}

// buildOpt adds an option to the list if the value is not empty.
func (f *genericFactory[O, M]) buildOpt(opts *[]O, value string, o func(string) O) {
	if value != "" && f != nil {
		*opts = append(*opts, o(value))
	}
}
