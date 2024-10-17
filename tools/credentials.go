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
