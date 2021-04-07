// +build mage

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/carolynvs/magex/pkg"
	"github.com/magefile/mage/sh"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

const (
	// common terms and tools
	_go      = "go"
	npm      = "npm"
	docker   = "docker"
	vendor   = "vendor"
	_install = "install"
	run      = "run"

	frontendTool = "pnpm"
	artifactsDir = "artifacts"
	frontendDir  = "frontend"

	thisRepo     = "gitlab.com/verygoodsoftwarenotvirus/todo"
	localAddress = "http://localhost:8888"
)

var (
	cwd             string
	debug           bool
	verbose         bool
	containerRunner = docker
	logger          logging.Logger

	Aliases = map[string]interface{}{
		"dev":               Run,
		"loud":              Verbose,
		"fmt":               Format,
		"integration-tests": IntegrationTests,
	}
	_ = Aliases
)

type containerRunSpec struct {
	imageName,
	imageVersion string
	imageArgs []string
	runArgs   []string
}

func init() {
	logger = zerolog.NewLogger()

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	var err error
	if cwd, err = os.Getwd(); err != nil {
		logger.Error(err, "determining current working directory")
		panic(err)
	}

	logger = logger.WithValue("current_working_directory", cwd)
	logger.Debug("cwd determined")

	if !strings.HasSuffix(cwd, thisRepo) {
		panic("location invalid!")
	}
}

func Debug() {
	debug = true
	logger.SetLevel(logging.DebugLevel)
	logger.Debug("debug logger activated")
}

func Verbose() {
	verbose = true
	logger.Debug("verbose output activated")
}

// helpers

func clear() {
	const controlSequenceIntroducer = "\033["
	const clearSequence = "" +
		controlSequenceIntroducer + "2J" + // erase entire display
		controlSequenceIntroducer + "3J" + // erase entire display including scroll-back buffer
		controlSequenceIntroducer + "[1;1H" // reset cursor position to top of display

	fmt.Printf(clearSequence)
}

func runGoCommand(verbose bool, arguments ...string) error {
	var runCmd = sh.Run
	if verbose {
		runCmd = sh.RunV
	}

	if err := runCmd(_go, arguments...); err != nil {
		return err
	}

	return nil
}

func freshArtifactsDir() error {
	if err := os.RemoveAll(filepath.Join(cwd, artifactsDir)); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cwd, artifactsDir), fs.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cwd, artifactsDir, "search_indices"), fs.ModePerm); err != nil {
		return err
	}

	return nil
}

func validateDBProvider(dbProvider string) error {
	switch strings.TrimSpace(strings.ToLower(dbProvider)) {
	case sqlite, mariadb, postgres:
		return nil
	default:
		return fmt.Errorf("invalid database provider: %q", dbProvider)
	}
}

func doesNotMatch(input string, matcher func(string, string) bool, exclusions ...string) bool {
	included := true

	for _, exclusion := range exclusions {
		if !included {
			break
		}
		included = !matcher(input, exclusion)
	}

	return included
}

func doesNotStartWith(input string, exclusions ...string) bool {
	return doesNotMatch(input, strings.HasPrefix, exclusions...)
}

func doesNotEndWith(input string, exclusions ...string) bool {
	return doesNotMatch(input, strings.HasSuffix, exclusions...)
}

