package godevsum

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var version = "0.1.0"

var semverRegexp = `(\d+\.)?(\d+\.)?(\*|\d+)`

type gitRepo struct {
	url string
}

// TagsFromGitOutput parses "git ls-remote" standard output to return only
// the right side of the output (e.g. HEAD) ignoring the left side (i.e. the commit id)
func TagsFromGitOutput(stdout []byte) []string {
	var tag, result []string
	out := strings.Split(string(stdout), "\n")
	for _, line := range out {
		tag = strings.Split(line, "\t")
		if len(tag) == 2 {
			result = append(result, tag[1])
		}
	}
	return result
}

func (repo *gitRepo) remoteTags() []string {
	stdout, err := exec.Command("git", "ls-remote", "--tags", repo.url).Output()
	if err != nil {
		log.Fatal(err)
	}
	return TagsFromGitOutput(stdout)
}

func matchingTags(tags []string, regexpPrefix string) []string {
	// regexp brutally cargoculted from
	// http://stackoverflow.com/questions/82064/a-regex-for-version-number-parsing
	var validTag = regexp.MustCompile("^" + regexpPrefix + semverRegexp + `$`)
	var result []string
	for _, tag := range tags {
		if validTag.MatchString(tag) {
			result = append(result, tag)
		}
	}
	return result
}

// LatestVersionTag returns the latest tagged version of the Git project guessing
// it from the latest reference name matching regexpPrefix. For example, for Go
// url is "https://go.googlesource.com/go" and regexPrefix is "refs/tags/go"
//
// Returns empty string if no tag matches given regexpPrefix
func LatestVersionTag(url string, regexpPrefix string, tags ...string) string {
	switch {
	case len(tags) == 0:
		repo := &gitRepo{url: url}
		tags = matchingTags(repo.remoteTags(), regexpPrefix)
	case len(tags) > 0:
		tags = matchingTags(tags, regexpPrefix)
	}

	if len(tags) > 0 {
		return tags[len(tags)-1][len(regexpPrefix):]
	}
	return ""
}
