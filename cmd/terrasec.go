package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mazehall/terrasec/backend"
	"github.com/urfave/cli/v2"
)

const (
	version     = "v1.0.0"
	description = `Secure your secrets and the terraform state 
through encryption with your own keys and even within teams.`
)

var (
	terraform string = "terraform"
)

func Execute() {
	ts := &terrasec{url: "http://127.0.0.1:30303"}
	app := &cli.App{
		Name:        "terrasec",
		Usage:       "A terraform wrapper secured by gopass",
		Version:     version,
		Description: description,
		Commands: []*cli.Command{
			{
				Name:            "init",
				Usage:           "Prepare the http backend and call the terraform init command",
				Action:          ts.InitWithBackend,
				SkipFlagParsing: true,
			},
			{
				Name:            "http-backend",
				Usage:           "Only starts the http backend",
				Action:          ts.StartHttpOnly,
				SkipFlagParsing: false,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "chdir",
				Value: ".",
				Usage: "Placeholder for the corresponding option of terraform",
			}},
		Before: ts.beforeAll,
		Action: ts.CreateTerrasecCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type terrasec struct {
	url string
}
type RepoHcl struct {
	Repo ConfigHcl `hcl:"repository,block"`
}
type ConfigHcl struct {
	Kind    string   `hcl:",label"`
	Options hcl.Body `hcl:",remain"`
}

func (ts *terrasec) beforeAll(c *cli.Context) error {
	if runtime.GOOS == "windows" {
		terraform = terraform + ".exe"
	}
	if _, err := exec.LookPath(terraform); err != nil {
		return errors.New("please install terraform first")
	}

	var config RepoHcl
	dir := c.String("chdir")
	configFile := dir + "/terrasec.hcl"
	if err := hclsimple.DecodeFile(configFile, nil, &config); err != nil {
		return fmt.Errorf("failed to read configuration - file %s", configFile)
	}

	// server start
	port := detectPort(dir)
	url, err := backend.Http{
		Port:       port,
		ConfigFile: configFile,
		Kind:       config.Repo.Kind,
	}.Run()
	if err != nil {
		return err
	}
	ts.url = url

	fmt.Printf("I've just started the backend. It can be reached at: %s\n", ts.url)
	return nil
}

func runTerraform(arguments []string) error {
	fmt.Printf("generated command: terraform %s\n", strings.Join(arguments[:], " "))
	cmd := exec.Command(terraform, arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Fatalf("terrasec failed with %s\n", err)
	}
	return nil
}

func (ts *terrasec) CreateTerrasecCommand(c *cli.Context) error {
	arguments := append([]string{fmt.Sprintf("-chdir=%s", c.String("chdir"))}, c.Args().Slice()...)
	return runTerraform(arguments)
}

func (ts *terrasec) InitWithBackend(c *cli.Context) error {
	url := "http://" + ts.url + "/tfstate"
	arguments := append([]string{
		fmt.Sprintf("-chdir=%s", c.String("chdir")),
		"init",
		"--backend-config",
		"address=" + url,
		"--backend-config",
		"lock_address=" + url,
		"--backend-config",
		"unlock_address=" + url,
	}, c.Args().Slice()...)
	return runTerraform(arguments)
}

func (ts *terrasec) StartHttpOnly(c *cli.Context) error {
	WaitForCtrlC()
	return nil
}

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

func detectPort(dir string) string {
	// read file
	data, err := ioutil.ReadFile(filepath.Join(dir, ".terraform", "terraform.tfstate"))
	if err != nil {
		return ":0"
	}
	var stateBackend TfStateLocal

	if err = json.Unmarshal(data, &stateBackend); err != nil {
		fmt.Println("error:", err)
		return ":0"
	}
	u, err := url.Parse(stateBackend.Backend.Config.Address)
	if err != nil {
		return ":0"
	}
	return ":" + u.Port()
}

type TfStateLocal struct {
	Backend struct {
		Config struct {
			Address       string `json:"address"`
			LockAddress   string `json:"lock_address"`
			UnlockAddress string `json:"unlock_address"`
		} `json:"config"`
	} `json:"backend"`
}
