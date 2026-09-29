package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ekilmer/gh-attach/internal/cookies"
)

// BrowserSessionToken returns the login cookie from the browser account matched
// to the active gh login. Expires is zero when the browser did not record it.
func (s *Service) BrowserSessionToken(ctx context.Context, host string, source cookies.ResolveInput, verbose bool) (string, time.Time, error) {
	session, err := s.resolveBrowserSession(ctx, host, source, verbose)
	if err != nil {
		return "", time.Time{}, err
	}
	values := cookies.ValuesForHost(session.Cookies, "user_session", host)
	if len(values) != 1 {
		return "", time.Time{}, fmt.Errorf("expected one user_session cookie for %s, found %d", host, len(values))
	}
	if _, err := newTokenSession(host, values[0]); err != nil {
		return "", time.Time{}, err
	}
	for _, cookie := range session.Cookies {
		if cookie != nil && cookie.Name == "user_session" && cookie.Value == values[0] {
			return values[0], cookie.Expires, nil
		}
	}
	return "", time.Time{}, fmt.Errorf("user_session cookie for %s disappeared", host)
}
