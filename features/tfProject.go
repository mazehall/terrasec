package features_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	Simple             = "simple main.tf"
	TsConfigFileRepo   = "ts config with file repository"
	TsConfigGopassRepo = "ts config with gopass repository"

	srcDir = "fixture"
)

type TfProject struct {
	Path string
}

func (tf *TfProject) Prepare(style string) error {
	files := make(map[string]string)
	switch style {
	case TsConfigFileRepo:
		files["terrasec_file.hcl"] = "terrasec.hcl"
	case TsConfigGopassRepo:
		files["terrasec.hcl"] = "terrasec.hcl"
	case Simple:
		files["main.tf"] = "main.tf"
	}
	for src, file := range files {
		if cpErr := tf.copy(src, file); cpErr != nil {
			return cpErr
		}
	}
	return nil
}

func (tf *TfProject) Cleanup() {
	os.RemoveAll(tf.Path)
}

func (tf *TfProject) IsUrlInStatefile(url string) (bool, error) {
	b, err := ioutil.ReadFile(tf.Path + "/.terraform/terraform.tfstate")
	if err != nil {
		return false, err
	}
	return bytes.Contains(b, []byte(url)), nil
}

func (tf *TfProject) copy(srcFile string, file string) error {
	src := filepath.Join(srcDir, srcFile)
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(filepath.Join(tf.Path, file))
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}
