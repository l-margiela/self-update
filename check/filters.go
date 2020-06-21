package check

import (
	"os"
	"strings"
)

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
