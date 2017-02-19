package godevsum

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const version = "0.3.1"

// We are only interested in "stable" versions, so we ignore
// strings and only look for digits.
const versionRegexp = `(\d+\.)*(\d+)`

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
//
// Returns empty string if versions is an empty slice
func LatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	vs := make([]*Version, len(versions))
	for i, r := range versions {
		v, err := NewVersion(r)
		if err != nil {
			// default to version 0
			vs[i], _ = NewVersion("0")
		} else {
			vs[i] = v
		}
	}

	return BiggestVersion(vs)
}

// LatestTaggedVersion returns the latest tagged version of the Git project matching
// regexpPrefix. For example, for Go url is "https://go.googlesource.com/go" and
// regexPrefix is "refs/tags/go"
//
// Returns empty string if no tag matches given regexpPrefix
func LatestTaggedVersion(url string, regexpPrefix string, tags ...string) string {
	switch l := len(tags); {
	case l == 0:
		repo := &gitRepo{url: url}
		tags = matchingTags(repo.remoteTags(), regexpPrefix)
	case l > 0: // This case is only used to inject tags for testing purpose
		tags = matchingTags(tags, regexpPrefix)
	}

	// Remove the prefix from the tags
	versions := make([]string, len(tags))
	for i, tag := range tags {
		versions[i] = tag[len(regexpPrefix):]
	}

	return LatestVersion(versions)
}

// Version represents a (software) version, in the dotted form
// <major>.<minor>.<patch>.<etc...>.
//
// Plese refer to NewVersion() constructor.
type Version struct {
	head   *versionElement
	tail   *versionElement
	slice  []string
	length int
}

type versionElement struct {
	val  int
	next *versionElement
}

func (v *Version) addElement(value int) error {
	if value < 0 {
		return errors.New("value must be a positive integer")
	}
	if v.head == nil { // unitialized
		v.head = &versionElement{val: value}
		v.tail = v.head
	} else { // we need to append a new versionElement
		v.head.next = &versionElement{val: value}
		v.tail = v.head.next
	}
	v.length++
	return nil
}

// NewVersion creates a new Version parsing the input string.
//
// Valid string literals for a Version are "1.0" or "3.4.5.0" or "0".
func NewVersion(s string) (*Version, error) {
	if len(s) == 0 {
		return nil, errors.New("input string s can not be empty")
	}

	var validVer = regexp.MustCompile(`^` + versionRegexp + `$`)
	if !validVer.Match([]byte(s)) {
		return nil, errors.New(s + " is not a valid version")
	}

	ver := &Version{}
	ver.slice = strings.Split(s, ".")

	for _, val := range ver.slice {
		n, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return nil, err
		}
		err = ver.addElement(int(n))
		if err != nil {
			return nil, err
		}
	}

	return ver, nil
}

// String converts a Version type to a string.
func (v *Version) String() string {
	return strings.Join(v.slice, ".")
}

// Compare compares two Version vars. It returns -1, 0, or 1 if
// the receiver version is smaller, equal, or larger than the version argument.
func (v *Version) Compare(o *Version) int {
	switch {
	// simple case: both are unitialized
	case v.head == nil && o.head == nil:
		return 0

	// simple case: head value differs
	case v.head.val > o.head.val:
		return -1
	case v.head.val < o.head.val:
		return 1
	// under this line, head value is equal

	// simple case: one element on both side and is equal
	case v.length == 1 && o.length == 1:
		return 0
	}

	// for the other cases, we cut Version's head and run compare again
	var vSlice, oSlice []string
	switch {
	case v.length == 1 && o.length > 1:
		vSlice, oSlice = []string{"0"}, o.slice[1:]
	case o.length == 1 && v.length > 1:
		vSlice, oSlice = v.slice[1:], []string{"0"}
	default:
		vSlice, oSlice = v.slice[1:], o.slice[1:]
	}

	v1, _ := NewVersion(strings.Join(vSlice, "."))
	o1, _ := NewVersion(strings.Join(oSlice, "."))

	return v1.Compare(o1)
}

// BiggestVersion returns the biggest Version from array in string format
func BiggestVersion(array []*Version) string {
	if len(array) < 1 {
		return ""
	}

	var max = array[0]
	for _, v := range array {
		if max.Compare(v) == 1 {
			max = v
		}
	}

	return max.String()
}
