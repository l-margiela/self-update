package check

import (
	"os"
	"sort"

	"github.com/Masterminds/semver"
	"go.uber.org/zap"
)

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

func NewestCandidate(dir, currVersion string) (os.FileInfo, error) {
	cs, err := updateCandidates(".")
	if err != nil {
		return nil, err
	}

	if len(cs) == 0 {
		return nil, ErrNoCandidate
	}

	new, err := newest(currVersion, cs)
	if err != nil {
		return nil, err
	}

	return new, nil
}
