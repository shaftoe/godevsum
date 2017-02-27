package godevsum

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const version = "0.5.1"

// we are only interested in "stable" versions, so we ignore
// strings and only look for digits.
const versionRegexp = `(\d+\.)*(\d+)`

type gitRepo struct {
	url string
}

// GitPath returns the current binary path used to execute
// git commands.
func (gf *GitFetcher) GitPath() string {
	return gf.path
}

// SetGitPath sets given path as Git binary to be used.
// Returns error if path is not found or other errors occurred,
// for example when failing to enforce execution bit.
func (gf *GitFetcher) SetGitPath(path string, enforceExec bool) error {
	switch path {
	case "", "git":
		gf.path = "git"
		if !enforceExec {
			return nil
		}
	default:
		if len(path) < 4 || path[len(path)-4:] != "/git" {
			return errors.New(path + " is an invalid path, must end with /git")
		}
		if _, err := os.Lstat(path); err != nil {
			return err
		}
	}

	if enforceExec {
		if err := os.Chmod(path, 0755); err != nil {
			return err
		}
	}

	p, err := exec.LookPath(path)
	if err != nil {
		return err
	}
	gf.path = p
	return nil
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
// Returns empty string if versions is an empty slice. When unparsable
// string is found, return it together with an error
func LatestVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}

	vs := make([]*Version, len(versions))
	for i, raw := range versions {
		v, err := NewVersion(raw)
		if err != nil {
			return raw, err
		}
		vs[i] = v
	}

	return BiggestVersion(vs), nil
}

// GitFetcher represent the object which is responsible
// to retrieve information from the remote Git repository
type GitFetcher struct {
	// Unix path to the binary to be used
	path string

	// Mock data, used for testing
	mockedOutput []byte
}

// NewGitFetcher creates a GitFetcher type instance and returns
// a pointer to it. GitFetcher type is the interface to the Git binary
// used to fetch the data from the remote.
//
// If path is empty string, will try to use the default "git" found in
// the PATH if present, or fail otherwise.
func NewGitFetcher(path string, enforce bool) (*GitFetcher, error) {
	gf := &GitFetcher{path: path}
	err := gf.SetGitPath(path, enforce)
	return gf, err
}

func (gf *GitFetcher) fetchTags(repo *gitRepo) ([]string, error) {
	if gf.mockedOutput != nil {
		return TagsFromGitOutput(gf.mockedOutput), nil
	}
	cmd := exec.Command(gf.GitPath(), "ls-remote", "--tags", repo.url)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return []string{string(stdout)}, err
	}
	return TagsFromGitOutput(stdout), nil
}

// LatestTaggedVersion returns the latest tagged version of the Git project matching
// regexpPrefix. For example, for Go url is "https://go.googlesource.com/go" and
// regexPrefix is "refs/tags/go"
//
// Returns empty string if no tag matches given regexpPrefix
func LatestTaggedVersion(url string, regexpPrefix string, gf *GitFetcher) (string, error) {
	return lastVer(url, regexpPrefix, gf)
}

func lastVer(url string, regexpPrefix string, gf *GitFetcher) (string, error) {
	repo := &gitRepo{url: url}
	t, err := gf.fetchTags(repo)
	if err != nil {
		// first and only element of t contains the error message
		return t[0], err
	}
	mTags := matchingTags(t, regexpPrefix)
	versions := make([]string, len(mTags))

	// Remove the prefix from the tags
	for i, tag := range mTags {
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

// ReplaceHostWithIP performs a DNS lookup for the host in the given
// url and replaces the host string in the url with the IPv4 address.
func ReplaceHostWithIP(url string) (string, error) {
	// FIXME this is a weak check, add better match
	r := regexp.MustCompile(`^[a-zA-Z]+://`)
	if !r.MatchString(url) {
		return "", errors.New("Url " + url + " is invalid")
	}

	var host string

	var splitted = strings.Split(url, "/")
	if len(splitted) >= 3 {
		host = splitted[2]
	} else {
		return "", errors.New("Url " + url + " is invalid")
	}

	// TODO implement timeouts, maybe with goroutine
	addr, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}

	if len(addr) == 0 {
		return "", errors.New("DNS lookup for host " + host + " failed")
	}
	splitted[2] = addr[0]
	return strings.Join(splitted, "/"), nil
}
