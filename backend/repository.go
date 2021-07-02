package backend

import (
	"fmt"
)

type repository interface {
	getState() ([]byte, error)
	saveState(payload []byte) error
	DeleteState() error
	lockState(payload []byte) error
	unlockState(payload []byte) error
}

func GetRepo(configFile string, kind string) (repository, error) {
	switch c := kind; {
	case c == "gopass":
		return GopassRepo{configFile: configFile}, nil
	case c == "file":
		return FileRepo{configFile: configFile}, nil
	default:
		return nil, fmt.Errorf("wrong repository configuration: unknown provider \"%s\"", c)
	}
}

type GetStateError struct {
	message string
}

func (e *GetStateError) Error() string {
	return e.message
}
