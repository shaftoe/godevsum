package godevsum

import "testing"

func validate(result, expected string, t *testing.T) {
	if result != expected {
		t.Error("result:", result, "expected:", expected)
	}
}

func TestLatestMatchingTag(t *testing.T) {
	expected := ""
	s := "fückedUp\tbytes\nall overtheplace\t∆å…¡æ"
	tags := TagsFromGitOutput([]byte(s))
	result := LatestVersionTag("https://mock", "refs/tags/test", tags...)
	validate(result, expected, t)

	expected = ""
	s = "386f2a698332b61278883df6f97d79eb98fe3f29\trefs/heads/master\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test01\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test02\n386f2a698332b61278883df6f97d79eb98fe3f29\t1.2.3\n"
	tags = TagsFromGitOutput([]byte(s))
	result = LatestVersionTag("https://mock", "notExistent", tags...)
	validate(result, expected, t)

	expected = "02"
	result = LatestVersionTag("https://mock", "refs/tags/test", tags...)
	validate(result, expected, t)

	expected = "1.2.3"
	result = LatestVersionTag("https://mock", "", tags...)
	validate(result, expected, t)
}
