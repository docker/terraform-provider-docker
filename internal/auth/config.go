/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/go-jose/go-jose/v3/jwt"
)

// ConfigStore wraps the  config file and provides credential access methods
type ConfigStore struct {
	configFile *configfile.ConfigFile
}

// NewConfigStore creates a new ConfigStore with the default  config file
func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		configFile: config.LoadDefaultConfigFile(os.Stderr),
	}
}

// GetCredentialStorePullTokens retrieves the user credentials the the user is
// using to pull. This may be a username/password, a PAT, or an OAT.
func (c *ConfigStore) GetCredentialStorePullTokens(registryEntry string) (string, string, error) {
	return c.readUserCreds(registryEntry)
}

// GetCredentialStoreAccessTokens retrieves the JWT that Desktop uses for API
// sessions.
func (c *ConfigStore) GetCredentialStoreAccessTokens(registryEntry string) (string, string, error) {
	accessTokenKey := c.getAccessTokenConfigKey(registryEntry)
	username, accessToken, err := c.readUserCreds(accessTokenKey)
	if err == nil && accessToken != "" {
		// Check if the accessToken is a valid JWT and not expired
		if isJWTAcceptable(accessToken) {
			return username, accessToken, nil
		}
	}

	return "", "", fmt.Errorf("no valid JWT found for %s", registryEntry)
}

// readUserCreds reads user credentials from the config file for a given
// registry entry
func (c *ConfigStore) readUserCreds(registryEntry string) (string, string, error) {
	authConfig, err := c.configFile.GetAuthConfig(registryEntry)
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

// getAccessTokenConfigKey generates the config key for access tokens
// (used by Desktop)
func (c *ConfigStore) getAccessTokenConfigKey(configKey string) string {
	result := configKey
	if !strings.HasSuffix(result, "/") {
		result = result + "/"
	}
	return result + "access-token"
}

// Check if we think the JWT is still usable
// and worth sending. Not used for security, just used
// as a simple heuristic of which token is better.
func isJWTAcceptable(token string) bool {
	claims, err := getClaims(token)
	if err != nil {
		return false
	}
	if claims.Expiry == nil {
		return false
	}
	expiry := claims.Expiry.Time()
	return time.Now().Before(expiry)
}

// getClaims returns claims from an access token without verification.
func getClaims(accessToken string) (*jwt.Claims, error) {
	token, err := jwt.ParseSigned(accessToken)
	if err != nil {
		return nil, err
	}

	var claims jwt.Claims
	err = token.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
