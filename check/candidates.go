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

const (
	UpdateBinPrefix = "update"
)

func updateCandidates(prefix, dir string) ([]os.FileInfo, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fs = filter(fs, dirFilter)
	fs = filter(fs, sameFileFilter)
	fs = filter(fs, prefixFilter(prefix))

	if runtime.GOOS != "windows" {
		fs = filter(fs, executableFilter)
	}

	return fs, nil
}

type binVer struct {
	f os.FileInfo
	v *semver.Version
}

// byVersion implements sort.Interface for []binVer.
type byVersion []binVer

func (v byVersion) Len() int           { return len(v) }
func (v byVersion) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool { return v[i].v.LessThan(v[j].v) }

func versionFromBin(fname string) (*semver.Version, error) {
	outRaw, err := exec.Command("./"+fname, "-version").CombinedOutput()
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
