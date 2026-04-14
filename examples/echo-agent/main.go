package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sithumli/Beacon/pkg/sdk"
)

// EchoRequest represents the input for the echo capability
type EchoRequest struct {
	Message string `json:"message"`
}

// EchoResponse represents the output of the echo capability
type EchoResponse struct {
	Echo      string `json:"echo"`
	Processed bool   `json:"processed"`
}

func main() {
	serverAddr := flag.String("server", "localhost:8080", "A2A server address")
	host := flag.String("host", "localhost", "Agent host")
	port := flag.Int("port", 50052, "Agent port")
	flag.Parse()

	fmt.Println("Starting Echo Agent...")

	// Build the agent
	agent := sdk.NewAgent("EchoAgent").
		WithVersion("1.0.0").
		WithDescription("A simple agent that echoes back messages").
		WithEndpoint(*host, *port).
		WithAuthor("dex").
		WithTags("echo", "demo", "testing").
		WithCapability("echo", "Echoes back the input message", handleEcho).
		WithCapability("reverse", "Reverses the input message", handleReverse).
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

	// Start the agent (blocks until context is cancelled)
	if err := agent.Start(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Echo Agent stopped")
}

func handleEcho(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req EchoRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	resp := EchoResponse{
		Echo:      req.Message,
		Processed: true,
	}

	return json.Marshal(resp)
}

func handleReverse(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req EchoRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	// Reverse the string
	runes := []rune(req.Message)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	resp := EchoResponse{
		Echo:      string(runes),
		Processed: true,
	}

	return json.Marshal(resp)
}
