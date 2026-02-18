package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dl-alexandre/gdrv/internal/types"
	"golang.org/x/oauth2"
)

// OAuthFlow handles the OAuth2 authentication flow
type OAuthFlow struct {
	config       *oauth2.Config
	listener     net.Listener
	redirectURL  string
	state        string
	codeVerifier string
	codeChan     chan string
	errChan      chan error
}

// NewOAuthFlow creates a new OAuth flow handler
func NewOAuthFlow(config *oauth2.Config, listener net.Listener, redirectURL string) (*OAuthFlow, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth config not set")
	}

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	cfg := *config
	if redirectURL != "" {
		cfg.RedirectURL = redirectURL
	}
	if cfg.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL not set")
	}

	return &OAuthFlow{
		config:       &cfg,
		listener:     listener,
		redirectURL:  cfg.RedirectURL,
		state:        state,
		codeVerifier: verifier,
		codeChan:     make(chan string, 2),
		errChan:      make(chan error, 1),
	}, nil
}

// GetAuthURL returns the URL to redirect user for authentication
func (f *OAuthFlow) GetAuthURL() string {
	return f.config.AuthCodeURL(
		f.state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallengeS256(f.codeVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// StartCallbackServer starts the callback server and waits for auth code
func (f *OAuthFlow) StartCallbackServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", f.handleCallback)

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(f.listener); err != http.ErrServerClosed {
			f.errChan <- err
		}
	}()

	go func() {
		<-ctx.Done()
		server.Close()
	}()
}

// handleCallback processes the OAuth callback
func (f *OAuthFlow) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != f.state {
		f.errChan <- fmt.Errorf("invalid state parameter")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, callbackErrorHTML("Invalid state parameter. Please try authenticating again."))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error")
		if errMsg == "" {
			errMsg = "unknown error"
		}
		f.errChan <- fmt.Errorf("auth error: %s", errMsg)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, callbackErrorHTML(errMsg))
		return
	}

	f.codeChan <- code
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, callbackSuccessHTML(code))
}

