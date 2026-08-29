package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	PairID             string `json:"pair_id"`
	PhonePublicKey     string `json:"phone_public_key"`
	ComputerPublicKey  string `json:"computer_public_key"`
	ComputerPrivateKey string `json:"computer_private_key"`
	LockTimeoutSeconds int    `json:"lock_timeout_seconds"`
	PairedAt           string `json:"paired_at,omitempty"`
}

type State struct {
	mu              sync.RWMutex
	cfg             Config
	cfgPath         string
	pairCode        string
	pairMode        bool
	lastSeen        time.Time
	lastNonce       string
	lastChallengeAt time.Time
	wasPresent      bool
	dryRun          bool
}

type pairRequest struct {
	Code           string `json:"code"`
	PhonePublicKey string `json:"phone_public_key"`
}
type pairResponse struct {
	PairID            string `json:"pair_id"`
	ComputerPublicKey string `json:"computer_public_key"`
}
type challengeResponse struct {
	PairID            string `json:"pair_id"`
	Nonce             string `json:"nonce"`
	TimestampMS       int64  `json:"timestamp_ms"`
	ComputerPublicKey string `json:"computer_public_key"`
	SignatureB64      string `json:"signature_b64"`
}
type heartbeatRequest struct {
	PairID       string `json:"pair_id"`
	Nonce        string `json:"nonce"`
	TimestampMS  int64  `json:"timestamp_ms"`
	SignatureB64 string `json:"signature_b64"`
}
type statusResponse struct {
	Paired             bool   `json:"paired"`
	Present            bool   `json:"present"`
	LastSeen           string `json:"last_seen,omitempty"`
	LockTimeoutSeconds int    `json:"lock_timeout_seconds"`
	Platform           string `json:"platform"`
}

func main() {
	listen := flag.String("listen", "0.0.0.0:45873", "LAN listen address")
	cfgPath := flag.String("config", defaultConfigPath(), "config file path")
	timeout := flag.Int("timeout", 20, "seconds before locking after presence is lost")
	pairing := flag.Bool("pair", false, "allow/re-open pairing and print a one-time code")
	dryRun := flag.Bool("dry-run", false, "log lock/unlock actions without executing them")
	flag.Parse()

	st := &State{cfgPath: *cfgPath, dryRun: *dryRun}
	if err := st.loadOrCreate(*timeout); err != nil {
		log.Fatal(err)
	}
	if st.cfg.PhonePublicKey == "" || *pairing {
		st.pairMode = true
		st.pairCode = numericCode(6)
		fmt.Printf("\nNekoPresence pairing code: %s\n", st.pairCode)
		fmt.Printf("Computer public key: %s\n", st.cfg.ComputerPublicKey)
		fmt.Printf("Pair from Android using this computer's LAN IP and port 45873.\n\n")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/pair", st.handlePair)
	mux.HandleFunc("/v1/challenge", st.handleChallenge)
	mux.HandleFunc("/v1/heartbeat", st.handleHeartbeat)
	mux.HandleFunc("/v1/status", st.handleStatus)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })

	go st.monitor()
	srv := &http.Server{Addr: *listen, Handler: lanOnly(mux), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("NekoPresence agent listening on %s (%s)", *listen, runtime.GOOS)
	log.Fatal(srv.ListenAndServe())
}

