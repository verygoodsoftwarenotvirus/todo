// +build mage

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"

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
	cwd string
	debug,
	letHang,
	verbose bool
	containerRunner = docker
	logger          logging.Logger

	Aliases = map[string]interface{}{
		"dev":                Run,
		"loud":               Verbose,
		"fmt":                Format,
		"integration-tests":  IntegrationTests,
		"lintegration-tests": LintegrationTests,
	}
	_ = Aliases
)

type Backend mg.Namespace
type Frontend mg.Namespace

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

// bool vars

// Enables debug mode.
func Debug() {
	debug = true
	logger.SetLevel(logging.DebugLevel)
	logger.Debug("debug logger activated")
}

// Enables verbose mode.
func Verbose() {
	verbose = true
	logger.Debug("verbose output activated")
}

// Enables integration test instances to continue running after the tests complete.
func LetHang() {
	letHang = true
	logger.Debug("let hang activated")
}

// helpers

func runFunc(outLoud bool) func(string, ...string) error {
	var runCmd = sh.Run
	if outLoud || verbose {
		runCmd = sh.RunV
	}

	return runCmd
}

func runGoCommand(verbose bool, arguments ...string) error {
	if err := runFunc(verbose)(_go, arguments...); err != nil {
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

func PrintTestPackages() error {
	packages, err := determineTestablePackages()
	if err != nil {
		return err
	}

	for _, x := range packages {
		fmt.Println(x)
	}

	return nil
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

func runContainer(outLoud bool, runSpec containerRunSpec) error {
	containerRunArgs := append([]string{run}, runSpec.runArgs...)
	containerRunArgs = append(containerRunArgs, fmt.Sprintf("%s:%s", runSpec.imageName, runSpec.imageVersion))
	containerRunArgs = append(containerRunArgs, runSpec.imageArgs...)

	var runCmd = sh.Run
	if outLoud {
		runCmd = sh.RunV
	}

	return runCmd(containerRunner, containerRunArgs...)
}

func runCompose(composeFiles ...string) error {
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

// Install mage if necessary.
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

func ensureGoimports() error {
	present, checkErr := pkg.IsCommandAvailable("goimports", "", "")
	if checkErr != nil {
		return checkErr
	}

	if !present {
		return runGoCommand(false, "get", "golang.org/x/tools/cmd/goimports")
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

// Install all auxiliary dev tools.
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

// tool invokers

func fixFieldAlignment() {
	ensureFieldalignment()

	sh.Output("fieldalignment", "-fix", "./...")
}

func runGoimports() error {
	ensureGoimports()

	return runGoCommand(false, "-local", thisRepo)
}

// dependency stuff

// Generate the dependency injected build file.
func Wire() error {
	if err := ensureDependencyInjector(); err != nil {
		return err
	}

	return sh.RunV("wire", "gen", filepath.Join(thisRepo, "cmd", "server"))
}

// Delete existing dependency injected build file and regenerate it.
func Rewire() error {
	if err := os.Remove("cmd/server/wire_gen.go"); err != nil {
		return err
	}

	return Wire()
}

// Set up the Go vendor directory.
func Vendor() error {
	const mod = "mod"

	if _, err := os.ReadFile("go.mod"); os.IsNotExist(err) {
		if initErr := runGoCommand(false, mod, "init"); initErr != nil {
			return initErr
		}

		if tidyErr := runGoCommand(false, mod, "tidy"); tidyErr != nil {
			return tidyErr
		}
	}

	return runGoCommand(true, mod, vendor)
}

// Install frontend dependencies.
func FrontendVendor() error {
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	if err := sh.RunV(frontendTool, _install); err != nil {
		return err
	}

	return os.Chdir(cwd)
}

// Delete existing dependency store and re-establish it for the backend.
func (Backend) Revendor() error {
	if err := os.Remove("go.sum"); err != nil {
		return err
	}

	if err := os.RemoveAll(vendor); err != nil {
		return err
	}

	if err := Vendor(); err != nil {
		return err
	}

	return nil
}

// Delete existing dependency store and re-establish it for the frontend.
func (Frontend) Revendor() error {
	if err := os.RemoveAll("frontend/node_modules"); err != nil {
		return err
	}

	if err := FrontendVendor(); err != nil {
		return err
	}

	return nil
}

// meta stuff

// Produce line count report
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

func formatFrontend(outLoud bool) error {
	rf := sh.Run
	if outLoud {
		rf = sh.RunV
	}

	return runCommandsFromFrontendFolder(func() error {
		return rf(frontendTool, run, "format")
	})
}

// Format the frontend and backend code.
func Format() error {
	if err := formatBackend(); err != nil {
		return err
	}

	if err := formatFrontend(false); err != nil {
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

// Check to see if the frontend and backend are formatted correctly.
func CheckFormatting() error {
	if err := checkBackendFormatting(); err != nil {
		return err
	}

	if err := checkFrontendFormatting(); err != nil {
		return err
	}

	return nil
}

func dockerLint(outLoud bool) error {
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

	return runContainer(outLoud, dockerLintCmd)
}

// Lint the available dockerfiles.
func DockerLint() error {
	return dockerLint(true)
}

// Lint the backend code.
func Lint() error {
	const (
		lintImage        = "golangci/golangci-lint"
		lintImageVersion = "latest"
	)

	fixFieldAlignment()

	if err := dockerLint(verbose); err != nil {
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

	return runContainer(true, lintCmd)
}

func backendCoverage() error {
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

// Coverage generates a coverage report for the backend code.
func Coverage() error {
	return backendCoverage()
}

// Testing

func backendUnitTests(outLoud, quick bool) error {
	packagesToTest, err := determineTestablePackages()
	if err != nil {
		return err
	}

	var commandStartArgs []string
	if quick {
		commandStartArgs = []string{"test", "-cover", "-race", "-failfast"}
	} else {
		commandStartArgs = []string{"test", "-count", "5", "-race"}
	}

	fullCommand := append(commandStartArgs, packagesToTest...)
	if err = runGoCommand(outLoud, fullCommand...); err != nil {
		return err
	}

	return nil
}

// Run backend unit tests
func (Backend) UnitTests() error {
	return backendUnitTests(true, false)
}

// Run unit tests but exit upon first failure.
func Quicktest() error {
	if err := backendUnitTests(true, true); err != nil {
		return err
	}

	if err := frontendUnitTests(true); err != nil {
		return err
	}

	return nil
}

const (
	mariadb  = "mariadb"
	postgres = "postgres"
	sqlite   = "sqlite"
)

// Run a specific integration test.
func IntegrationTest(dbProvider string) error {
	dbProvider = strings.TrimSpace(strings.ToLower(dbProvider))

	if err := validateDBProvider(dbProvider); err != nil {
		return nil
	}

	err := runCompose(
		"environments/testing/compose_files/integration_tests/integration-tests-base.yaml",
		fmt.Sprintf("environments/testing/compose_files/integration_tests/integration-tests-%s.yaml", dbProvider),
	)
	if err != nil {
		return err
	}

	return nil
}

// Run integration tests.
func IntegrationTests() error {
	if err := IntegrationTest(sqlite); err != nil {
		return err
	}
	if err := IntegrationTest(postgres); err != nil {
		return err
	}
	if err := IntegrationTest(mariadb); err != nil {
		return err
	}

	return nil
}

// Run the integration tests and collect coverage information.
func IntegrationCoverage() error {
	if err := freshArtifactsDir(); err != nil {
		return err
	}

	err := runCompose(
		"environments/testing/compose_files/integration_tests/integration-tests-base.yaml",
		"environments/testing/compose_files/integration_tests/integration-coverage.yaml",
	)

	if err != nil {
		return err
	}

	return nil
}

// Run the integration tests and then the linter.
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

	if err := runCompose("environments/testing/compose_files/load_tests/load-tests-base.yaml", fmt.Sprintf("environments/testing/compose_files/load_tests/load-tests-%s.yaml", dbProvider)); err != nil {
		return err
	}

	return nil
}

// Run load tests.
func LoadTests() error {
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

// Run the browser-driven tests.
func BrowserDrivenTests() error {
	if err := runCompose("environments/testing/compose_files/frontend-tests.yaml"); err != nil {
		return err
	}

	return nil
}

// Development

// Generate configuration files.
func Configs() error {
	return runGoCommand(true, run, "cmd/tools/config_gen/main.go")
}

// Run the service in docker-compose.
func Run() error {
	if err := freshArtifactsDir(); err != nil {
		return err
	}

	if err := runCompose("environments/local/docker-compose.yaml"); err != nil {
		return err
	}

	return nil
}

func runCommandsFromFrontendFolder(funcs ...func() error) error {
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}

	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}

	return os.Chdir(cwd)
}

// Populate empty test files for the sake of coverage collection.
func ScaffoldFrontendTests() error {
	err := filepath.Walk("frontend/src",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(info.Name(), ".ts") && !strings.HasSuffix(info.Name(), ".test.ts") && !strings.HasSuffix(info.Name(), ".d.ts") {
				sansExtension := path[:len(path)-3]
				newFileName := sansExtension + ".test.ts"
				if _, statErr := os.Stat(newFileName); os.IsNotExist(statErr) {
					output := fmt.Sprintf(`import './%s'; // TODO: test me!
`, filepath.Base(sansExtension))

					if writeErr := os.WriteFile(newFileName, []byte(output), 0644); writeErr != nil {
						return writeErr
					}
				} else {
					return statErr
				}
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func frontendUnitTests(outLoud bool) error {
	rf := sh.Run
	if outLoud {
		rf = sh.RunV
	}

	return runCommandsFromFrontendFolder(func() error {
		return rf(frontendTool, run, "test")
	})
}

// Unit test the frontend code.
func (Frontend) UnitTests() error {
	return frontendUnitTests(true)
}

// Watch for changes to the frontend files, build, and then serve them.
func (Frontend) AutoBuild() error {
	if err := os.RemoveAll(filepath.Join(frontendDir, "dist", "build")); err != nil {
		return err
	}

	return runCommandsFromFrontendFolder(func() error {
		return sh.RunV(frontendTool, run, "autobuild")
	})
}

// Runs the frontend in frontend-only mode.
func (Frontend) Only() error {
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

// Create test users in a running instance of the service.
func ScaffoldUsers(userCount uint) error {
	fullArgs := []string{
		run,
		filepath.Join(thisRepo, "/cmd/tools/data_scaffolder"),
		fmt.Sprintf("--url=%s", localAddress),
		fmt.Sprintf("--count=%d", userCount),
		"--debug",
	}

	if userCount == 1 {
		fullArgs = append(fullArgs, "--single-user-mode")
	}

	if err := runGoCommand(true, fullArgs...); err != nil {
		return err
	}

	return nil
}

// Create a test user in a running instance of the service.
func ScaffoldUser() error {
	return ScaffoldUsers(1)
}
