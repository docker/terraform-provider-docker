// go:build windows

package tools

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/wincred"
)

func ReadCredentialsFromStore(serverAddress string) (credentials.Credentials, error) {
	helper := wincred.Wincred{}
	username, password, err := helper.Get(serverAddress)
	if err != nil {
		return credentials.Credentials{}, err
	}

	return credentials.Credentials{
		Username: username,
		Secret:   password,
	}, nil
}
