package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Validator provides input validation functions
type Validator struct {
	// Discord ID regex pattern
	discordIDPattern *regexp.Regexp
	// Channel mention pattern
	channelPattern *regexp.Regexp
	// URL pattern for validation
	urlPattern *regexp.Regexp
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		discordIDPattern: regexp.MustCompile(`^\d{17,19}$`),
		channelPattern:   regexp.MustCompile(`^<#\d{17,19}>$`),
		urlPattern:       regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
	}
}

// ValidateDiscordID validates a Discord ID format
func (v *Validator) ValidateDiscordID(id string) error {
	if id == "" {
		return fmt.Errorf("Discord ID cannot be empty")
	}
	
	if !v.discordIDPattern.MatchString(id) {
		return fmt.Errorf("invalid Discord ID format: %s", id)
	}
	
	return nil
}

// ValidateChannelMention validates a Discord channel mention
func (v *Validator) ValidateChannelMention(mention string) error {
	if mention == "" {
		return fmt.Errorf("channel mention cannot be empty")
	}
	
	if !v.channelPattern.MatchString(mention) {
		return fmt.Errorf("invalid channel mention format: %s", mention)
	}
	
	return nil
}

// ValidateURL validates a URL format
func (v *Validator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	if !v.urlPattern.MatchString(url) {
		return fmt.Errorf("invalid URL format: %s", url)
	}
	
	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (v *Validator) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Limit length to prevent DoS
	if len(input) > 2000 { // Discord message limit
		input = input[:2000]
	}
	
	return input
}

// ValidateDiscordToken validates a Discord bot token format
func (v *Validator) ValidateDiscordToken(token string) error {
	if token == "" {
		return fmt.Errorf("Discord token cannot be empty")
	}
	
	// Basic format validation
	if len(token) < 50 {
		return fmt.Errorf("Discord token too short")
	}
	
	if !strings.Contains(token, ".") {
		return fmt.Errorf("invalid Discord token format")
	}
	
	// Check for common test tokens or placeholders
	lowerToken := strings.ToLower(token)
	if strings.Contains(lowerToken, "your_token") || 
	   strings.Contains(lowerToken, "bot_token") ||
	   strings.Contains(lowerToken, "example") ||
	   strings.Contains(lowerToken, "test") {
		return fmt.Errorf("placeholder token detected - please use a real Discord bot token")
	}
	
	return nil
}

// RateLimitInfo holds rate limiting information
type RateLimitInfo struct {
	Remaining int
	Reset     time.Time
	Limit     int
}

// SecurityHeaders adds security headers to HTTP responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Strict transport security (HTTPS only)
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", 
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self'; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'")
		
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions policy
		w.Header().Set("Permissions-Policy", 
			"geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=()")
		
		next.ServeHTTP(w, r)
	})
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// LogSanitizer sanitizes log messages to prevent log injection
type LogSanitizer struct{}

// NewLogSanitizer creates a new log sanitizer
func NewLogSanitizer() *LogSanitizer {
	return &LogSanitizer{}
}

// Sanitize sanitizes a log message
func (ls *LogSanitizer) Sanitize(message string) string {
	// Remove control characters that could be used for log injection
	message = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`).ReplaceAllString(message, "")
	
	// Remove ANSI escape sequences
	message = regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(message, "")
	
	// Limit length to prevent log flooding
	if len(message) > 1000 {
		message = message[:1000] + "..."
	}
	
	return message
}

// IPWhitelist manages IP address whitelisting
type IPWhitelist struct {
	allowedIPs map[string]bool
}

// NewIPWhitelist creates a new IP whitelist
func NewIPWhitelist(ips []string) *IPWhitelist {
	whitelist := &IPWhitelist{
		allowedIPs: make(map[string]bool),
	}
	
	for _, ip := range ips {
		whitelist.allowedIPs[ip] = true
	}
	
	return whitelist
}

// IsAllowed checks if an IP address is whitelisted
func (iw *IPWhitelist) IsAllowed(ip string) bool {
	return iw.allowedIPs[ip]
}

// Add adds an IP address to the whitelist
func (iw *IPWhitelist) Add(ip string) {
	iw.allowedIPs[ip] = true
}

// Remove removes an IP address from the whitelist
func (iw *IPWhitelist) Remove(ip string) {
	delete(iw.allowedIPs, ip)
}

// Global validator instance
var globalValidator = NewValidator()

// Package-level validation functions
func ValidateDiscordID(id string) error {
	return globalValidator.ValidateDiscordID(id)
}

func ValidateChannelMention(mention string) error {
	return globalValidator.ValidateChannelMention(mention)
}

func ValidateURL(url string) error {
	return globalValidator.ValidateURL(url)
}

func SanitizeInput(input string) string {
	return globalValidator.SanitizeInput(input)
}

func ValidateDiscordToken(token string) error {
	return globalValidator.ValidateDiscordToken(token)
}