package godevsum

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"testing"
)

func validateExpectedStr(result, expected string, t *testing.T) {
	if result != expected {
		t.Error("result:", result, "expected:", expected)
	}
}

func validateExpectedInt(test string, expected, result int, t *testing.T) {
	if result != expected {
		t.Error(test, ": expected", expected, "received", result)
	}
}

func TestChangelogUpToDate(t *testing.T) {
	filename := "CHANGELOG.md"
	match := []byte("## [" + version + "]")
	sr := `## \[` + versionRegexp + `\]`

	changelog, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error("Can not read", filename)
	}

	curIndex := bytes.Index(changelog, match)
	if curIndex < 0 {
		t.Error("Entry for version", version, "not found in", filename)
	}

	r, err := regexp.Compile(sr)
	if err != nil {
		t.Error("Can not compile regexp", sr)
	}

	firstIndex := r.FindIndex(changelog)
	switch {
	case firstIndex == nil:
		t.Error("Could not find any string matching", sr, "in", filename)
	case curIndex > firstIndex[0]:
		t.Error(version, "is not the latest entry in", filename)
	}
}

func TestLatestTaggedVersion(t *testing.T) {
	var result, str, expected string
	var tags []string

	expected = ""
	str = "fückedUp\tbytes\nall overtheplace\t∆å…¡æ"
	tags = TagsFromGitOutput([]byte(str))
	result = LatestTaggedVersion("https://mock", "refs/tags/test", tags...)
	validateExpectedStr(result, expected, t)

	expected = ""
	str = "386f2a698332b61278883df6f97d79eb98fe3f29\trefs/heads/master\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test01\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test02\n386f2a698332b61278883df6f97d79eb98fe3f29\t12.34"
	tags = TagsFromGitOutput([]byte(str))
	result = LatestTaggedVersion("https://mock", "notExistent", tags...)
	validateExpectedStr(result, expected, t)

	expected = "02"
	result = LatestTaggedVersion("https://mock", "refs/tags/test", tags...)
	validateExpectedStr(result, expected, t)

	expected = "12.34"
	result = LatestTaggedVersion("https://mock", "", tags...)
	validateExpectedStr(result, expected, t)

	expected = "12.34.56"
	str = str + "\n386f2a698332b61278883df6f97d79eb98fe3f29\trefs/tags/12.34.56\n"
	tags = TagsFromGitOutput([]byte(str))
	result = LatestTaggedVersion("https://mock", "refs/tags/", tags...)
	validateExpectedStr(result, expected, t)
}

func TestLatestVersion(t *testing.T) {
	var tags []string
	var expected, result string

	tags = []string{"1", "2.3.4", "0.12", "5.4.3", "5.3.4", "0.0"}
	result = LatestVersion(tags)
	expected = "5.4.3"
	validateExpectedStr(result, expected, t)

	tags = []string{"1001.1002", "2.3", "0.12", "5.4.3", "0.0.0", "1001.130"}
	result = LatestVersion(tags)
	expected = "1001.1002"
	validateExpectedStr(result, expected, t)

	tags = []string{}
	result = LatestVersion(tags)
	expected = ""
	validateExpectedStr(result, expected, t)

	tags = []string{"TEST", "1.2.3.4"}
	result = LatestVersion(tags)
	expected = "1.2.3.4"
	validateExpectedStr(result, expected, t)
}

func TestVersionInternals(t *testing.T) {
	n := Version{}

	if n.length != 0 {
		t.Error("unitialized Version should hold 0 length")
	}
	err := n.addElement(-1)
	if err == nil {
		t.Error("adding negative value should return an error")
	}

	n.addElement(1)
	if n.head != n.tail {
		t.Error("head and tail should be identical")
	}
	validateExpectedInt("n.head.val", 1, n.head.val, t)

	max := 10
	for i := 2; i <= max; i++ {
		n.addElement(i)
		validateExpectedInt("n.tail.val", i, n.tail.val, t)
	}
	validateExpectedInt("n.length", max, n.length, t)
	validateExpectedInt("n.tail.val", max, n.tail.val, t)
	if n.tail.next != nil {
		t.Error("n.tail.Next: expected nil, received", n.tail.next)
	}
}

