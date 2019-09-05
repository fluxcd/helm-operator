package v2

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (h *HelmV2) DependencyUpdate(chartPath string) error {
	var hasLockFile bool

	// sanity check: does the chart directory exist
	if chartPath == "" {
		return errors.New("empty path to chart supplied")
	}
	chartInfo, err := os.Stat(chartPath)
	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("chart path %s does not exist", chartPath)
	case err != nil:
		return err
	case !chartInfo.IsDir():
		return fmt.Errorf("chart path %s is not a directory", chartPath)
	}

	// check if the requirements file exists
	reqFilePath := filepath.Join(chartPath, "requirements.yaml")
	reqInfo, err := os.Stat(reqFilePath)
	if err != nil || reqInfo.IsDir() {
		return nil
	}

	// We are going to use `helm dep build`, which tries to update the
	// dependencies in charts/ by looking at the file
	// `requirements.lock` in the chart directory. If the lockfile
	// does not match what is specified in requirements.yaml, it will
	// error out.
	//
	// If that file doesn't exist, `helm dep build` will fall back on
	// `helm dep update`, which populates the charts/ directory _and_
	// creates the lockfile. So that it will have the same behaviour
	// the next time it attempts a release, remove the lockfile if it
	// was created by helm.
	lockfilePath := filepath.Join(chartPath, "requirements.lock")
	info, err := os.Stat(lockfilePath)
	hasLockFile = (err == nil && !info.IsDir())
	if !hasLockFile {
		defer os.Remove(lockfilePath)
	}

	cmd := exec.Command("helm", "repo", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not update repo: %s", string(out))
	}

	cmd = exec.Command("helm", "dep", "build", ".")
	cmd.Dir = chartPath

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not update dependencies in %s: %s", chartPath, string(out))
	}

	return nil
}
