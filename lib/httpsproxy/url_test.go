package httpsproxy_test

import (
	"net/url"
	"testing"

	"github.com/fishy/blynk-proxy/lib/httpsproxy"
)

func TestRewriteURL(t *testing.T) {
	targetHost := "blynk-cloud.com"
	selfURL, err := url.Parse("https://blynk-proxy.appspot.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Run(
		"wrong host",
		func(t *testing.T) {
			orig := "https://blynk.cc"
			expect := orig
			actual := httpsproxy.RewriteURL(orig, targetHost, selfURL)
			if actual != expect {
				t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
			}
		},
	)

	t.Run(
		"http blynk-cloud",
		func(t *testing.T) {
			orig := "http://blynk-cloud.com"
			expect := "https://blynk-proxy.appspot.com"
			actual := httpsproxy.RewriteURL(orig, targetHost, selfURL)
			if actual != expect {
				t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
			}
		},
	)

	t.Run(
		"https blynk-cloud",
		func(t *testing.T) {
			orig := "https://blynk-cloud.com/foo/bar?foo=bar&baz=qux#asdf"
			expect := "https://blynk-proxy.appspot.com/foo/bar?foo=bar&baz=qux#asdf"
			actual := httpsproxy.RewriteURL(orig, targetHost, selfURL)
			if actual != expect {
				t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
			}
		},
	)

	t.Run(
		"invalid url",
		func(t *testing.T) {
			orig := "al1i7y4hnelf  1lanlsu"
			expect := orig
			actual := httpsproxy.RewriteURL(orig, targetHost, selfURL)
			if actual != expect {
				t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
			}
		},
	)

	t.Run(
		"no selfURL",
		func(t *testing.T) {
			orig := "https://blynk-cloud.com/foo/bar?foo=bar&baz=qux#asdf"
			expect := orig
			actual := httpsproxy.RewriteURL(orig, targetHost, nil)
			if actual != expect {
				t.Errorf("RewriteURL(%q) expected %q, got %q", orig, expect, actual)
			}
		},
	)
}
