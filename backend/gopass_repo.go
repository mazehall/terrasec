package backend

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type GopassHcl struct {
	Repo GopassConfig `hcl:"repository,block"`
}

type GopassConfig struct {
	Kind    string            `hcl:",label"`
	State   string            `hcl:"state"`
	Secrets map[string]string `hcl:"secret,optional"`
}

type GopassRepo struct {
	configFile string
	synced     bool
	config     GopassHcl
}

// part of repository interface
func (r GopassRepo) getState() ([]byte, error) {
	if err := r.prepare(); err != nil {
		return []byte{}, err
	}
	cmd := exec.Command("gopass", "show", "-f", r.config.Repo.State)
	secret, err := cmd.Output()
	if err != nil {
		// state not clearly found -> new state
		return []byte{}, &GetStateError{
			message: err.Error(),
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(string(secret))
	if err != nil {
		return []byte{}, fmt.Errorf("state decoding failed with error: %s\nPlease repair gopass entry: %s", err, r.config.Repo.State)
	}
	return decoded, nil
}

// part of repository interface
func (r GopassRepo) saveState(payload []byte) error {
	if err := r.prepare(); err != nil {
		return err
	}
	cmd := exec.Command("gopass", "insert", "-f", r.config.Repo.State)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if nil != err {
		return fmt.Errorf("obtaining stdin: %s", err.Error())
	}
	payload = append(payload, []byte("\n")...)
	encoder := base64.NewEncoder(base64.StdEncoding, stdin)
	if _, err := encoder.Write(payload); nil != err {
		return fmt.Errorf("encoding failed with %s", err)
	}

	if err := cmd.Start(); nil != err {
		return fmt.Errorf("gopass failed with %s", err)
	}
	stdin.Close()
	encoder.Close()
	if err := cmd.Wait(); nil != err {
		return fmt.Errorf("gopass failed with %s", err)
	}

	return nil
}

// part of repository interface
func (r GopassRepo) DeleteState() error {
	if err := r.prepare(); err != nil {
		return err
	}
	cmd := exec.Command("gopass", "rm", "-f", r.config.Repo.State)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gopass failed with %s", err)
	}
	return nil
}

// part of repository interface
func (r GopassRepo) lockState(payload []byte) error { return nil }

// part of repository interface
func (r GopassRepo) unlockState(payload []byte) error { return nil }

func (r *GopassRepo) prepare() error {
	if err := r.setConfig(); err != nil {
		return err
	}
	if err := r.gopassSync(); err != nil {
		return err
	}
	return nil
}

func (r *GopassRepo) setConfig() error {
	if r.config.Repo.Kind != "" {
		return nil
	}

	if err := hclsimple.DecodeFile(r.configFile, nil, &r.config); err != nil {
		return fmt.Errorf("failed to load configuration: %s", err)
	}
	// fmt.Printf("Gopass Configuration is %#v", r.config)
	return nil
}

func (r *GopassRepo) gopassSync() error {
	if r.synced {
		return nil
	}
	cmd := exec.Command("gopass", "sync")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to sync with remote: check your gopass state\n%s", err)
	}
	r.synced = true
	return nil
}
