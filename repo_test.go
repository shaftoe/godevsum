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

func TestLastVer(t *testing.T) {
	var result, str, expected string
	gf, _ := NewGitFetcher("git", false)

	expected = ""
	str = "fückedUp\tbytes\nall overtheplace\t∆å…¡æ"
	gf.mockedOutput = []byte(str)
	// tags = TagsFromGitOutput([]byte(str))
	result, _ = lastVer("https://mock", "refs/tags/test", gf)
	validateExpectedStr(result, expected, t)

	expected = ""
	str = "386f2a698332b61278883df6f97d79eb98fe3f29\trefs/heads/master\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test01\na839bf2d274aaecd509b51ec37cb51842d4de348\trefs/tags/test02\n386f2a698332b61278883df6f97d79eb98fe3f29\t12.34"
	gf.mockedOutput = []byte(str)
	result, _ = lastVer("https://mock", "notExistent", gf)
	validateExpectedStr(result, expected, t)

	expected = "02"
	result, _ = lastVer("https://mock", "refs/tags/test", gf)
	validateExpectedStr(result, expected, t)

	expected = "12.34"
	result, _ = lastVer("https://mock", "", gf)
	validateExpectedStr(result, expected, t)

	expected = "12.34.56"
	str = str + "\n386f2a698332b61278883df6f97d79eb98fe3f29\trefs/tags/12.34.56\n"
	gf.mockedOutput = []byte(str)
	result, _ = lastVer("https://mock", "refs/tags/", gf)
	validateExpectedStr(result, expected, t)
}

func TestLatestVersion(t *testing.T) {
	var tags []string
	var expected, result string

	tags = []string{"1", "2.3.4", "0.12", "5.4.3", "5.3.4", "0.0"}
	result, _ = LatestVersion(tags)
	expected = "5.4.3"
	validateExpectedStr(result, expected, t)

	tags = []string{"1001.1002", "2.3", "0.12", "5.4.3", "0.0.0", "1001.130"}
	result, _ = LatestVersion(tags)
	expected = "1001.1002"
	validateExpectedStr(result, expected, t)

	tags = []string{}
	result, _ = LatestVersion(tags)
	expected = ""
	validateExpectedStr(result, expected, t)

	tags = []string{"1.2.3.4", "TEST", "10"}
	result, err := LatestVersion(tags)
	if err == nil || result != "TEST" {
		t.Error("LatestVersion should fail if not all tags are valid")
	}
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

func TestReplaceHostWithIp(t *testing.T) {
	cases := [][]string{
		{"https://localhost/restofurl", "https://::1/restofurl"},
		{"git://localhost/restofurl/is/long", "git://::1/restofurl/is/long"},
		{"http://localhost/restofurl", "http://::1/restofurl"},
	}

	for _, c := range cases {
		input, expected := c[0], c[1]
		result, _ := ReplaceHostWithIP(input)
		validateExpectedStr(result, expected, t)
	}

	if r, err := ReplaceHostWithIP("http2://google.com"); r != "" || err == nil {
		t.Error("https is not a valid transport and should fail")
	}
	if r, err := ReplaceHostWithIP("http://bogus"); r != "" || err == nil {
		t.Error("Bogus url should return empty string and error")
	}
	if r, err := ReplaceHostWithIP("http://bogusdomain"); r != "" || err == nil {
		t.Error("Bogus url should return empty string and error")
	}
}

func TestGitFetcher(t *testing.T) {
	gf, _ := NewGitFetcher("", false)

	if res := gf.GitPath(); res != "git" {
		t.Error("Default git path should be 'git', received", res)
	}
	if err := gf.SetGitPath("/etc/passwd", false); err == nil {
		t.Error("Paths not ending in /git should return error")
	}
	if err := gf.SetGitPath("bsdfafd/git", false); err == nil {
		t.Error("Not-existing paths should return error")
	}
	if err := gf.SetGitPath("/usr/bin/git", false); err != nil {
		t.Error("'/usr/bin/git' should be valid input, returned " + err.Error())
	}
	if res := gf.GitPath(); res != "/usr/bin/git" {
		t.Error("Updated git path should be '/usr/bin/git', received", res)
	}
}
