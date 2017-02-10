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
	sr := `## \[` + semverRegexp + `\]`

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

func TestLatestMatchingTag(t *testing.T) {
	expected := ""
	s := "fückedUp\tbytes\nall overtheplace\t∆å…¡æ"
	tags := TagsFromGitOutput([]byte(s))
	result := LatestVersionTag("https://mock", "refs/tags/test", tags...)
	validate(result, expected, t)

	expected = ""
	s = "386f2a698332b61278883df6f97d79eb98fe3f29\trefs/heads/master\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test01\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test02\n386f2a698332b61278883df6f97d79eb98fe3f29\t12.34"
	tags = TagsFromGitOutput([]byte(s))
	result = LatestVersionTag("https://mock", "notExistent", tags...)
	validate(result, expected, t)

	expected = "02"
	result = LatestVersionTag("https://mock", "refs/tags/test", tags...)
	validate(result, expected, t)

	expected = "12.34"
	result = LatestVersionTag("https://mock", "", tags...)
	validate(result, expected, t)

	expected = "12.34.56"
	s = s + "\n386f2a698332b61278883df6f97d79eb98fe3f29\trefs/tags/12.34.56\n"
	tags = TagsFromGitOutput([]byte(s))
	result = LatestVersionTag("https://mock", "refs/tags/", tags...)
	validate(result, expected, t)
}
