package detect

import (
	"os"
	"os/exec"
	"strings"
)

// Known LLM provider env var prefixes
var llmProviderEnvPrefixes = map[string]string{
	"ANTHROPIC_API_KEY": "claude",
	"OPENAI_API_KEY":    "openai",
	"GEMINI_API_KEY":    "gemini",
	"GOOGLE_API_KEY":    "gemini",
	"GROQ_API_KEY":      "groq",
	"REPLICATE_API_KEY": "replicate",
	"XAI_API_KEY":       "grok",
}

// Known coding agents
var knownCodingAgents = []string{
	"claude", // claude-code CLI
	"opencode",
	"aider",
	"cursor",
}

// DetectLLM detects LLM-related configuration
func DetectLLM() LLMDetected {
	result := LLMDetected{
		Providers: []string{},
	}

	// Detect providers from environment variables
	for envVar, provider := range llmProviderEnvPrefixes {
		if os.Getenv(envVar) != "" {
			result.Providers = append(result.Providers, provider)
		}
	}

	// Detect local LLM runtime (ollama)
	if isToolInstalled("ollama") {
		result.Local = &LocalLLM{
			Runtime: "ollama",
			Models:  getOllamaModels(),
		}
	}

	// Detect coding agents
	var agents []string
	for _, agent := range knownCodingAgents {
		if isToolInstalled(agent) {
			agents = append(agents, agent)
		}
	}
	if len(agents) > 0 {
		result.Coding = &Coding{Agents: agents}
	}

	return result
}

// getOllamaModels lists pulled ollama models
func getOllamaModels() []string {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var models []string
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// Skip header line
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			// Model name is first field, may include :tag
			modelName := strings.Split(fields[0], ":")[0]
			if modelName != "" {
				models = append(models, modelName)
			}
		}
	}

	return models
}
