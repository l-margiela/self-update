package check

import (
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver"
	"go.uber.org/zap"
)

// newest returns the newest binary.
//
// It does not take into account the commit hash.
func newest(currV string, fs []os.FileInfo) (Candidate, error) {
	curr, err := semver.NewVersion(currV)
	if err != nil {
		return Candidate{}, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return Candidate{}, err
	}

	var newer []Candidate
	for _, f := range fs {
		// FIXME: Potencial security vulnerability; research if fpath can be a malicious value.
		fpath := path.Join(cwd, f.Name())

		zap.L().Debug("check version", zap.String("bin", fpath))

		new, err := versionFromBin(fpath)
		if err != nil {
			zap.L().Debug("check version", zap.String("bin", fpath), zap.Error(err))
			continue
		}

		if curr.GreaterThan(new) || curr.Equal(new) {
			continue
		}

		newer = append(newer, Candidate{fpath, new})
	}
	sort.Sort(byVersion(newer))

	if len(newer) == 0 {
		return Candidate{}, ErrNoCandidate
	}

	return newer[len(newer)-1], nil
}

func NewestCandidate(dir, currVersion string) (Candidate, error) {
	cs, err := updateCandidates(".")
	if err != nil {
		return Candidate{}, err
	}

	if len(cs) == 0 {
		return Candidate{}, ErrNoCandidate
	}

	new, err := newest(currVersion, cs)
	if err != nil {
		return Candidate{}, err
	}

	return new, nil
}