func (s *State) loadOrCreate(timeout int) error {
	b, err := os.ReadFile(s.cfgPath)
	if err == nil {
		if err := json.Unmarshal(b, &s.cfg); err != nil {
			return err
		}
		if s.cfg.LockTimeoutSeconds == 0 {
			s.cfg.LockTimeoutSeconds = timeout
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	s.cfg = Config{
		PairID:             randomHex(16),
		ComputerPublicKey:  hex.EncodeToString(pub),
		ComputerPrivateKey: base64.StdEncoding.EncodeToString(priv),
		LockTimeoutSeconds: timeout,
	}
	return s.save()
}

func (s *State) save() error {
	if err := os.MkdirAll(filepath.Dir(s.cfgPath), 0700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s.cfg, "", "  ")
	return os.WriteFile(s.cfgPath, b, 0600)
}

func (s *State) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.pairMode {
		http.Error(w, "pairing closed", 403)
		return
	}
	var req pairRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if strings.TrimSpace(req.Code) != s.pairCode {
		http.Error(w, "bad pairing code", 401)
		return
	}
	raw, err := hex.DecodeString(strings.TrimSpace(req.PhonePublicKey))
	if err != nil || len(raw) != ed25519.PublicKeySize {
		http.Error(w, "bad phone key", 400)
		return
	}
	s.cfg.PhonePublicKey = strings.ToLower(strings.TrimSpace(req.PhonePublicKey))
	s.cfg.PairedAt = time.Now().UTC().Format(time.RFC3339)
	s.pairMode = false
	s.pairCode = ""
	if err := s.save(); err != nil {
		http.Error(w, "save failed", 500)
		return
	}
	writeJSON(w, pairResponse{PairID: s.cfg.PairID, ComputerPublicKey: s.cfg.ComputerPublicKey})
	log.Printf("paired Android key %s…", s.cfg.PhonePublicKey[:12])
}

func (s *State) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.PhonePublicKey == "" {
		http.Error(w, "not paired", 403)
		return
	}
	nonce := randomHex(32)
	ts := time.Now().UnixMilli()
	payload := fmt.Sprintf("%s|%s|%d", s.cfg.PairID, nonce, ts)
	privBytes, err := base64.StdEncoding.DecodeString(s.cfg.ComputerPrivateKey)
	if err != nil {
		http.Error(w, "key error", 500)
		return
	}
	sig := ed25519.Sign(ed25519.PrivateKey(privBytes), []byte(payload))
	s.lastNonce = nonce
	s.lastChallengeAt = time.Now()
	writeJSON(w, challengeResponse{PairID: s.cfg.PairID, Nonce: nonce, TimestampMS: ts, ComputerPublicKey: s.cfg.ComputerPublicKey, SignatureB64: base64.StdEncoding.EncodeToString(sig)})
}

func (s *State) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req heartbeatRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.PairID != s.cfg.PairID || req.Nonce == "" || req.Nonce != s.lastNonce {
		http.Error(w, "challenge mismatch", 401)
		return
	}
	if time.Since(s.lastChallengeAt) > 15*time.Second || abs(time.Now().UnixMilli()-req.TimestampMS) > 15000 {
		http.Error(w, "stale", 401)
		return
	}
	pubRaw, _ := hex.DecodeString(s.cfg.PhonePublicKey)
	sig, err := base64.StdEncoding.DecodeString(req.SignatureB64)
	if err != nil {
		http.Error(w, "signature", 400)
		return
	}
	payload := fmt.Sprintf("%s|%s|%d", req.PairID, req.Nonce, req.TimestampMS)
	if !ed25519.Verify(ed25519.PublicKey(pubRaw), []byte(payload), sig) {
		http.Error(w, "signature", 401)
		return
	}
	s.lastSeen = time.Now()
	s.lastNonce = ""
	writeJSON(w, statusResponse{Paired: true, Present: true, LastSeen: s.lastSeen.UTC().Format(time.RFC3339), LockTimeoutSeconds: s.cfg.LockTimeoutSeconds, Platform: runtime.GOOS})
}

func (s *State) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p := !s.lastSeen.IsZero() && time.Since(s.lastSeen) <= time.Duration(s.cfg.LockTimeoutSeconds)*time.Second
	last := ""
	if !s.lastSeen.IsZero() {
		last = s.lastSeen.UTC().Format(time.RFC3339)
	}
	writeJSON(w, statusResponse{Paired: s.cfg.PhonePublicKey != "", Present: p, LastSeen: last, LockTimeoutSeconds: s.cfg.LockTimeoutSeconds, Platform: runtime.GOOS})
}

func (s *State) monitor() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for range t.C {
		s.mu.Lock()
		present := !s.lastSeen.IsZero() && time.Since(s.lastSeen) <= time.Duration(s.cfg.LockTimeoutSeconds)*time.Second
		shouldLock := s.wasPresent && !present
		shouldUnlock := !s.wasPresent && present
		s.wasPresent = present
		s.mu.Unlock()
		if shouldLock {
			if err := lockSession(s.dryRun); err != nil {
				log.Printf("lock failed: %v", err)
			}
		}
		if shouldUnlock {
			if err := unlockSession(s.dryRun); err != nil {
				log.Printf("unlock note: %v", err)
			}
		}
	}
}

func lanOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "forbidden", 403)
			return
		}
		ip := net.ParseIP(host)
		if ip == nil || !(ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()) {
			http.Error(w, "LAN only", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func randomHex(n int) string { b := make([]byte, n); rand.Read(b); return hex.EncodeToString(b) }
func numericCode(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	out := ""
	for _, x := range b {
		out += strconv.Itoa(int(x) % 10)
	}
	return out
}
func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
func defaultConfigPath() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("PROGRAMDATA"); d != "" {
			return filepath.Join(d, "NekoPresenceKey", "config.json")
		}
	}
	if os.Geteuid() == 0 {
		return "/etc/nekopresence/config.json"
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "nekopresence", "config.json")
	}
	return "./nekopresence-config.json"
}
