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

// Package envvar contains constants and defaults for various environment
// variables for the provider.
package envvar

import "os"

// Acceptance test related environment variables
const (
	AccTestOrganization = "ACCTEST_DOCKER_ORG"
)

// defaults is a map of pre-configured defaults for each envvar
var defaults = map[string]string{
	AccTestOrganization: "dockerterraform",
}

// GetWithDefault returns the value of the environment variable or the
// pre-configured default if empty.
func GetWithDefault(key string) string {
	ret := os.Getenv(key)
	if len(ret) > 0 {
		return ret
	}
	return defaults[key]
}
