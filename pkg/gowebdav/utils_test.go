package gowebdav

import (
	"fmt"
	"net/url"
	"testing"
)

func TestJoin(t *testing.T) {
	eq(t, "/", "", "")
	eq(t, "/", "/", "/")
	eq(t, "/foo", "", "/foo")
	eq(t, "foo/foo", "foo/", "/foo")
	eq(t, "foo/foo", "foo/", "foo")
}

func eq(t *testing.T, expected string, s0 string, s1 string) {
	s := Join(s0, s1)
	if s != expected {
		t.Error("For", "'"+s0+"','"+s1+"'", "expected", "'"+expected+"'", "got", "'"+s+"'")
	}
}

func ExamplePathEscape() {
	fmt.Println(PathEscape(""))
	fmt.Println(PathEscape("/"))
	fmt.Println(PathEscape("/web"))
	fmt.Println(PathEscape("/web/"))
	fmt.Println(PathEscape("/w e b/d a v/s%u&c#k:s/"))

	// Output:
	//
	// /
	// /web
	// /web/
	// /w%20e%20b/d%20a%20v/s%25u&c%23k:s/
}

func TestEscapeURL(t *testing.T) {
	ex := "https://foo.com/w%20e%20b/d%20a%20v/s%25u&c%23k:s/"
	u, _ := url.Parse("https://foo.com" + PathEscape("/w e b/d a v/s%u&c#k:s/"))
	if ex != u.String() {
		t.Error("expected: " + ex + " got: " + u.String())
	}
}

func TestFixSlashes(t *testing.T) {
	expected := "/"

	if got := FixSlashes(""); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	expected = "/path/"

	if got := FixSlashes("path"); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	if got := FixSlashes("/path"); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	if got := FixSlashes("path/"); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}
}
