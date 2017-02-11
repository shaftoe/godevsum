package godevsum

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"testing"
)

func validate(result, expected string, t *testing.T) {
	if result != expected {
		t.Error("result:", result, "expected:", expected)
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
	validate(result, expected, t)

	expected = ""
	str = "386f2a698332b61278883df6f97d79eb98fe3f29\trefs/heads/master\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test01\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test02\n386f2a698332b61278883df6f97d79eb98fe3f29\t12.34"
	tags = TagsFromGitOutput([]byte(str))
	result = LatestTaggedVersion("https://mock", "notExistent", tags...)
	validate(result, expected, t)

	expected = "2.0.0" // FIXME we should return 02 without paddings
	result = LatestTaggedVersion("https://mock", "refs/tags/test", tags...)
	validate(result, expected, t)

	expected = "12.34.0" // FIXME we should return 12.34 without padding
	result = LatestTaggedVersion("https://mock", "", tags...)
	validate(result, expected, t)

	expected = "12.34.56"
	str = str + "\n386f2a698332b61278883df6f97d79eb98fe3f29\trefs/tags/12.34.56\n"
	tags = TagsFromGitOutput([]byte(str))
	result = LatestTaggedVersion("https://mock", "refs/tags/", tags...)
	validate(result, expected, t)
}

func TestLatestVersion(t *testing.T) {
	var tags []string
	var expected, result string

	tags = []string{"1", "2.3.4", "0.12", "5.4.3", "5.3.4", "0.0"}
	result = LatestVersion(tags)
	expected = "5.4.3"
	validate(result, expected, t)

	tags = []string{"1001.1002", "2.3", "0.12", "5.4.3", "0.0.0", "1001.130"}
	result = LatestVersion(tags)
	expected = "1001.1002.0" // FIXME we should return v1001.1002 without padding
	validate(result, expected, t)

	tags = []string{}
	result = LatestVersion(tags)
	expected = ""
	validate(result, expected, t)

	tags = []string{"TEST", "1.2.3.4"} // This is not a valid semver
	result = LatestVersion(tags)
	expected = "0.0.0"
	validate(result, expected, t)
}
