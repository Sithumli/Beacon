package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sithumli/Beacon/internal/broker"
	"github.com/Sithumli/Beacon/internal/health"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
	"github.com/Sithumli/Beacon/web"
	pb "github.com/Sithumli/Beacon/api/proto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack = "\033[40m"
)

var corsOrigins string

func main() {
	// Parse flags
	httpPort := flag.Int("port", 8080, "HTTP server port")
	grpcPort := flag.Int("grpc-port", 50051, "gRPC server port")
	dbPath := flag.String("db", "beacon.db", "SQLite database path")
	inMemory := flag.Bool("memory", false, "Use in-memory storage")
	debug := flag.Bool("debug", false, "Enable debug logging")
	enableGRPC := flag.Bool("grpc", true, "Enable gRPC server")
	flag.StringVar(&corsOrigins, "cors-origins", "", "Allowed CORS origins (comma-separated, empty for '*' in dev)")
	flag.Parse()

	// Configure logging - suppress during startup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel) // Quiet during startup
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Print banner first
	printBanner()

	// Initialize storage
	var dataStore store.Store
	var err error
	var storageType string
	if *inMemory {
		storageType = "memory"
		dataStore = store.NewMemoryStore()
	} else {
		storageType = *dbPath
		dataStore, err = store.NewSQLiteStore(*dbPath)
		if err != nil {
			fmt.Printf("  %s%s ERROR %s %sFailed to initialize SQLite: %v%s\n", Red, Bold, Reset, Dim, err, Reset)
			fmt.Printf("  %sTip: Use --memory flag for in-memory storage%s\n\n", Dim, Reset)
			os.Exit(1)
		}
	}
	defer dataStore.Close()

	// Initialize services
	registryService := registry.NewService(dataStore)
	brokerService := broker.NewService(dataStore, registryService)
	healthMonitor := health.NewMonitor(dataStore, health.DefaultConfig())

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register routes
	registryHandler := registry.NewHTTPHandler(registryService)
	registryHandler.RegisterRoutes(mux)

	brokerHandler := broker.NewHTTPHandler(brokerService)
	brokerHandler.RegisterRoutes(mux)

	webHandler := web.NewHandler(registryService, brokerService)
	webHandler.RegisterRoutes(mux)

	// Create HTTP server
	httpAddr := fmt.Sprintf(":%d", *httpPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("\n  %s%s ERROR %s HTTP server failed: %v%s\n", Red, Bold, Reset, err, Reset)
			os.Exit(1)
		}
	}()

	// Start gRPC server if enabled
	var grpcServer *grpc.Server
	if *enableGRPC {
		grpcAddr := fmt.Sprintf(":%d", *grpcPort)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			fmt.Printf("\n  %s%s ERROR %s Failed to start gRPC: %v%s\n", Red, Bold, Reset, err, Reset)
			os.Exit(1)
		}

		grpcServer = grpc.NewServer()
		registryGRPC := registry.NewGRPCServer(registryService)
		pb.RegisterRegistryServiceServer(grpcServer, registryGRPC)

		brokerGRPC := broker.NewGRPCServer(brokerService)
		pb.RegisterBrokerServiceServer(grpcServer, brokerGRPC)

		reflection.Register(grpcServer)

		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				fmt.Printf("\n  %s%s ERROR %s gRPC server failed: %v%s\n", Red, Bold, Reset, err, Reset)
			}
		}()
	}

	// Start health monitor
	ctx, cancel := context.WithCancel(context.Background())
	go healthMonitor.Start(ctx)

	// Print server info
	printServerInfo(*httpPort, *grpcPort, *enableGRPC, storageType)

	// Re-enable logging for runtime
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Printf("\n\n  %s%s Shutting down gracefully...%s\n", Dim, Bold, Reset)

	// Graceful shutdown
	cancel()

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	httpServer.Shutdown(shutdownCtx)

	fmt.Printf("  %s%s Beacon stopped.%s\n\n", Green, Bold, Reset)
}

func printBanner() {
	fmt.Print("\033[H\033[2J") // Clear screen

	banner := `
    %s%s╔══════════════════════════════════════════════════════════════╗%s
    %s%s║%s                                                              %s%s║%s
    %s%s║%s    %s%s██████╗ ███████╗ █████╗  ██████╗ ██████╗ ███╗   ██╗%s     %s%s║%s
    %s%s║%s    %s%s██╔══██╗██╔════╝██╔══██╗██╔════╝██╔═══██╗████╗  ██║%s     %s%s║%s
    %s%s║%s    %s%s██████╔╝█████╗  ███████║██║     ██║   ██║██╔██╗ ██║%s     %s%s║%s
    %s%s║%s    %s%s██╔══██╗██╔══╝  ██╔══██║██║     ██║   ██║██║╚██╗██║%s     %s%s║%s
    %s%s║%s    %s%s██████╔╝███████╗██║  ██║╚██████╗╚██████╔╝██║ ╚████║%s     %s%s║%s
    %s%s║%s    %s%s╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝%s     %s%s║%s
    %s%s║%s                                                              %s%s║%s
    %s%s║%s           %sAgent Discovery & Task Routing Platform%s           %s%s║%s
    %s%s║%s                                                              %s%s║%s
    %s%s╚══════════════════════════════════════════════════════════════╝%s

`
	fmt.Printf(banner,
		Cyan, Bold, Reset,
		Cyan, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Yellow, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Dim, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset, Cyan, Bold, Reset,
		Cyan, Bold, Reset,
	)
}

