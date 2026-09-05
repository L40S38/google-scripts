// Package auth provides a shared OAuth2 desktop-app flow for the Google API
// samples in this repository.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GetClient returns an HTTP client authorized for the given scopes. It reads
// the OAuth client secret from credentialsPath and caches the resulting
// token in tokenPath together with the scopes it was granted for, so a
// cached token is only reused when it actually covers the requested scopes;
// otherwise the browser consent step runs again.
func GetClient(ctx context.Context, credentialsPath, tokenPath string, scopes ...string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("reading credentials file %q: %w", credentialsPath, err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing client secret file: %w", err)
	}

	tok, err := tokenFromFile(tokenPath, scopes)
	if err != nil {
		tok, err = tokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokenPath, scopes, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

// tokenFromWeb runs the loopback OAuth flow used by "Desktop app" clients:
// it starts a local HTTP server on an ephemeral port, prints the consent URL
// for the user to open, and waits for Google to redirect the browser back to
// that server with the authorization code. Desktop clients no longer support
// the old out-of-band flow of displaying a code for the user to paste.
func tokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting local callback listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	config.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d", port)

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generating state token: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			switch {
			case query.Get("state") != state:
				http.Error(w, "invalid state parameter", http.StatusBadRequest)
				errCh <- fmt.Errorf("callback received with mismatched state")
			case query.Get("error") != "":
				msg := query.Get("error")
				http.Error(w, "authorization failed: "+msg, http.StatusBadRequest)
				errCh <- fmt.Errorf("authorization declined: %s", msg)
			case query.Get("code") == "":
				http.Error(w, "missing code parameter", http.StatusBadRequest)
				errCh <- fmt.Errorf("callback received without an authorization code")
			default:
				fmt.Fprintln(w, "認証が完了しました。このタブは閉じてターミナルに戻ってください。")
				codeCh <- query.Get("code")
			}
		}),
	}
	go server.Serve(listener)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("ブラウザで次のURLを開いて認証してください:\n%v\n\n認証完了を待っています...\n", authURL)

	select {
	case code := <-codeCh:
		tok, err := config.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchanging authorization code: %w", err)
		}
		return tok, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// tokenCache is the on-disk format written to tokenPath. Storing the granted
// scopes alongside the token lets tokenFromFile reject a cached token that
// doesn't cover the scopes the current caller is requesting.
type tokenCache struct {
	Scopes string        `json:"scopes"`
	Token  *oauth2.Token `json:"token"`
}

func tokenFromFile(path string, scopes []string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cache tokenCache
	if err := json.NewDecoder(f).Decode(&cache); err != nil {
		return nil, err
	}
	if cache.Scopes != scopeKey(scopes) {
		return nil, fmt.Errorf("cached token in %q covers different scopes", path)
	}
	return cache.Token, nil
}

func saveToken(path string, scopes []string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("caching oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tokenCache{Scopes: scopeKey(scopes), Token: token})
}

// scopeKey normalizes a scope list into an order-independent cache key.
func scopeKey(scopes []string) string {
	sorted := append([]string(nil), scopes...)
	sort.Strings(sorted)
	return strings.Join(sorted, " ")
}
