package repositoryutils

import "fmt"

func NewID(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
