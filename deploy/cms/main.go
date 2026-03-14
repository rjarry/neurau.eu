package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	clientID      string
	clientSecret  string
	webhookSecret string
	siteDir       string
	listenAddr    string

	states   = make(map[string]time.Time)
	statesMu sync.Mutex

	buildMu sync.Mutex
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func main() {
	clientID = mustEnv("GITHUB_CLIENT_ID")
	clientSecret = mustEnv("GITHUB_CLIENT_SECRET")
	webhookSecret = mustEnv("GITHUB_WEBHOOK_SECRET")
	siteDir = os.Getenv("SITE_DIR")
	if siteDir == "" {
		siteDir = "/var/www/neurau.eu"
	}
	listenAddr = os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8090"
	}

	go cleanupStates()

	http.HandleFunc("/oauth/auth", handleAuth)
	http.HandleFunc("/oauth/callback", handleCallback)
	http.HandleFunc("/webhook", handleWebhook)

	log.Printf("listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func randomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func cleanupStates() {
	for {
		time.Sleep(time.Minute)
		statesMu.Lock()
		for k, t := range states {
			if time.Since(t) > 10*time.Minute {
				delete(states, k)
			}
		}
		statesMu.Unlock()
	}
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, err := randomState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Printf("oauth: failed to generate state: %v", err)
		return
	}

	statesMu.Lock()
	states[state] = time.Now()
	statesMu.Unlock()

	params := url.Values{
		"client_id":    {clientID},
		"redirect_uri": {r.URL.Query().Get("redirect_uri")},
		"scope":        {"public_repo"},
		"state":        {state},
	}

	target := "https://github.com/login/oauth/authorize?" + params.Encode()
	http.Redirect(w, r, target, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		oauthError(w, "missing code or state")
		return
	}

	statesMu.Lock()
	_, ok := states[state]
	if ok {
		delete(states, state)
	}
	statesMu.Unlock()

	if !ok {
		oauthError(w, "invalid or expired state")
		return
	}

	token, err := exchangeCode(code)
	if err != nil {
		oauthError(w, "token exchange failed")
		log.Printf("oauth: token exchange failed: %v", err)
		return
	}

	oauthSuccess(w, token)
}

func exchangeCode(code string) (string, error) {
	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest(http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github: %s", result.Error)
	}

	return result.AccessToken, nil
}

func oauthSuccess(w http.ResponseWriter, token string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><body><script>
(function() {
  window.opener.postMessage(
    'authorization:github:success:{"token":"%s","provider":"github"}',
    document.referrer
  );
})();
</script></body></html>`, token)
}

func oauthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `<!DOCTYPE html><html><body><script>
(function() {
  window.opener.postMessage(
    'authorization:github:error:{"message":"%s"}',
    document.referrer
  );
})();
</script></body></html>`, msg)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(body, sig) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		log.Printf("webhook: invalid signature")
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "pong")
		return
	}
	if event != "push" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ignored")
		return
	}

	var payload struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if payload.Ref != "refs/heads/main" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ignored branch")
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "building")

	go rebuild()
}

func verifySignature(body []byte, sig string) bool {
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}
	sig = strings.TrimPrefix(sig, "sha256=")
	expected, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), expected)
}

func rebuild() {
	buildMu.Lock()
	defer buildMu.Unlock()

	log.Printf("build: pulling changes")

	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = siteDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("build: git pull failed: %v", err)
		return
	}

	log.Printf("build: running hugo")

	cmd = exec.Command("hugo", "--minify")
	cmd.Dir = siteDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("build: hugo failed: %v", err)
		return
	}

	log.Printf("build: done")
}