func TestVersionApi(t *testing.T) {
	v := &Version{}

	validateExpectedInt("v.length", 0, v.length, t)
	if v.head != nil {
		t.Error("Version default val should be nil pointer")
	}

	v, err := NewVersion("1.notgood")
	if err == nil {
		t.Error("NewVersion should return error with bogus strings")
	}

	v, err = NewVersion("1.-2")
	if err == nil {
		t.Error("NewVersion should return error with negative strings")
	}

	v, _ = NewVersion("0")
	validateExpectedInt("v.head.val", 0, v.head.val, t)
	validateExpectedInt("v.length", 1, v.length, t)
	if len(v.slice) != 1 {
		t.Error("v.slice length should be 1, received", len(v.slice))
	}
	if v.slice[0] != "0" {
		t.Error("v.slice value should be 0, received", v.slice[0])
	}

	v, _ = NewVersion("1.0.0")
	validateExpectedInt("v.head.val", 1, v.head.val, t)
	validateExpectedInt("v.length", 3, v.length, t)

	v, _ = NewVersion("0.1.2.3.45")
	validateExpectedInt("v.head.val", 0, v.head.val, t)
	validateExpectedInt("v.tail.val", 45, v.tail.val, t)
	validateExpectedInt("v.length", 5, v.length, t)
}

func TestVersionCompare(t *testing.T) {
	var v1, v2 *Version

	v1, _ = NewVersion("0")
	v2, _ = NewVersion("0")
	validateExpectedInt("", 0, v1.Compare(v1), t)
	validateExpectedInt("", 0, v2.Compare(v2), t)
	validateExpectedInt("", 0, v1.Compare(v2), t)
	validateExpectedInt("", 0, v2.Compare(v1), t)

	v1, _ = NewVersion("0")
	v2, _ = NewVersion("0.0.0.0.0")
	validateExpectedInt("", 0, v1.Compare(v2), t)
	validateExpectedInt("", 0, v2.Compare(v1), t)

	v1, _ = NewVersion("1.2")
	v2, _ = NewVersion("3.4.5")
	validateExpectedInt("", 1, v1.Compare(v2), t)
	validateExpectedInt("", -1, v2.Compare(v1), t)

	v1, _ = NewVersion("6.0.10")
	v2, _ = NewVersion("6.1.0")
	validateExpectedInt("", 1, v1.Compare(v2), t)
	validateExpectedInt("", -1, v2.Compare(v1), t)

	v1, _ = NewVersion("10.9.8.7.6")
	v2, _ = NewVersion("10.9.8.7.6")
	validateExpectedInt("", 0, v1.Compare(v2), t)
	validateExpectedInt("", 0, v2.Compare(v1), t)

	v1, _ = NewVersion("10.9.8.7.6.1")
	v2, _ = NewVersion("10.9.8.7.6")
	validateExpectedInt("", -1, v1.Compare(v2), t)
	validateExpectedInt("", 1, v2.Compare(v1), t)

	v1, _ = NewVersion("0.0.0.1.0")
	v2, _ = NewVersion("0")
	validateExpectedInt("", -1, v1.Compare(v2), t)
	validateExpectedInt("", 1, v2.Compare(v1), t)
}

func TestBiggestVersion(t *testing.T) {
	max := "10.1.0"
	var c []*Version
	var biggest string

	v1, _ := NewVersion("1.2.3.4")
	v2, _ := NewVersion("0")
	v3, _ := NewVersion("9.8")
	v4, _ := NewVersion("1.0.0.0.0.1")
	v5, _ := NewVersion("1.0.0.0.0.2")
	v6, _ := NewVersion("2.3")
	v7, _ := NewVersion(max)
	v8, _ := NewVersion("10.0.9")

	c = []*Version{v1, v2, v3, v4, v5, v6, v7, v8}
	biggest = BiggestVersion(c)
	validateExpectedStr(biggest, max, t)

	c = []*Version{v7}
	biggest = BiggestVersion(c)
	validateExpectedStr(biggest, max, t)

	c = []*Version{}
	biggest = BiggestVersion(c)
	validateExpectedStr(biggest, "", t)
}
