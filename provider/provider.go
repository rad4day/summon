package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultPath returns the default path where providers are located
var DefaultPath = getDefaultPath()

// Resolve resolves a filepath to a provider
// Checks the CLI arg, environment and then default path
func Resolve(providerArg string) (string, error) {
	provider := providerArg

	if provider == "" {
		provider = os.Getenv("SUMMON_PROVIDER")
	}

	if provider == "" {
		providers, _ := ioutil.ReadDir(DefaultPath)
		if len(providers) == 1 {
			provider = providers[0].Name()
		} else if len(providers) > 1 {
			return "", fmt.Errorf("More than one provider found in %s, please specify one\n", DefaultPath)
		}
	}

	if provider == "" {
		return "", fmt.Errorf("Could not resolve a provider!")
	}

	provider, err := expandPath(provider)
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(provider); err != nil {
		return "", err
	}

	return provider, nil
}

// Call shells out to a provider and return its output
// If call succeeds, stdout is returned with no error
// If call fails, "" is return with error containing stderr
func Call(provider, specPath string) (string, error) {
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	cmd := exec.Command(provider, specPath)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()

	if err != nil {
		errstr := err.Error()
		if stdErr.Len() > 0 {
			errstr += ": " + strings.TrimSpace(stdErr.String())
		}
		return "", fmt.Errorf(errstr)
	}

	return stdOut.String(), nil
}

// Given a provider name, it returns a path to executable prefixed with DefaultPath. If
// the provider has any other pattern (eg. `./provider-name`, `/foo/provider-name`), the
// parameter is assumed to be a path to the provider and not just a name.
func expandPath(provider string) (string, error) {
	// Base returns just the last path segment.
	// If it's different, that means it's a (rel or abs) path
	if filepath.Base(provider) != provider {
		return filepath.Abs(provider)
	}

	return filepath.Join(DefaultPath, provider), nil
}

func getDefaultPath() string {
	pathOverride := os.Getenv("SUMMON_PROVIDER_PATH")

	if pathOverride != "" {
		return pathOverride
	}

	dir := "/usr/local/lib/summon"

	if runtime.GOOS == "windows" {
		// Try to get the appropriate "Program Files" directory but if one doesn't
		// exist, use a hardcoded value we think should be right.
		programFilesDir := os.Getenv("ProgramW6432")
		if programFilesDir == "" {
			programFilesDir = filepath.Join("C:", "Program Files")
		}

		dir = filepath.Join(programFilesDir, "Cyberark Conjur", "Summon", "Providers")
	}

	// found default installation directory
	if _, err := os.Stat(dir); err == nil {
		return dir
	}

	// finally: enable portable installation with Providers dir next to executable
	// if the direcotries above were not found
	exec, _ := os.Executable()
	execDir := filepath.Dir(exec)
	dir = filepath.Join(execDir, "Providers")

	return dir
}

// GetAllProviders creates slice of all file names in the default path
func GetAllProviders(providerDir string) ([]string, error) {
	files, err := ioutil.ReadDir(providerDir)
	if err != nil {
		return make([]string, 0), err
	}

	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name()
	}
	return names, nil
}
