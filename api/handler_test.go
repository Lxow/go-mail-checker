package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthHandler tests the /health endpoint
// Because untested code is just a disaster waiting to happen
func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "GET request should succeed",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "POST request should fail",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:           "PUT request should fail",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			HealthHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check content type for successful requests
			if !tt.expectError {
				expectedContentType := "application/json"
				if ct := w.Header().Get("Content-Type"); ct != expectedContentType {
					t.Errorf("Expected Content-Type %s, got %s", expectedContentType, ct)
				}

				// Parse and validate the response
				var response HealthResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to parse response JSON: %v", err)
				}

				if response.Status != "healthy" {
					t.Errorf("Expected status 'healthy', got '%s'", response.Status)
				}

				if response.Message == "" {
					t.Error("Expected non-empty message in health response")
				}
			}
		})
	}
}

// TestCheckDomainHandler tests the /check-domain endpoint
// The main attraction - testing our MX lookup functionality
func TestCheckDomainHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParam     string
		expectedStatus int
		expectError    bool
		description    string
	}{
		{
			name:           "Valid domain with MX records",
			method:         "GET",
			queryParam:     "?domain=gmail.com",
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully find MX records for gmail.com",
		},
		{
			name:           "Missing domain parameter",
			method:         "GET",
			queryParam:     "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should fail when no domain is provided",
		},
		{
			name:           "Empty domain parameter",
			method:         "GET",
			queryParam:     "?domain=",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should fail when domain is empty",
		},
		{
			name:           "Invalid domain format",
			method:         "GET",
			queryParam:     "?domain=invalid-domain",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should fail for malformed domain",
		},
		{
			name:           "Another invalid domain",
			method:         "GET",
			queryParam:     "?domain=.com",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should fail for domain starting with dot",
		},
		{
			name:           "POST request should fail",
			method:         "POST",
			queryParam:     "?domain=gmail.com",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			description:    "Should reject non-GET requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/check-domain"+tt.queryParam, nil)
			w := httptest.NewRecorder()

			CheckDomainHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for test: %s", tt.expectedStatus, w.Code, tt.description)
			}

			// Check content type
			expectedContentType := "application/json"
			if ct := w.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("Expected Content-Type %s, got %s", expectedContentType, ct)
			}

			// Parse response based on expectation
			if tt.expectError {
				var errorResp ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Fatalf("Failed to parse error response JSON: %v", err)
				}

				if errorResp.Error == "" {
					t.Error("Expected non-empty error in error response")
				}
			} else {
				var response DomainCheckResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to parse response JSON: %v", err)
				}

				if response.Domain == "" {
					t.Error("Expected non-empty domain in response")
				}

				// For gmail.com, we expect MX records to exist
				if tt.queryParam == "?domain=gmail.com" && !response.HasMX {
					t.Error("Expected gmail.com to have MX records")
				}
			}
		})
	}
}

// TestIsValidDomain tests our domain validation function
// Because regex is tricky and domain validation is trickier
func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		domain string
		valid  bool
		reason string
	}{
		{"gmail.com", true, "standard domain should be valid"},
		{"google.com", true, "another standard domain should be valid"},
		{"sub.domain.com", true, "subdomain should be valid"},
		{"example.org", true, "different TLD should be valid"},
		{"", false, "empty string should be invalid"},
		{"invaliddomain", false, "domain without TLD should be invalid"},
		{".com", false, "domain starting with dot should be invalid"},
		{"domain.", false, "domain ending with dot should be invalid"},
		{"domain..com", false, "domain with double dots should be invalid"},
		{"domain com", false, "domain with space should be invalid"},
		{"domain_with_underscore.com", false, "domain with underscore should be invalid"},
		{"verylongdomainnamethatshouldnotbevalidbecauseitistoolongandexceedsthelimitofdnsandthisshouldmakeittoolongwithmorethantwohundredfiftythreecharactersasrequiredbythednsstandardsomeletsmakeitevenlongerandlongerandlongertoexceedthelimitandnowweneedtoaddmoretexttogetoveratotaloftwohundredfiftythreecharacters.com", false, "very long domain should be invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := isValidDomain(tt.domain)
			if result != tt.valid {
				t.Errorf("isValidDomain('%s') = %v, expected %v (%s)", tt.domain, result, tt.valid, tt.reason)
			}
		})
	}
}

// BenchmarkHealthHandler benchmarks the health endpoint
// Because performance matters (even for simple endpoints)
func BenchmarkHealthHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/health", nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		HealthHandler(w, req)
	}
}

// BenchmarkCheckDomainHandler benchmarks the domain check endpoint
// Warning: This will actually perform DNS lookups, so it's network-dependent
func BenchmarkCheckDomainHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/check-domain?domain=gmail.com", nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		CheckDomainHandler(w, req)
	}
}