package kubernetes

import (
	"fmt"
	"os"
)

// kubeConfigPath is the path to the default kubeconfig file.
var (
	kubeConfigPath = os.ExpandEnv("$HOME/.kube/config")
)

// CreateTmpKubeConfigFile creates a temporary kubeconfig file by copying the
// default kubeconfig. This is useful for isolating the kubeconfig used by the
// application from the user's default kubeconfig.
//
// TODO: Use the Kubernetes client-go library to get the kubeconfig instead of
// reading from a file.
func CreateTmpKubeConfigFile() (string, error) {
	// TODO: use kubernetes client-go to get the kubeconfig instead of reading from a file.
	kubeConfig, err := os.ReadFile(kubeConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	// Create a temporary file to store the kubeconfig
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary kubeconfig file: %w", err)
	}

	// Write the kubeconfig content to the temporary file
	if _, err := tmpFile.WriteString(string(kubeConfig)); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write kubeconfig to temporary file: %w", err)
	}

	// Close the file and return its name
	tmpFile.Close()
	return tmpFile.Name(), nil
}
