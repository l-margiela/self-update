package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"go.uber.org/zap"
)

const (
	updateBinPrefix = "update"
)

var ErrNoCandidate = errors.New("no candidate")

type filterPredicate func(f os.FileInfo) bool

// dirFilter filters out directories.
func dirFilter(f os.FileInfo) bool {
	return !f.IsDir()
}

// sameFileFilter filters out the binary itself.
func sameFileFilter(f os.FileInfo) bool {
	return os.Args[0] != f.Name()
}

// prefixFilter filters out binaries whose name is not prefixed with `prefix`.
func prefixFilter(prefix string) filterPredicate {
	return func(f os.FileInfo) bool {
		return strings.HasPrefix(f.Name(), prefix)
	}
}

// executableFilter filters out files without executable bit.
//
// It is overly simplified as it only checks if any of
//
// 1. User
// 2. Group
// 3. Others
//
// executable bit is set.
//
// It should check current proccess' ability to execute the file instead.
func executableFilter(f os.FileInfo) bool {
	return f.Mode()&0111 != 0
}

// filter filters out `fs` with `pred`.
func filter(fs []os.FileInfo, pred filterPredicate) []os.FileInfo {
	var filtered []os.FileInfo
	for _, f := range fs {
		if !pred(f) {
			continue
		}

		filtered = append(filtered, f)
	}

	return filtered
}

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

// newest returns the newest binary.
//
// It does not take into account the commit hash.
func newest(currV string, fs []os.FileInfo) (os.FileInfo, error) {
	curr, err := semver.NewVersion(currV)
	if err != nil {
		return nil, err
	}

	var newer []binVer
	for _, f := range fs {
		// Potencial security vulnerability; research if f.Name() can be a malicious value.
		new, err := versionFromBin(f.Name())
		if err != nil {
			zap.L().Debug("check version", zap.String("bin", f.Name()), zap.Error(err))
			continue
		}

		if curr.GreaterThan(new) || curr.Equal(new) {
			continue
		}

		newer = append(newer, binVer{f, new})
	}
	sort.Sort(byVersion(newer))

	if len(newer) == 0 {
		return nil, ErrNoCandidate
	}
	return newer[len(newer)-1].f, nil
}

func NewestCandidate(prefix, dir string) (os.FileInfo, error) {
	cs, err := updateCandidates(prefix, "./")
	if err != nil {
		return nil, err
	}
	if len(cs) == 0 {
		return nil, ErrNoCandidate
	}

	new, err := newest(Version, cs)
	if err != nil {
		return nil, err
	}

	return new, nil
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
	new, err := NewestCandidate(Version, "./")
	if err != nil {
		if errors.Is(err, ErrNoCandidate) {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				zap.L().Error("write response", zap.Error(err))
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			zap.L().Error("write response", zap.Error(err))
		}
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("cadidate: %v", new.Name()))); err != nil {
		zap.L().Error("write response", zap.Error(err))
	}
}
