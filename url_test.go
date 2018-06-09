package main_test

import (
	"testing"

	lib "github.com/fishy/blynk-proxy"
)

func TestRewriteURL(t *testing.T) {
	var orig, actual, expect string

	// should not rewrite
	orig = "https://blynk.cc"
	expect = orig
	actual = lib.RewriteURL(orig)
	if actual != expect {
		t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
	}

	// http blynk-cloud
	orig = "http://blynk-cloud.com"
	expect = "https://blynk-proxy.herokuapp.com"
	actual = lib.RewriteURL(orig)
	if actual != expect {
		t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
	}

	// https blynk-cloud
	orig = "https://blynk-cloud.com/foo/bar?foo=bar&baz=qux#asdf"
	expect = "https://blynk-proxy.herokuapp.com/foo/bar?foo=bar&baz=qux#asdf"
	actual = lib.RewriteURL(orig)
	if actual != expect {
		t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
	}

	// invalid url, should not rewrite
	orig = "al1i7y4hnelf  1lanlsu"
	expect = orig
	actual = lib.RewriteURL(orig)
	if actual != expect {
		t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
	}
}
