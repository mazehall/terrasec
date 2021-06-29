package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type fileHcl struct {
	Repo fileConfig `hcl:"repository,block"`
}

type fileConfig struct {
	Kind  string `hcl:",label"`
	State string `hcl:"state"`
}

type FileRepo struct {
	configFile string
	config     fileHcl
}

// part of repository interface
func (r FileRepo) getState() ([]byte, error) {
	if err := r.prepare(); err != nil {
		return []byte{}, err
	}
	secret, err := os.ReadFile(filepath.Join(filepath.Dir(r.configFile), r.config.Repo.State))
	if err != nil {
		return []byte{}, err
	}
	// fmt.Println(string(secret))
	return secret, nil
}

// part of repository interface
func (r FileRepo) saveState(payload []byte) error {
	if err := r.prepare(); err != nil {
		return err
	}
	if err := os.WriteFile(
		filepath.Join(filepath.Dir(r.configFile), r.config.Repo.State),
		payload,
		0644); err != nil {
		return err
	}
	return nil
}

// part of repository interface
func (r FileRepo) deleteState() error { return nil }

// part of repository interface
func (r FileRepo) lockState(payload []byte) error { return nil }

// part of repository interface
func (r FileRepo) unlockState(payload []byte) error { return nil }

func (r *FileRepo) prepare() error {
	if err := r.setConfig(); err != nil {
		return err
	}
	return nil
}

func (r *FileRepo) setConfig() error {
	if r.config.Repo.Kind != "" {
		return nil
	}
	if err := hclsimple.DecodeFile(r.configFile, nil, &r.config); err != nil {
		return fmt.Errorf("failed to load configuration: %s", err)
	}
	// fmt.Printf("Gopass Configuration is %#v", r.config)
	return nil
}
