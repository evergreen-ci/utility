package utility

import (
	"os"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// gitIgnoreFileMatcher contains the information for building a list of files in the given directory.
// It adds the files to include in the fileNames array and uses the ignorer to determine if a given
// file matches and should be added.
type gitIgnoreFileMatcher struct {
	ignorer *ignore.GitIgnore
	prefix  string
}

// NewGitIgnoreFileMatcher returns a FileMatcher that matches the
// expressions rooted at the given prefix. The expressions should be gitignore
// ignore expressions: antyhing that would be matched - and therefore ignored by
// git - is matched.
func NewGitIgnoreFileMatcher(prefix string, exprs ...string) FileMatcher {
	ignorer := ignore.CompileIgnoreLines(exprs...)
	m := &gitIgnoreFileMatcher{
		ignorer: ignorer,
		prefix:  prefix,
	}
	return m
}

func (m *gitIgnoreFileMatcher) Match(file string, info os.FileInfo) bool {
	file = strings.TrimLeft(strings.TrimPrefix(file, m.prefix), string(os.PathSeparator))
	return !info.IsDir() && m.ignorer.MatchesPath(file)
}
