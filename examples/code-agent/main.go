package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sithumli/Beacon/pkg/sdk"
)

// CodeRequest represents a code generation request
type CodeRequest struct {
	Prompt   string `json:"prompt"`
	Language string `json:"language,omitempty"`
}

// CodeResponse represents the generated code
type CodeResponse struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Model    string `json:"model"`
}

// OllamaRequest is the request format for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse is the response format from Ollama API
type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

var (
	ollamaURL   string
	ollamaModel string
	httpClient  *http.Client
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "A2A server address")
	host := flag.String("host", "localhost", "Agent host")
	port := flag.Int("port", 50053, "Agent port")
	flag.StringVar(&ollamaURL, "ollama-url", "http://localhost:11434", "Ollama API URL")
	flag.StringVar(&ollamaModel, "ollama-model", "codellama", "Ollama model to use")
	flag.Parse()

	httpClient = &http.Client{Timeout: 120 * time.Second}

	fmt.Println("Starting Code Generation Agent...")
	fmt.Printf("Using Ollama at %s with model %s\n", ollamaURL, ollamaModel)

	// Build the agent
	agent := sdk.NewAgent("CodeAgent").
		WithVersion("1.0.0").
		WithDescription("An LLM-powered agent that generates code from natural language").
		WithEndpoint(*host, *port).
		WithAuthor("dex").
		WithTags("code", "llm", "generation", "ollama").
		WithCapability("code-generation", "Generate code from natural language prompts", handleCodeGeneration).
		WithCapability("code-review", "Review and suggest improvements to code", handleCodeReview).
		WithCapability("code-explain", "Explain what code does", handleCodeExplain).
		Build()

	// Register with the platform
	if err := agent.Register(*serverAddr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Registered with ID: %s\n", agent.GetID())
	fmt.Printf("Listening on %s:%d\n", *host, *port)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start the agent
	if err := agent.Start(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Code Agent stopped")
}

func handleCodeGeneration(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req CodeRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	language := req.Language
	if language == "" {
		language = "python"
	}

	prompt := fmt.Sprintf(`Generate %s code for the following request. Only output the code, no explanations.

Request: %s

Code:`, language, req.Prompt)

	response, err := callOllama(ctx, prompt)
	if err != nil {
		return nil, err
	}

	resp := CodeResponse{
		Code:     response,
		Language: language,
		Model:    ollamaModel,
	}

	return json.Marshal(resp)
}

func handleCodeReview(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language,omitempty"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	prompt := fmt.Sprintf(`Review the following code and suggest improvements. Focus on:
1. Bugs or potential issues
2. Performance improvements
3. Code style and best practices

Code:
%s

Review:`, req.Code)

	response, err := callOllama(ctx, prompt)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{
		"review": response,
		"model":  ollamaModel,
	}

	return json.Marshal(resp)
}

func handleCodeExplain(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	prompt := fmt.Sprintf(`Explain what the following code does in simple terms:

%s

Explanation:`, req.Code)

	response, err := callOllama(ctx, prompt)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{
		"explanation": response,
		"model":       ollamaModel,
	}

	return json.Marshal(resp)
}

func callOllama(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  ollamaModel,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}
