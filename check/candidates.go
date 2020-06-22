package check

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
)

// Candidate expresses an upgrade candidate.
type Candidate struct {
	Path    string // Binary path
	Version *semver.Version
}

func updateCandidates(fs []os.FileInfo) []os.FileInfo {
	fs = filter(fs, dirFilter)
	fs = filter(fs, sameFileFilter)

	if runtime.GOOS != "windows" {
		fs = filter(fs, executableFilter)
	}

	return fs
}

func updateCandidatesFromDir(dir string) ([]os.FileInfo, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fs = updateCandidates(fs)

	return fs, nil
}

// byVersion implements sort.Interface for []Candidate.
type byVersion []Candidate

func (v byVersion) Len() int           { return len(v) }
func (v byVersion) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool { return v[i].Version.LessThan(v[j].Version) }

func versionFromBin(fpath string) (*semver.Version, error) {
	outRaw, err := exec.Command(fpath, "-version").CombinedOutput()
	if err != nil {
		return nil, err
	}

	out := strings.TrimSpace(string(outRaw))

	new, err := semver.NewVersion(out)
	if err != nil {
		return nil, fmt.Errorf(`parse version "%s": %w`, out, err)
	}

	return new, nil
}
