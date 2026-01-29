package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

// Request payload matching what the SDK sends
type SignRequest struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Body      string `json:"body"`
	Timestamp int64  `json:"timestamp"`
}

// Env vars
var (
	BuilderKey        = os.Getenv("BUILDER_KEY")
	BuilderSecret     = os.Getenv("BUILDER_SECRET")
	BuilderPassphrase = os.Getenv("BUILDER_PASSPHRASE")
)

func main() {
	if BuilderKey == "" || BuilderSecret == "" || BuilderPassphrase == "" {
		log.Fatal("Missing BUILDER_KEY, BUILDER_SECRET, or BUILDER_PASSPHRASE env vars")
	}

	http.HandleFunc("/v1/sign-builder", handleSign)
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Signer service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func handleSign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Logic from auth.BuildL2Headers but simplified for Builder
	// Message = timestamp + method + path + body
	message := fmt.Sprintf("%d%s%s", req.Timestamp, req.Method, req.Path)
	if req.Body != "" {
		message += req.Body
	}

	// Sign using the Secret (held securely on this server)
	sig, err := auth.SignHMAC(BuilderSecret, message)
	if err != nil {
		log.Printf("Signing error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Return the headers needed by Polymarket
	resp := map[string]string{
		auth.HeaderPolyBuilderAPIKey:     BuilderKey,
		auth.HeaderPolyBuilderPassphrase: BuilderPassphrase,
		auth.HeaderPolyBuilderTimestamp:  fmt.Sprintf("%d", req.Timestamp),
		auth.HeaderPolyBuilderSignature:  sig,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
