package godevsum

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

const version = "0.2.0"

// regexp brutally cargoculted from
// http://stackoverflow.com/questions/82064/a-regex-for-version-number-parsing
const versionRegexp = `(\d+\.)?(\d+\.)?(\d+)`

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
	var validTag = regexp.MustCompile("^" + regexpPrefix + versionRegexp + `$`)
	var result []string
	for _, tag := range tags {
		if validTag.MatchString(tag) {
			result = append(result, tag)
		}
	}
	return result
}

// LatestVersion returns biggest version according to semver semantic
// Returns empty string if versions is an empty slice
func LatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	vs := make([]*semver.Version, len(versions))
	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			// TODO log this to stderr or at least not to the user directly
			fmt.Println("Warning:", r, "is not a valid semver string, defaulting to 0.0.0")
			vs[i], err = semver.NewVersion("0.0.0")
		} else {
			vs[i] = v
		}
	}

	sort.Sort(semver.Collection(vs))
	return vs[len(vs)-1].String()
}

// LatestTaggedVersion returns the latest tagged version of the Git project matching
// regexpPrefix. For example, for Go url is "https://go.googlesource.com/go" and
// regexPrefix is "refs/tags/go"
//
// Returns empty string if no tag matches given regexpPrefix
func LatestTaggedVersion(url string, regexpPrefix string, tags ...string) string {
	switch {
	case len(tags) == 0:
		repo := &gitRepo{url: url}
		tags = matchingTags(repo.remoteTags(), regexpPrefix)
	case len(tags) > 0: // This case is only used to inject tags for testing purpose
		tags = matchingTags(tags, regexpPrefix)
	}

	// Remove the prefix from the tags
	versions := make([]string, len(tags))
	for i, tag := range tags {
		versions[i] = tag[len(regexpPrefix):]
	}

	return LatestVersion(versions)
}
