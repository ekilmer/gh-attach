package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/sudosubin/gh-attach/internal/browserprovider"
	"github.com/sudosubin/gh-attach/internal/cookies"
)

func TestBrowserSessionTokenReportsCookieExpiry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	expires := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	svc := NewService(nil)
	svc.loginResolver = fakeLoginResolver{login: "octocat"}
	svc.providers = map[cookies.Browser]browserprovider.BrowserProvider{
		cookies.BrowserFirefox: stubProvider{sessions: []browserprovider.BrowserSession{{
			Browser: cookies.BrowserFirefox,
			Cookies: []*http.Cookie{
				{Name: "dotcom_user", Value: "octocat", Domain: "github.com", Path: "/"},
				{Name: "user_session", Value: "browser-session", Domain: "github.com", Path: "/", Expires: expires},
			},
		}}},
	}

	token, gotExpiry, err := svc.BrowserSessionToken(t.Context(), "github.com", cookies.ResolveInput{Browser: "firefox"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if token != "browser-session" || !gotExpiry.Equal(expires) {
		t.Fatalf("token and expiry = %q, %s; want browser-session, %s", token, gotExpiry, expires)
	}
}

func TestBrowserSessionTokenRequiresLoginCookie(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	svc := NewService(nil)
	svc.loginResolver = fakeLoginResolver{login: "octocat"}
	svc.providers = map[cookies.Browser]browserprovider.BrowserProvider{
		cookies.BrowserFirefox: stubProvider{sessions: []browserprovider.BrowserSession{{
			Browser: cookies.BrowserFirefox,
			Cookies: []*http.Cookie{{Name: "dotcom_user", Value: "octocat", Domain: "github.com", Path: "/"}},
		}}},
	}

	if _, _, err := svc.BrowserSessionToken(t.Context(), "github.com", cookies.ResolveInput{Browser: "firefox"}, false); err == nil {
		t.Fatal("expected an error for a missing user_session cookie")
	}
}
