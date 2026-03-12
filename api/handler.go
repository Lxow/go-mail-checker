package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// HealthResponse represents the API health status
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// DomainCheckResponse represents the domain MX check result
type DomainCheckResponse struct {
	Domain    string   `json:"domain"`
	HasMX     bool     `json:"has_mx"`
	MXRecords []string `json:"mx_records"`
	Message   string   `json:"message,omitempty"`
}

// ErrorResponse represents an API error
// For when things go sideways (which they inevitably do)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HealthHandler handles the /health endpoint
// Simple but effective - like a good Swiss watch
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Only GET requests allowed, sorry POST enthusiasts
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
		return
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Message:   "Go Mail Checker API is running smoothly!",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// CheckDomainHandler handles the /check-domain endpoint
// The main event - where the MX magic happens ✨
func CheckDomainHandler(w http.ResponseWriter, r *http.Request) {
	// Only GET requests, we're not changing anything here
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
		return
	}

	// Extract domain parameter
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	if domain == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Missing domain parameter", "Please provide a domain using ?domain=example.com")
		return
	}

	// Validate domain format (basic regex check)
	if !isValidDomain(domain) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid domain format", fmt.Sprintf("'%s' doesn't look like a valid domain", domain))
		return
	}

	log.Printf("Checking MX records for domain: %s", domain)

	// Look up MX records with timeout
	mxRecords, err := lookupMXRecords(domain)
	if err != nil {
		log.Printf("MX lookup failed for %s: %v", domain, err)
		writeErrorResponse(w, http.StatusInternalServerError, "MX lookup failed", err.Error())
		return
	}

	response := DomainCheckResponse{
		Domain:    domain,
		HasMX:     len(mxRecords) > 0,
		MXRecords: mxRecords,
	}

	if len(mxRecords) > 0 {
		response.Message = fmt.Sprintf("Found %d MX record(s) for %s", len(mxRecords), domain)
		log.Printf("Found %d MX record(s) for %s", len(mxRecords), domain)
	} else {
		response.Message = fmt.Sprintf("No MX records found for %s (domain might not accept emails)", domain)
		log.Printf("No MX records found for %s", domain)
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// isValidDomain checks if a domain string is roughly valid
// Not bulletproof, but catches the obvious mistakes
func isValidDomain(domain string) bool {
	// Check overall length first
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Basic domain regex - allows letters, numbers, dots, and hyphens
	// Updated to properly handle subdomains
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(domain)
}

// lookupMXRecords performs the actual MX lookup with timeout
// The heart of our operation - where DNS magic happens
func lookupMXRecords(domain string) ([]string, error) {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set a reasonable timeout for DNS lookups
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second, // 5 seconds should be plenty
			}
			return d.DialContext(ctx, network, address)
		},
	}

	// Look up MX records
	mxRecords, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed: %v", err)
	}

	// Convert MX records to string slice
	result := make([]string, len(mxRecords))
	for i, mx := range mxRecords {
		// Format: "priority hostname" (e.g., "10 mail.google.com.")
		result[i] = fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, "."))
	}

	return result, nil
}

// writeJSONResponse writes a JSON response with proper headers
// Because JSON without proper headers is just sadness
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", "1.0")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("💥 Failed to encode JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes a standardized error response
// For when Murphy's Law strikes (and it always does)
func writeErrorResponse(w http.ResponseWriter, statusCode int, error string, message string) {
	errorResp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	writeJSONResponse(w, statusCode, errorResp)
}