func determineTestablePackages() ([]string, error) {
	var out []string

	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			included := doesNotStartWith(
				path,
				".",
				".git",
				".idea",
				"cmd",
				artifactsDir,
				"development",
				"environments",
				frontendDir,
				"tests",
				vendor,
			) && doesNotEndWith(path, "mock", "testutil", "fakes")

			if info.IsDir() && included {
				entries, err := fs.ReadDir(os.DirFS(path), ".")
				if err != nil {
					return err
				}

				var goFilesPresent bool
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".go") {
						goFilesPresent = true
					}
				}

				if goFilesPresent {
					out = append(out, filepath.Join(thisRepo, path))
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func runContainer(runSpec containerRunSpec) error {
	containerRunArgs := append([]string{run}, runSpec.runArgs...)
	containerRunArgs = append(containerRunArgs, fmt.Sprintf("%s:%s", runSpec.imageName, runSpec.imageVersion))
	containerRunArgs = append(containerRunArgs, runSpec.imageArgs...)

	return sh.RunV(containerRunner, containerRunArgs...)
}

func runCompose(letHang bool, composeFiles ...string) error {
	fullCommand := []string{}
	for _, f := range composeFiles {
		if f == "" {
			return errors.New("empty filepath provided to docker-compose")
		}
		fullCommand = append(fullCommand, "--file", f)
	}

	fullCommand = append(fullCommand,
		"up",
		"--build",
		"--force-recreate",
		"--remove-orphans",
		"--renew-anon-volumes",
		"--always-recreate-deps",
	)

	if !letHang {
		fullCommand = append(fullCommand, "--abort-on-container-exit")
	}

	return sh.RunV("docker-compose", fullCommand...)
}

// tool ensurers

// Install Mage if necessary
func EnsureMage() error {
	return pkg.EnsureMage("v1.11.0")
}

func ensureDependencyInjector() error {
	present, checkErr := pkg.IsCommandAvailable("wire", "", "")
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return runGoCommand(false, _install, "github.com/google/wire/cmd/wire")
	}

	return nil
}

func ensureFieldalignment() error {
	present, checkErr := pkg.IsCommandAvailable("fieldalignment", "", "")
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return runGoCommand(false, _install, "golang.org/x/tools/...")
	}

	return nil
}

func ensureLineCounter() error {
	present, checkErr := pkg.IsCommandAvailable("scc", "3.0.0", "--version")
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return runGoCommand(false, _install, "github.com/boyter/scc")
	}

	return nil
}

func ensureFrontendInstaller() error {
	present, checkErr := pkg.IsCommandAvailable("pnpm", "5.18.8", "--version")
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return sh.Run(npm, _install, "--global", "pnpm")
	}

	return nil
}

func checkForDocker() error {
	present, checkErr := pkg.IsCommandAvailable(docker, "20.10.5", `--format="{{.Client.Version}}"`)
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return fmt.Errorf("%s is not installed", docker)
	}

	return nil
}

func EnsureDevTools() error {
	if err := ensureDependencyInjector(); err != nil {
		return err
	}

	if err := ensureFieldalignment(); err != nil {
		return err
	}

	if err := ensureLineCounter(); err != nil {
		return err
	}

	if err := ensureFrontendInstaller(); err != nil {
		return err
	}

	return nil
}

// end tool ensurers

// dependency stuff

func Wire() error {
	if err := ensureDependencyInjector(); err != nil {
		return err
	}

	return sh.RunV("wire", "gen", filepath.Join(thisRepo, "cmd", "server"))
}

func Rewire() error {
	if err := os.Remove("cmd/server/wire_gen.go"); err != nil {
		return err
	}

	return Wire()
}

func Vendor() error {
	if _, err := os.ReadFile("go.mod"); os.IsNotExist(err) {
		if initErr := runGoCommand(false, "mod", "init"); initErr != nil {
			return initErr
		}

		if tidyErr := runGoCommand(false, "mod", "tidy"); tidyErr != nil {
			return tidyErr
		}
	}

	return runGoCommand(true, "mod", vendor)
}

func FrontendVendor() error {
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, _install); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

func Revendor() error {
	if err := os.Remove("go.sum"); err != nil {
		return err
	}

	if err := os.RemoveAll(vendor); err != nil {
		return err
	}

	if err := os.RemoveAll("frontend/node_modules"); err != nil {
		return err
	}

	if err := Vendor(); err != nil {
		return err
	}

	if err := FrontendVendor(); err != nil {
		return err
	}

	return nil
}

// meta stuff