func printServerInfo(httpPort, grpcPort int, grpcEnabled bool, storage string) {
	// Status section
	fmt.Printf("  %s%s STATUS %s\n", Green, Bold, Reset)
	fmt.Printf("  %s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Dim, Reset)
	fmt.Printf("  %s●%s Server      %s%sRunning%s\n", Green, Reset, Green, Bold, Reset)
	fmt.Printf("  %s●%s Storage     %s%s\n", Green, Reset, storage, Reset)
	fmt.Printf("  %s●%s Health      %sMonitoring active%s\n\n", Green, Reset, Dim, Reset)

	// Endpoints section
	fmt.Printf("  %s%s ENDPOINTS %s\n", Blue, Bold, Reset)
	fmt.Printf("  %s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Dim, Reset)
	fmt.Printf("  %sDashboard%s    → %s%shttp://localhost:%d%s\n", White, Reset, Cyan, Bold, httpPort, Reset)
	fmt.Printf("  %sHTTP API%s     → %s%shttp://localhost:%d/api/v1%s\n", White, Reset, Cyan, Bold, httpPort, Reset)
	if grpcEnabled {
		fmt.Printf("  %sgRPC%s         → %s%slocalhost:%d%s\n", White, Reset, Cyan, Bold, grpcPort, Reset)
	}
	fmt.Println()

	// API Reference
	fmt.Printf("  %s%s API REFERENCE %s\n", Magenta, Bold, Reset)
	fmt.Printf("  %s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Dim, Reset)

	fmt.Printf("  %sAgents%s\n", Yellow, Reset)
	fmt.Printf("    %sPOST%s   /api/v1/agents      %sRegister agent%s\n", Green, Reset, Dim, Reset)
	fmt.Printf("    %sGET%s    /api/v1/agents      %sList all agents%s\n", Blue, Reset, Dim, Reset)
	fmt.Printf("    %sGET%s    /api/v1/agents/:id  %sGet agent details%s\n", Blue, Reset, Dim, Reset)
	fmt.Printf("    %sDELETE%s /api/v1/agents/:id  %sDeregister agent%s\n", Red, Reset, Dim, Reset)
	fmt.Printf("    %sPOST%s   /api/v1/heartbeat   %sSend heartbeat%s\n", Green, Reset, Dim, Reset)
	fmt.Printf("    %sGET%s    /api/v1/discover    %sFind by capability%s\n", Blue, Reset, Dim, Reset)
	fmt.Println()

	fmt.Printf("  %sTasks%s\n", Yellow, Reset)
	fmt.Printf("    %sPOST%s   /api/v1/tasks       %sCreate task%s\n", Green, Reset, Dim, Reset)
	fmt.Printf("    %sGET%s    /api/v1/tasks       %sList all tasks%s\n", Blue, Reset, Dim, Reset)
	fmt.Printf("    %sGET%s    /api/v1/tasks/:id   %sGet task details%s\n", Blue, Reset, Dim, Reset)
	fmt.Printf("    %sPATCH%s  /api/v1/tasks/:id   %sUpdate task status%s\n", Yellow, Reset, Dim, Reset)
	fmt.Printf("    %sPOST%s   /api/v1/route       %sRoute to capable agent%s\n", Green, Reset, Dim, Reset)
	fmt.Println()

	if grpcEnabled {
		fmt.Printf("  %sgRPC Services%s\n", Yellow, Reset)
		fmt.Printf("    %s•%s RegistryService  %sRegister, Discover, Heartbeat, Watch%s\n", Cyan, Reset, Dim, Reset)
		fmt.Printf("    %s•%s BrokerService    %sSendTask, RouteTask, Subscribe%s\n", Cyan, Reset, Dim, Reset)
		fmt.Println()
	}

	// Ready message
	fmt.Printf("  %s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Dim, Reset)
	fmt.Printf("  %s%s Ready! %sPress Ctrl+C to stop%s\n\n", Green, Bold, Dim, Reset)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestOrigin := r.Header.Get("Origin")
		if requestOrigin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// No CORS policy configured: do not allow cross-origin access by default
		if corsOrigins == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the request origin is in the allowed list
		allowed := false
		for _, o := range strings.Split(corsOrigins, ",") {
			if strings.TrimSpace(o) == requestOrigin {
				allowed = true
				break
			}
		}
		if !allowed {
			// Origin not allowed, don't set CORS headers
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Vary", "Origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
