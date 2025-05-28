package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
	"github.com/zinrai/x-scheduler/internal/executor"
	"github.com/zinrai/x-scheduler/internal/poster"
	"github.com/zinrai/x-scheduler/pkg/logger"
)

const (
	Version = "0.2.0"
)

func main() {
	var (
		executeFlag  = flag.Bool("execute", false, "Execute posts scheduled for today")
		validateFlag = flag.Bool("validate", false, "Validate configuration file")
		versionFlag  = flag.Bool("version", false, "Show version information")
		verboseFlag  = flag.Bool("verbose", false, "Enable verbose logging")
		helpFlag     = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	// Set log level
	if *verboseFlag {
		logger.SetLevel(logger.DEBUG)
	}

	// Handle version flag
	if *versionFlag {
		fmt.Printf("x-scheduler version %s\n", Version)
		os.Exit(0)
	}

	// Handle help flag
	if *helpFlag {
		showHelp()
		os.Exit(0)
	}

	// Validate flags and get config path
	configPath := validateFlagsAndGetConfigPath(*executeFlag, *validateFlag)

	// Execute the requested operation
	if err := runOperation(*executeFlag, *validateFlag, configPath); err != nil {
		logger.Fatal("Operation failed: %v", err)
	}
}

// Validates command line flags and returns config path
func validateFlagsAndGetConfigPath(execute, validate bool) string {
	// Count and validate active flags
	activeFlags := countActiveFlags(execute, validate)

	if activeFlags == 0 {
		fmt.Fprintf(os.Stderr, "Error: Must specify one of -execute or -validate\n")
		showUsage()
		os.Exit(1)
	}

	if activeFlags > 1 {
		fmt.Fprintf(os.Stderr, "Error: Only one operation flag can be specified\n")
		showUsage()
		os.Exit(1)
	}

	// Get and validate config file path
	return getConfigFilePath()
}

// Counts the number of active operation flags
func countActiveFlags(execute, validate bool) int {
	count := 0
	if execute {
		count++
	}
	if validate {
		count++
	}
	return count
}

// Gets and validates the config file path argument
func getConfigFilePath() string {
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: Config file path is required\n")
		showUsage()
		os.Exit(1)
	}
	return args[0]
}

func runOperation(execute, validate bool, configPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	switch {
	case validate:
		return runValidate(cfg, configPath)
	case execute:
		return runExecute(cfg, configPath)
	default:
		return fmt.Errorf("no operation specified")
	}
}

func runValidate(cfg *config.Config, configPath string) error {
	logger.Info("Validating configuration: %s", configPath)

	enabledPosts := cfg.GetEnabledPosts()
	futurePosts := executor.FilterFuturePosts(cfg.Posts, time.Now())

	fmt.Printf("Configuration validation successful\n")
	fmt.Printf("Total posts: %d\n", len(cfg.Posts))
	fmt.Printf("Enabled posts: %d\n", len(enabledPosts))
	fmt.Printf("Future posts for today: %d\n", len(futurePosts))

	// Check poster (xurl) availability
	if err := poster.Validate(); err != nil {
		fmt.Printf("Warning: Poster validation failed: %v\n", err)
		fmt.Printf("Make sure xurl is installed and configured properly\n")
	} else {
		fmt.Printf("Poster: xurl command available\n")
	}

	if len(futurePosts) > 0 {
		showUpcomingPosts(futurePosts)
	}

	return nil
}

// Displays upcoming posts information
func showUpcomingPosts(futurePosts []config.Post) {
	fmt.Printf("\nUpcoming posts for today:\n")
	for i, post := range futurePosts {
		if i >= 5 { // Show only first 5
			fmt.Printf("... and %d more\n", len(futurePosts)-5)
			break
		}
		if post.Test {
			fmt.Printf("  [TEST] %s\n", truncateContent(post.Content, 50))
		} else {
			fmt.Printf("  %s: %s\n",
				post.ScheduledAt.Format("15:04"),
				truncateContent(post.Content, 50))
		}
	}
}

func runExecute(cfg *config.Config, configPath string) error {
	logger.Info("Executing posts scheduled for today")

	// Create executor and execute posts
	exec := executor.NewExecutor()
	return exec.Execute(cfg)
}

func showHelp() {
	fmt.Printf("x-scheduler - X (Twitter) post scheduler\n\n")
	fmt.Printf("USAGE:\n")
	fmt.Printf("  x-scheduler [flags] <config.yaml>\n\n")
	fmt.Printf("FLAGS:\n")
	fmt.Printf("  -execute    Execute posts scheduled for today\n")
	fmt.Printf("  -validate   Validate configuration file\n")
	fmt.Printf("  -verbose    Enable verbose logging\n")
	fmt.Printf("  -version    Show version information\n")
	fmt.Printf("  -help       Show this help message\n\n")
	fmt.Printf("EXAMPLES:\n")
	fmt.Printf("  x-scheduler -validate config.yaml\n")
	fmt.Printf("  x-scheduler -execute config.yaml\n\n")
	fmt.Printf("SCHEDULING:\n")
	fmt.Printf("  Run daily via cron to process scheduled posts:\n")
	fmt.Printf("  0 0 * * * /usr/local/bin/x-scheduler -execute /path/to/config.yaml\n\n")
	fmt.Printf("REQUIREMENTS:\n")
	fmt.Printf("  xurl        X API command-line tool for OAuth 2.0 authentication\n")
	fmt.Printf("              Install from: https://github.com/xdevplatform/xurl\n")
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: x-scheduler [flags] <config.yaml>\n")
	fmt.Fprintf(os.Stderr, "Run 'x-scheduler -help' for more information.\n")
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}
