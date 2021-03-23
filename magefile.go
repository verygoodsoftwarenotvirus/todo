// +build mage

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carolynvs/magex/pkg"
	"github.com/magefile/mage/sh"
)

const (
	Go       = "go"
	thisRepo = "gitlab.com/verygoodsoftwarenotvirus/todo"
)

// Install Mage if necessary
func EnsureMage() error {
	return pkg.EnsureMage("")
}

var (
	cwd             string
	containerRunner = "docker"
)

func init() {
	podmanInstalled, err := pkg.IsCommandAvailable("podman", "3.0.1", "--version")
	if err != nil {
		panic(err)
	}

	if podmanInstalled {
		containerRunner = "podman"
	}

	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func pipe(fromCmd string, fromArgs []string, toCmd string, toArgs []string) (string, error) {
	from := exec.Command(fromCmd, fromArgs...)
	to := exec.Command(toCmd, toArgs...)

	r, w := io.Pipe()
	from.Stdout = w
	to.Stdin = r

	var b2 bytes.Buffer
	to.Stdout = &b2

	if err := from.Start(); err != nil {
		return "", err
	}

	if err := to.Start(); err != nil {
		return "", err
	}

	if err := from.Wait(); err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	if err := to.Wait(); err != nil {
		return "", err
	}

	if _, err := io.Copy(os.Stdout, &b2); err != nil {
		return "", err
	}

	out, err := to.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
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
				"artifacts",
				"development",
				"environments",
				"frontend",
				"tests",
				"vendor",
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

func Quicktest() error {
	packagesToTest, err := determineTestablePackages()
	if err != nil {
		return err
	}

	fullCommand := append([]string{"test", "-cover", "-race", "-failfast"}, packagesToTest...)
	if err = sh.Run(Go, fullCommand...); err != nil {
		return err
	}

	return nil
}

func runContainer(containerArgs []string, imageName string, imageArgs ...string) error {
	containerRunArgs := append([]string{"run"}, containerArgs...)
	containerRunArgs = append(containerRunArgs, imageName)
	containerRunArgs = append(containerRunArgs, imageArgs...)

	return sh.Run(containerRunner, containerRunArgs...)
}

const lintImage = "golangci/golangci-lint:latest"

func Lint() error {
	if err := sh.Run(containerRunner, "pull", lintImage); err != nil {
		return err
	}

	if err := runContainer(
		[]string{
			"--rm",
			"--volume",
			fmt.Sprintf("%s:%s", cwd, cwd),
			fmt.Sprintf("--workdir=%s", cwd),
		},
		lintImage,
		"golangci-lint",
		"run",
		"--config=.golangci.yml",
		"./...",
	); err != nil {
		return err
	}

	//if err := sh.Run(
	//	containerRunner,
	//	"run",
	//	"--rm",
	//	"--volume",
	//	fmt.Sprintf("%s:%s", cwd, cwd),
	//	fmt.Sprintf("--workdir=%s", cwd),
	//	"--env=GO111MODULE=on",
	//	lintImage,
	//	"golangci-lint",
	//	"run",
	//	"--config=.golangci.yml",
	//	"./...",
	//); err != nil {
	//	return err
	//}

	return nil
}