func LineCount() error {
	logger.Debug("lineCount called")
	if err := ensureLineCounter(); err != nil {
		logger.Debug("error ensuring line counter")
		return err
	}

	if err := sh.RunV(
		"scc", "",
		"--include-ext", _go,
		"--exclude-dir", vendor,
		"--exclude-dir", "frontend/node_modules"); err != nil {
		logger.Debug("error fetching line count")
		return err
	}

	logger.Debug("fetched line count")
	return nil
}

// Quality

func formatBackend() error {
	var goFiles []string

	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(info.Name(), ".go") {
				goFiles = append(goFiles, path)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return sh.Run("gofmt", append([]string{"-s", "-w"}, goFiles...)...)
}

func formatFrontend() error {
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, run, "format"); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

func Format() error {
	if err := formatBackend(); err != nil {
		return err
	}

	if err := formatFrontend(); err != nil {
		return err
	}

	return nil
}

func checkBackendFormatting() error {
	badFiles, err := sh.Output("gofmt", "-l", ".")
	if err != nil {
		return err
	}

	if len(badFiles) > 0 {
		return errors.New(badFiles)
	}

	return nil
}

func checkFrontendFormatting() error {
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, run, "format:check"); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

func CheckFormatting() error {
	if err := checkBackendFormatting(); err != nil {
		return err
	}

	if err := checkFrontendFormatting(); err != nil {
		return err
	}

	return nil
}

func DockerLint() error {
	const (
		dockerLintImage        = "openpolicyagent/conftest"
		dockerLintImageVersion = "v0.21.0"
	)

	var dockerfiles []string

	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(info.Name(), ".Dockerfile") {
				dockerfiles = append(dockerfiles, path)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	dockerLintCmd := containerRunSpec{
		runArgs: []string{
			"--rm",
			"--volume",
			fmt.Sprintf("%s:%s", cwd, cwd),
			fmt.Sprintf("--workdir=%s", cwd),
		},
		imageName:    dockerLintImage,
		imageVersion: dockerLintImageVersion,
		imageArgs: append([]string{
			"test",
			"--policy",
			"docker_security.rego",
		}, dockerfiles...),
	}

	return runContainer(dockerLintCmd)
}

func Lint() error {
	const (
		lintImage        = "golangci/golangci-lint"
		lintImageVersion = "latest"
	)

	if err := DockerLint(); err != nil {
		return err
	}

	if err := sh.Run(containerRunner, "pull", lintImage); err != nil {
		return err
	}

	lintCmd := containerRunSpec{
		runArgs: []string{
			"--rm",
			"--volume",
			fmt.Sprintf("%s:%s", cwd, cwd),
			fmt.Sprintf("--workdir=%s", cwd),
		},
		imageName:    lintImage,
		imageVersion: lintImageVersion,
		imageArgs: []string{
			"golangci-lint",
			run,
			"--config=.golangci.yml",
			"./...",
		},
	}

	return runContainer(lintCmd)
}

func Coverage() error {
	if err := freshArtifactsDir(); err != nil {
		return err
	}

	coverageFileOutputPath := filepath.Join(artifactsDir, "coverage.out")

	packagesToTest, err := determineTestablePackages()
	if err != nil {
		return err
	}

	testCommand := append([]string{
		"test",
		fmt.Sprintf("-coverprofile=%s", coverageFileOutputPath),
		"-covermode=atomic",
		"-race",
	}, packagesToTest...)

	if err = runGoCommand(false, testCommand...); err != nil {
		return err
	}

	coverCommand := []string{
		"tool",
		"cover",
		fmt.Sprintf("-func=%s/coverage.out", artifactsDir),
	}

	results, err := sh.Output(_go, coverCommand...)
	if err != nil {
		return err
	}

	// byte array jesus please forgive me
	rawCoveragePercentage := string([]byte(results)[len(results)-6 : len(results)])

	fmt.Println(strings.TrimSpace(rawCoveragePercentage))

	return nil
}

// Testing

func Quicktest() error {
	packagesToTest, err := determineTestablePackages()
	if err != nil {
		return err
	}

	fullCommand := append([]string{"test", "-cover", "-race", "-failfast"}, packagesToTest...)
	if err = runGoCommand(true, fullCommand...); err != nil {
		return err
	}

	return nil
}

const (
	mariadb  = "mariadb"
	postgres = "postgres"
	sqlite   = "sqlite"
)

func runIntegrationTest(dbProvider string) error {
	dbProvider = strings.TrimSpace(strings.ToLower(dbProvider))

	if err := validateDBProvider(dbProvider); err != nil {
		return nil
	}

	if err := runCompose(
		false,
		"environments/testing/compose_files/integration_tests/integration-tests-base.yaml",
		fmt.Sprintf("environments/testing/compose_files/integration_tests/integration-tests-%s.yaml", dbProvider),
	); err != nil {
		return err
	}

	return nil
}

func IntegrationTests() error {
	if err := runIntegrationTest(sqlite); err != nil {
		return err
	}
	if err := runIntegrationTest(postgres); err != nil {
		return err
	}
	if err := runIntegrationTest(mariadb); err != nil {
		return err
	}

	return nil
}

func IntegrationCoverage() error {
	if err := freshArtifactsDir(); err != nil {
		return err
	}

	err := runCompose(
		false,
		"environments/testing/compose_files/integration_tests/integration-tests-base.yaml",
		"environments/testing/compose_files/integration_tests/integration-coverage.yaml",
	)

	if err != nil {
		return err
	}

	return nil
}

func LintegrationTests() error {
	if err := IntegrationTests(); err != nil {
		return err
	}

	if err := Lint(); err != nil {
		return err
	}

	return nil
}

func runLoadTest(dbProvider string) error {
	dbProvider = strings.TrimSpace(strings.ToLower(dbProvider))

	if err := validateDBProvider(dbProvider); err != nil {
		return nil
	}

	if err := runCompose(
		false,
		"environments/testing/compose_files/load_tests/load-tests-base.yaml",
		fmt.Sprintf("environments/testing/compose_files/load_tests/load-tests-%s.yaml", dbProvider),
	); err != nil {
		return err
	}

	return nil
}

func Loadtests() error {
	if err := runLoadTest(sqlite); err != nil {
		return err
	}

	if err := runLoadTest(postgres); err != nil {
		return err
	}

	if err := runLoadTest(mariadb); err != nil {
		return err
	}

	return nil
}

func BrowserDrivenTests() error {
	if err := runCompose(false, "environments/testing/compose_files/frontend-tests.yaml"); err != nil {
		return err
	}

	return nil
}

// Development

func Configs() error {
	return runGoCommand(true, run, "cmd/tools/config_gen/main.go")
}

func Run() error {
	if err := freshArtifactsDir(); err != nil {
		return err
	}

	if err := runCompose(false, "environments/local/docker-compose.yaml"); err != nil {
		return err
	}

	return nil
}

func FrontendAutobuild() error {
	if err := os.RemoveAll(filepath.Join(frontendDir, "dist", "build")); err != nil {
		return err
	}

	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, run, "autobuild"); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

func FrontendOnly() error {
	if err := os.RemoveAll(filepath.Join(frontendDir, "dist", "build")); err != nil {
		return err
	}

	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, run, "start:frontend-only"); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

func ScaffoldUsers(userCount uint) error {
	fullArgs := []string{
		run,
		filepath.Join(cwd, "/cmd/tools/data_scaffolder"),
		fmt.Sprintf("--url=%s", localAddress),
		fmt.Sprintf("--count=%d", userCount),
		"--debug",
	}

	if userCount == 1 {
		fullArgs = append(fullArgs, "--single-user-mode")
	}

	if err := runGoCommand(false, fullArgs...); err != nil {
		return err
	}

	return nil
}
