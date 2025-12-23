package detect

import (
	"os"
	"regexp"
	"strings"
)

// Patterns that suggest API keys/secrets
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^.*_API_KEY$`),
	regexp.MustCompile(`^.*_SECRET_KEY$`),
	regexp.MustCompile(`^.*_ACCESS_KEY$`),
	regexp.MustCompile(`^.*_TOKEN$`),
	regexp.MustCompile(`^ANTHROPIC_.*`),
	regexp.MustCompile(`^OPENAI_.*`),
	regexp.MustCompile(`^GEMINI_.*`),
	regexp.MustCompile(`^GROQ_.*`),
	regexp.MustCompile(`^REPLICATE_.*`),
	regexp.MustCompile(`^XAI_.*`),
	regexp.MustCompile(`^HUGGING_FACE_.*`),
	regexp.MustCompile(`^HF_.*`),
}

// Skip these even if they match patterns
var secretSkipList = map[string]bool{
	"GITHUB_TOKEN": true, // Pact uses this internally
	"GH_TOKEN":     true, // GitHub CLI token
}

// Common secret names to specifically look for
var commonSecrets = []string{
	"ANTHROPIC_API_KEY",
	"OPENAI_API_KEY",
	"GEMINI_API_KEY",
	"GROQ_API_KEY",
	"XAI_API_KEY",
	"REPLICATE_API_TOKEN",
	"HUGGING_FACE_TOKEN",
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
}

// DetectSecrets scans environment for secrets
// existingSecrets is the list from pact.json (can be nil)
func DetectSecrets(existingSecrets []string) []SecretDetected {
	var detected []SecretDetected
	existingSet := make(map[string]bool)
	for _, s := range existingSecrets {
		existingSet[s] = true
	}

	// Track what we've already added
	seen := make(map[string]bool)

	// First, check common secrets
	for _, name := range commonSecrets {
		if secretSkipList[name] {
			continue
		}
		if _, exists := os.LookupEnv(name); exists {
			detected = append(detected, SecretDetected{
				Name:       name,
				InEnv:      true,
				InKeychain: false, // Will be updated by caller if they have keyring access
				InPactJSON: existingSet[name],
			})
			seen[name] = true
		}
	}

	// Then scan all env vars for patterns
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]

		if seen[name] || secretSkipList[name] {
			continue
		}

		for _, pattern := range secretPatterns {
			if pattern.MatchString(name) {
				detected = append(detected, SecretDetected{
					Name:       name,
					InEnv:      true,
					InKeychain: false,
					InPactJSON: existingSet[name],
				})
				seen[name] = true
				break
			}
		}
	}

	return detected
}
