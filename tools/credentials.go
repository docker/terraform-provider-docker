package tools

import (
	"fmt"
	"os"

	"github.com/docker/cli/cli/config"
)

// getUserCreds retrieves user credentials from the Docker config file.
func GetUserCreds(registryEntry string) (string, string, error) {
	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	authConfig, err := dockerConfig.GetAuthConfig(registryEntry)
	if err != nil {
		return "", "", fmt.Errorf("get auth config: %w", err)
	}
	username := authConfig.Username
	secret := authConfig.Password
	if authConfig.IdentityToken != "" {
		secret = authConfig.IdentityToken
	}

	return username, secret, nil
}