func callbackSuccessHTML(code string) string {
	safeCode := html.EscapeString(code)
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>gdrv – Authentication Successful</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;
background:#0f172a;color:#e2e8f0;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#1e293b;border-radius:12px;padding:2.5rem;max-width:520px;width:90%;
box-shadow:0 25px 50px -12px rgba(0,0,0,.5);text-align:center}
.icon{font-size:3rem;margin-bottom:1rem}
h1{font-size:1.5rem;font-weight:600;margin-bottom:.5rem;color:#f8fafc}
.sub{color:#94a3b8;margin-bottom:1.5rem;font-size:.95rem;line-height:1.5}
.code-label{font-size:.8rem;text-transform:uppercase;letter-spacing:.05em;color:#64748b;margin-bottom:.5rem;text-align:left}
.code-wrap{position:relative;margin-bottom:1.5rem}
.code{background:#0f172a;border:1px solid #334155;border-radius:8px;padding:.85rem 3rem .85rem 1rem;
font-family:"SF Mono",SFMono-Regular,Consolas,"Liberation Mono",Menlo,monospace;font-size:.8rem;
color:#38bdf8;word-break:break-all;text-align:left;width:100%;cursor:text;
-webkit-user-select:all;user-select:all;display:block;line-height:1.5}
.copy-btn{position:absolute;right:.5rem;top:50%;transform:translateY(-50%);background:#334155;
border:none;color:#e2e8f0;border-radius:6px;padding:.4rem .65rem;cursor:pointer;font-size:.8rem;
transition:background .15s}
.copy-btn:hover{background:#475569}
.copy-btn.copied{background:#059669;color:#fff}
.hint{color:#64748b;font-size:.8rem}
</style>
</head>
<body>
<div class="card">
<div class="icon">&#10003;</div>
<h1>Authentication Successful</h1>
<p class="sub">gdrv received the authorization. You can close this tab and return to your terminal.</p>
<div class="code-label">Authorization Code</div>
<div class="code-wrap">
<code class="code" id="code">` + safeCode + `</code>
<button class="copy-btn" id="copyBtn" onclick="copyCode()">Copy</button>
</div>
<p class="hint">If the CLI didn't capture this automatically, copy the code and paste it in your terminal.</p>
</div>
<script>
function copyCode(){
var c=document.getElementById("code").textContent;
navigator.clipboard.writeText(c).then(function(){
var b=document.getElementById("copyBtn");b.textContent="Copied";b.classList.add("copied");
setTimeout(function(){b.textContent="Copy";b.classList.remove("copied")},2000)})
}
</script>
</body>
</html>`
}

func callbackErrorHTML(errMsg string) string {
	safeMsg := html.EscapeString(errMsg)
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>gdrv – Authentication Failed</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;
background:#0f172a;color:#e2e8f0;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#1e293b;border-radius:12px;padding:2.5rem;max-width:520px;width:90%;
box-shadow:0 25px 50px -12px rgba(0,0,0,.5);text-align:center}
.icon{font-size:3rem;margin-bottom:1rem}
h1{font-size:1.5rem;font-weight:600;margin-bottom:.5rem;color:#fca5a5}
.sub{color:#94a3b8;margin-bottom:1.5rem;font-size:.95rem;line-height:1.5}
.error-box{background:#0f172a;border:1px solid #7f1d1d;border-radius:8px;padding:.85rem 1rem;
font-family:"SF Mono",SFMono-Regular,Consolas,"Liberation Mono",Menlo,monospace;font-size:.85rem;
color:#f87171;word-break:break-all;text-align:left;margin-bottom:1.5rem;line-height:1.5}
.hint{color:#64748b;font-size:.8rem}
</style>
</head>
<body>
<div class="card">
<div class="icon">&#10007;</div>
<h1>Authentication Failed</h1>
<p class="sub">Something went wrong during the OAuth flow.</p>
<div class="error-box">` + safeMsg + `</div>
<p class="hint">Return to your terminal and try again with <code>gdrv auth login</code>.</p>
</div>
</body>
</html>`
}

// WaitForCode waits for the authorization code
func (f *OAuthFlow) WaitForCode(timeout time.Duration) (string, error) {
	select {
	case code := <-f.codeChan:
		return code, nil
	case err := <-f.errChan:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("authentication timed out")
	}
}

// WaitForCodeWithContext waits for the authorization code, respecting context cancellation.
func (f *OAuthFlow) WaitForCodeWithContext(ctx context.Context, timeout time.Duration) (string, error) {
	select {
	case code := <-f.codeChan:
		return code, nil
	case err := <-f.errChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(timeout):
		return "", fmt.Errorf("authentication timed out")
	}
}

// ExchangeCode exchanges auth code for tokens
func (f *OAuthFlow) ExchangeCode(ctx context.Context, code string) (*types.Credentials, error) {
	token, err := f.config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", f.codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return &types.Credentials{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiryDate:   token.Expiry,
		Scopes:       f.config.Scopes,
		Type:         types.AuthTypeOAuth,
	}, nil
}

// Close cleans up resources
func (f *OAuthFlow) Close() {
	if f.listener != nil {
		f.listener.Close()
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func codeChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// OAuthAuthOptions controls OAuth authentication behavior.
type OAuthAuthOptions struct {
	NoBrowser bool
}

// Authenticate performs the full OAuth flow.
//
// Code can arrive via the loopback callback server (browser redirect) or via
// manual paste on stdin (agent / headless fallback). Both race into codeChan;
// whichever delivers first wins. In browser mode the paste prompt is deferred
// 5 seconds so the happy path stays clean.
func (m *Manager) Authenticate(ctx context.Context, profile string, openBrowser func(string) error, opts OAuthAuthOptions) (*types.Credentials, error) {
	if m.oauthConfig == nil {
		return nil, fmt.Errorf("OAuth config not set")
	}

	flow, err := newLoopbackFlow(m.oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer flow.Close()

	authURL := flow.GetAuthURL()
	flow.StartCallbackServer(ctx)

	done := make(chan struct{})
	noBrowser := opts.NoBrowser || isHeadlessEnv()
	if noBrowser {
		fmt.Printf("Manual authentication required.\n")
		fmt.Printf("Open this URL in a browser and approve access:\n%s\n", authURL)
		fmt.Printf("\nPaste the authorization code here: ")
		go readCodeFromStdin(os.Stdin, flow.codeChan, done)
	} else {
		fmt.Printf("Opening browser for authentication...\n")
		fmt.Printf("If browser doesn't open, visit: %s\n", authURL)
		if err := openBrowser(authURL); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
			fmt.Printf("Open this URL manually in a browser:\n%s\n", authURL)
		}
		go showDelayedStdinPrompt(flow.codeChan, done)
	}

	code, err := flow.WaitForCodeWithContext(ctx, 5*time.Minute)
	close(done)
	if err != nil {
		return nil, err
	}

	creds, err := flow.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := m.SaveCredentials(profile, creds); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	return creds, nil
}

func showDelayedStdinPrompt(codeChan chan<- string, done <-chan struct{}) {
	select {
	case <-time.After(5 * time.Second):
		fmt.Printf("\nStill waiting... paste the authorization code here: ")
		readCodeFromStdin(os.Stdin, codeChan, done)
	case <-done:
		return
	}
}

func readCodeFromStdin(r io.Reader, codeChan chan<- string, done <-chan struct{}) {
	lineCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(r)
		if scanner.Scan() {
			lineCh <- scanner.Text()
		}
		close(lineCh)
	}()

	select {
	case line, ok := <-lineCh:
		if ok {
			code := strings.TrimSpace(line)
			if code != "" {
				codeChan <- code
			}
		}
	case <-done:
		return
	}
}

func newLoopbackFlow(config *oauth2.Config) (*OAuthFlow, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", addr.Port)
	return NewOAuthFlow(config, listener, redirectURL)
}

func isHeadlessEnv() bool {
	if os.Getenv("GDRV_NO_BROWSER") != "" {
		return true
	}
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return true
	}
	// DISPLAY/WAYLAND_DISPLAY check only applies to Linux; macOS and Windows
	// have their own display servers and never set these variables.
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return true
	}
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		return true
	}
	return false
}
