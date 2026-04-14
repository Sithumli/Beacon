package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Sithumli/Beacon/pkg/sdk"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
	client     *sdk.Client
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "a2a",
		Short: "A2A Platform CLI - Manage agents and tasks",
		Long: `A2A Platform CLI provides commands to interact with the A2A Discovery Platform.

Use this tool to:
- List and inspect registered agents
- Discover agents by capability
- Send tasks to agents
- Monitor task status`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip client init for help commands
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}
			var err error
			client, err = sdk.NewClient(serverAddr)
			if err != nil {
				return fmt.Errorf("failed to connect to server: %w", err)
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if client != nil {
				client.Close()
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:8080", "A2A server address")

	// Add subcommands
	rootCmd.AddCommand(agentsCmd())
	rootCmd.AddCommand(discoverCmd())
	rootCmd.AddCommand(taskCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func agentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agents",
	}

	// List agents
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			agents, err := client.ListAgents(ctx)
			if err != nil {
				return err
			}

			if len(agents) == 0 {
				fmt.Println("No agents registered")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tVERSION\tSTATUS\tCAPABILITIES")
			fmt.Fprintln(w, "--\t----\t-------\t------\t------------")
			for _, agent := range agents {
				caps := make([]string, len(agent.Capabilities))
				for i, c := range agent.Capabilities {
					caps[i] = c.Name
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					truncate(agent.ID, 8),
					agent.Name,
					agent.Version,
					agent.Status,
					strings.Join(caps, ", "))
			}
			w.Flush()

			return nil
		},
	}

	// Get agent details
	getCmd := &cobra.Command{
		Use:   "get <agent-id>",
		Short: "Get agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			agent, err := client.GetAgent(ctx, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Agent: %s\n", agent.Name)
			fmt.Printf("  ID:          %s\n", agent.ID)
			fmt.Printf("  Version:     %s\n", agent.Version)
			fmt.Printf("  Description: %s\n", agent.Description)
			fmt.Printf("  Endpoint:    %s:%d\n", agent.Host, agent.Port)
			fmt.Printf("  Status:      %s\n", agent.Status)
			fmt.Printf("  Registered:  %s\n", agent.RegisteredAt.Format(time.RFC3339))
			fmt.Printf("  Heartbeat:   %s\n", agent.LastHeartbeat.Format(time.RFC3339))
			fmt.Printf("  Capabilities:\n")
			for _, cap := range agent.Capabilities {
				fmt.Printf("    - %s: %s\n", cap.Name, cap.Description)
			}

			return nil
		},
	}

	cmd.AddCommand(listCmd, getCmd)
	return cmd
}

func discoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discover <capability>",
		Short: "Find agents with a specific capability",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			agents, err := client.DiscoverAgents(ctx, args[0])
			if err != nil {
				return err
			}

			if len(agents) == 0 {
				fmt.Printf("No agents found with capability: %s\n", args[0])
				return nil
			}

			fmt.Printf("Found %d agent(s) with capability '%s':\n\n", len(agents), args[0])
			for _, agent := range agents {
				fmt.Printf("  %s (%s)\n", agent.Name, agent.ID)
				fmt.Printf("    Endpoint: %s:%d\n", agent.Host, agent.Port)
				fmt.Printf("    Status:   %s\n\n", agent.Status)
			}

			return nil
		},
	}
}

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	// Send task
	var toAgent, capability, payloadStr string
	var fromAgent string
	var wait bool

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send a task to an agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Parse payload
			var payload interface{}
			if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
				return fmt.Errorf("invalid payload JSON: %w", err)
			}

			var task *sdk.TaskInfo
			var err error

			if toAgent != "" {
				// Send to specific agent
				task, err = client.SendTask(ctx, fromAgent, toAgent, capability, payload)
			} else {
				// Route to any agent with capability
				task, err = client.RouteTask(ctx, fromAgent, capability, payload)
			}
			if err != nil {
				return err
			}

			fmt.Printf("Task created: %s\n", task.ID)
			fmt.Printf("  To Agent:   %s\n", task.ToAgent)
			fmt.Printf("  Capability: %s\n", task.Capability)
			fmt.Printf("  Status:     %s\n", task.Status)

			if wait {
				fmt.Println("\nWaiting for completion...")
				task, err = client.WaitForTask(ctx, task.ID, time.Second)
				if err != nil {
					return err
				}
				fmt.Printf("\nTask completed!\n")
				fmt.Printf("  Status:   %s\n", task.Status)
				fmt.Printf("  Duration: %s\n", task.Duration())
				if task.IsSuccess() && len(task.Result) > 0 {
					fmt.Printf("  Result:   %s\n", string(task.Result))
				}
				if task.Error != "" {
					fmt.Printf("  Error:    %s\n", task.Error)
				}
			}

			return nil
		},
	}

	sendCmd.Flags().StringVarP(&toAgent, "to", "t", "", "Target agent ID (optional, will auto-route if not specified)")
	sendCmd.Flags().StringVarP(&capability, "capability", "c", "", "Capability name (required)")
	sendCmd.Flags().StringVarP(&payloadStr, "payload", "p", "{}", "Task payload as JSON")
	sendCmd.Flags().StringVarP(&fromAgent, "from", "f", "cli", "Source agent ID")
	sendCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for task completion")
	sendCmd.MarkFlagRequired("capability")

	// Get task status
	statusCmd := &cobra.Command{
		Use:   "status <task-id>",
		Short: "Get task status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			task, err := client.GetTask(ctx, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Task: %s\n", task.ID)
			fmt.Printf("  From:       %s\n", task.FromAgent)
			fmt.Printf("  To:         %s\n", task.ToAgent)
			fmt.Printf("  Capability: %s\n", task.Capability)
			fmt.Printf("  Status:     %s\n", task.Status)
			fmt.Printf("  Created:    %s\n", task.CreatedAt.Format(time.RFC3339))
			fmt.Printf("  Updated:    %s\n", task.UpdatedAt.Format(time.RFC3339))
			if task.CompletedAt != nil {
				fmt.Printf("  Completed:  %s\n", task.CompletedAt.Format(time.RFC3339))
				fmt.Printf("  Duration:   %s\n", task.Duration())
			}
			if len(task.Payload) > 0 {
				fmt.Printf("  Payload:    %s\n", string(task.Payload))
			}
			if len(task.Result) > 0 {
				fmt.Printf("  Result:     %s\n", string(task.Result))
			}
			if task.Error != "" {
				fmt.Printf("  Error:      %s\n", task.Error)
			}

			return nil
		},
	}

	cmd.AddCommand(sendCmd, statusCmd)
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("A2A Platform CLI v1.0.0")
		},
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
