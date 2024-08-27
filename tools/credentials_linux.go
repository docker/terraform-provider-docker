// go:build linux

package tools

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/secretservice"
)

func ReadCredentialsFromStore(serverAddress string) (credentials.Credentials, error) {
	helper := secretservice.Secretservice{}
	username, password, err := helper.Get(serverAddress)
	if err != nil {
		return credentials.Credentials{}, err
	}

	return credentials.Credentials{
		Username: username,
		Secret:   password,
	}, nil
}
