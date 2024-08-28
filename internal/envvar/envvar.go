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
