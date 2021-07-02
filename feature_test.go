package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	feature "github.com/mazehall/terrasec/features"
)

var (
	command string = "tmp/terrasec"
)

type scenario struct {
	name        string
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	args        []string
	output      []byte
	errorOutput []byte
	update      chan bool
	tf          feature.TfProject
}

func (s *scenario) thereIsATerminal() error {
	_, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	return nil
}

func (s *scenario) iMakeATerrasecCallWith(arg1 string) error {
	// https://github.com/chickenzord/empatpuluh/blob/master/main.go
	// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
	s.args = append([]string{}, "--chdir", s.tf.Path)
	s.args = append(s.args, strings.Split(arg1, " ")...)
	s.cmd = exec.Command(command, s.args...)
	s.update = make(chan bool, 1)
	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err = s.cmd.Start(); err != nil {
		return err
	}

	go func(stdout io.Reader) {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)

		for scanner.Scan() {
			bytes := scanner.Bytes()
			s.output = append(s.output, bytes...)
			s.update <- true
		}
		close(s.update)
		// fmt.Printf("Finished Scan -> %s", s.name)
	}(stdout)
	go func(stderr io.Reader) {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)

		for scanner.Scan() {
			bytes := scanner.Bytes()
			s.errorOutput = append(s.errorOutput, bytes...)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
		// fmt.Printf("Finished Scan -> %s", s.name)
	}(stderr)
	return nil
}

func (s *scenario) iShouldGetOutputWithPattern(arg1 string) error {
	for range s.update {
		// fmt.Print(string(s.output))
	}
	matched, err := regexp.Match(arg1, s.output)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("Pattern\n %s\n not found in\n %s", arg1, string(s.output))
	}
	return nil
}

func (s *scenario) iShouldGetErrorOutputWithPattern(arg1 string) error {
	for range s.update {
	}
	matched, err := regexp.Match(arg1, s.errorOutput)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("Pattern\n %s\n not found in\n %s", arg1, string(s.errorOutput))
	}
	return nil
}

func (s *scenario) thereIsANewTerraformProject() error {
	if err := s.tf.Prepare(feature.Simple); err != nil {
		return err
	}
	return nil
}

func (s *scenario) thereIsAnExistingTerrasecProject() error {
	if err := s.tf.Prepare(feature.Simple); err != nil {
		return err
	}
	if err := s.tf.Prepare(feature.TsConfigFileRepo); err != nil {
		return err
	}
	cmd := exec.Command(command, "--chdir", s.tf.Path, "init")
	if err := cmd.Run(); err != nil {
		return err
	}
	s.tf.Kind = "file"
	return nil
}

func (s *scenario) theSavedStateIsBrokenInTermsOfContent() error {
	if err := s.tf.Prepare(feature.FailState); err != nil {
		return err
	}
	return nil
}

func (s *scenario) thereIsATerrasecConfigWithRepository(arg1 string) error {
	var repoType string
	switch arg1 {
	case "file":
		repoType = feature.TsConfigFileRepo
	case "gopass":
		repoType = feature.TsConfigGopassRepo
	}
	if err := s.tf.Prepare(repoType); err != nil {
		return err
	}
	return nil
}

func (s *scenario) terrasecShouldStartAnHttpServer() error {
	for range s.update {
		if bytes.Contains(s.output, []byte("It can be reached at: 127.0.0.1:")) {
			return nil
		}
	}
	return fmt.Errorf("Server start message not found in: %s", string(s.output))
}

func (s *scenario) theCommandShouldRunProperly() error {
	if s.cmd != nil {
		for range s.update {
		}
		if err := s.cmd.Wait(); nil != err {
			return fmt.Errorf("Process exit code was: %d\n%s", s.cmd.ProcessState.ExitCode(), string(s.output)+string(s.errorOutput))
		}
		// fmt.Printf("Output %v", string(s.output))
		if s.cmd.ProcessState.ExitCode() == 0 {
			return nil
		}
		return fmt.Errorf("Process exit code was: %d\n%s", s.cmd.ProcessState.ExitCode(), string(s.output))
	}
	return nil
}

func (s *scenario) theCommandShouldExitWithError() error {
	if s.cmd != nil {
		for range s.update {
		}
		if err := s.cmd.Wait(); nil != err {
			return nil
		}
		// fmt.Printf("Output %v", string(s.output))
		if s.cmd.ProcessState.ExitCode() == 0 {
			return fmt.Errorf("Process exited without error\n%s", string(s.output))
		}
	}
	return fmt.Errorf("No process was started\n%s", string(s.output))
}

func (s *scenario) atTheEndTheServerShouldBeStopped() error {
	re := regexp.MustCompile(`It can be reached at: (?P<url>[0-9:\.]+)`)
	matches := re.FindSubmatch(s.output)
	url := string(matches[re.SubexpIndex("url")])

	if _, e := http.Get("http://" + url); e != nil {
		return nil
	}
	return fmt.Errorf("server seems to be still running at %s", url)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			fmt.Println("Can't build tmp directory on this machine")
			os.Exit(1)
		}
		if runtime.GOOS == "windows" {
			command = fmt.Sprintf("%s.exe", command)
		}
		cmd := exec.Command("go", "build", "-o", command)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Can't build %s on this machine\n", command)
			os.Exit(1)
		}
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := &scenario{}
	dir, err := os.MkdirTemp("tmp", "*")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	s.tf = feature.TfProject{Path: dir}
	ctx.BeforeScenario(func(sc *messages.Pickle) {
		s.name = sc.Name
	})
	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		s.tf.Cleanup()
	})
	ctx.Step(`^there is a terminal$`, s.thereIsATerminal)
	ctx.Step(`^I make a terrasec call with "([^"]*)"$`, s.iMakeATerrasecCallWith)
	ctx.Step(`^I should get output with pattern "([^"]*)"$`, s.iShouldGetOutputWithPattern)
	ctx.Step(`^there is a new terraform project$`, s.thereIsANewTerraformProject)
	ctx.Step(`^there is a terrasec config with "([^"]*)" repository$`, s.thereIsATerrasecConfigWithRepository)
	ctx.Step(`^the command should run properly$`, s.theCommandShouldRunProperly)
	ctx.Step(`^terrasec should start an http server$`, s.terrasecShouldStartAnHttpServer)
	ctx.Step(`^at the end the server should be stopped$`, s.atTheEndTheServerShouldBeStopped)
	ctx.Step(`^there is an existing terrasec project$`, s.thereIsAnExistingTerrasecProject)
	ctx.Step(`^I should get error output with pattern "([^"]*)"$`, s.iShouldGetErrorOutputWithPattern)
	ctx.Step(`^the command should exit with error$`, s.theCommandShouldExitWithError)
	ctx.Step(`^the saved state is broken in terms of content$`, s.theSavedStateIsBrokenInTermsOfContent)
}
