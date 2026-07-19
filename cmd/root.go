package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/broman0x/cure-code/internal/ui"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/broman0x/cure-code/internal/agent"
	"github.com/broman0x/cure-code/internal/ai"
	"github.com/broman0x/cure-code/internal/config"
	"github.com/broman0x/cure-code/internal/mcp"
	"github.com/broman0x/cure-code/internal/memory"
	"github.com/broman0x/cure-code/internal/tools"
	"github.com/broman0x/cure-code/internal/version"
)


var (
	cfgFile     string
	doInstall   bool
	doUninstall bool
	showVersion bool
	resumeID    string
	yoloMode    bool
	permissionMode string
	allowCommands  []string
	sandboxProfile string
	SkipPause   bool
)

var rootCmd = &cobra.Command{
	Use:   "curecode",
	Short: "AI Coding Agent by bromanprjkt",
	Args:  cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// [EN] Ensure config dirs exist for all commands
		// [ID] Pastikan direktori konfigurasi ada untuk semua perintah
		if err := config.EnsureConfigDirs(); err != nil {
			color.Yellow("  [!] Config dir error: %v", err)
		} else {
			color.Yellow("  [D] Config dirs ready")
		}

		if showVersion {
			SkipPause = true
			fmt.Printf("CuRe Code v%s\n", version.Version)
			fmt.Println(version.BuildName + " by " + version.Author)
			return nil
		}
		if doInstall {
			SkipPause = true
			return runSelfInstall()
		}
		if doUninstall {
			SkipPause = true
			return runSelfUninstall()
		}
		if resumeID != "" {
			if len(args) > 0 {
				SkipPause = true
				return runOneShot(strings.Join(args, " "), resumeID)
			}
			return runREPL(resumeID)
		}
		if len(args) > 0 {
			SkipPause = true
			return runOneShot(strings.Join(args, " "), "")
		}

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			SkipPause = true
			scanner := bufio.NewScanner(os.Stdin)
			var input []string
			for scanner.Scan() {
				input = append(input, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			if len(input) > 0 {
				return runOneShot(strings.Join(input, "\n"), "")
			}
		}

		return runREPL("")
	},
}

// [EN] Execute adds all child commands to the root command and sets flags appropriately.
// [ID] Execute menambahkan semua sub-perintah ke perintah akar dan mengatur flag dengan sesuai.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
	rootCmd.Flags().BoolVar(&doInstall, "install", false, "install to PATH")
	rootCmd.Flags().BoolVar(&doUninstall, "uninstall", false, "uninstall from PATH")
	rootCmd.Flags().StringVar(&resumeID, "resume", "", "resume a saved session by ID")
	rootCmd.Flags().BoolVar(&yoloMode, "yolo", false, "skip tool execution confirmations")
	rootCmd.Flags().StringVar(&permissionMode, "permission-mode", "default", "permission mode: default | accept_edits | bypass")
	rootCmd.Flags().StringArrayVar(&allowCommands, "allow-command", nil, "auto-allow run_command when command starts with this prefix (repeatable)")
	rootCmd.Flags().StringVar(&sandboxProfile, "sandbox-profile", "off", "shell sandbox profile: off | workspace-write | workspace-write-no-network | read-only")
	cobra.MousetrapHelpText = ""
}

func initConfig() {

	godotenv.Load()

	godotenv.Load(config.GetEnvPath())

	// [EN] Ensure config directories exist (creates ~/.config/curecode/ and sessions/)
	// [ID] Pastikan direktori konfigurasi ada
	_ = config.EnsureConfigDirs()
}

func createAgent() (*agent.Agent, error) {
	cfg := config.Load()

	var provider agent.FunctionCallingProvider
	var err error

	// [EN] Check API keys in env FIRST, before trying config's last provider
	// [ID] Periksa API key di env TERLEBIH DAHULU, sebelum mencoba last provider dari config
	apiKeyProviders := []struct {
		name     string
		model    string
		envKey   string
	}{
		{"gemini", "gemini-2.5-flash", "GEMINI_API_KEY"},
		{"openai", "gpt-4o-mini", "OPENAI_API_KEY"},
		{"claude", "claude-sonnet-4-20250514", "ANTHROPIC_API_KEY"},
		{"nvidia", "nvidia/nemotron-3-super-120b-a12b", "NVIDIA_API_KEY"},
		{"xai", "grok-2-1212", "XAI_API_KEY"},
		{"deepseek", "deepseek-coder", "DEEPSEEK_API_KEY"},
		{"openrouter", "anthropic/claude-3.5-sonnet", "OPENROUTER_API_KEY"},
		{"custom", "custom", "CUSTOM_API_URL"},
	}

	for _, p := range apiKeyProviders {
		if os.Getenv(p.envKey) != "" {
			modelToUse := p.model
			if p.name == "custom" {
				if envModel := os.Getenv("CUSTOM_MODEL"); envModel != "" {
					modelToUse = envModel
				} else if cfg.LastProvider == "custom" {
					modelToUse = cfg.LastModel
				}
			}
			provider, err = ai.CreateFCProvider(p.name, modelToUse)
			if err == nil {
				color.Green("  [OK] Using %s (via %s)\n", p.name, p.envKey)
				a := agent.NewAgent(provider, mustGetwd())
				a.YOLO = yoloMode
				if err := configureAgentRuntime(a); err != nil {
					return nil, err
				}
				return a, nil
			}
		}
	}

	// [EN] Try Ollama (local, no API key needed)
	// [ID] Coba Ollama (lokal, tidak perlu API key)
	provider, err = ai.CreateFCProvider("ollama", "llama3")
	if err == nil {
		a := agent.NewAgent(provider, mustGetwd())
		a.YOLO = yoloMode
		if err := configureAgentRuntime(a); err != nil {
			return nil, err
		}
		return a, nil
	}

	// [EN] Fall back to config's last provider only if no API keys available
	// [ID] Fallback ke last provider dari config hanya jika tidak ada API key
	if cfg.LastProvider != "" && cfg.LastModel != "" {
		provider, err = ai.CreateFCProvider(cfg.LastProvider, cfg.LastModel)
		if err == nil {
			a := agent.NewAgent(provider, mustGetwd())
			a.YOLO = yoloMode
			if err := configureAgentRuntime(a); err != nil {
				return nil, err
			}
			return a, nil
		}
		color.Yellow("  [!] Failed to load last provider (%s): %v. Falling back...\n", cfg.LastProvider, err)
	}

	return nil, fmt.Errorf("no AI provider available. Set GEMINI_API_KEY, OPENAI_API_KEY, or ensure Ollama is running")
}

func configureAgentRuntime(a *agent.Agent) error {
	mode := permissionMode
	if yoloMode {
		mode = "bypass"
	}
	if err := a.SetPermissionMode(mode); err != nil {
		return err
	}
	if err := a.SetShellSandboxProfile(sandboxProfile); err != nil {
		return err
	}
	for _, prefix := range allowCommands {
		if err := validateAllowCommandPrefix(prefix); err != nil {
			return err
		}
		a.AddAllowedCommandPrefix(prefix)
	}
	return nil
}

func validateAllowCommandPrefix(prefix string) error {
	p := strings.TrimSpace(strings.ToLower(prefix))
	if p == "" {
		return fmt.Errorf("allow-command prefix cannot be empty")
	}
	banned := []string{
		"sh -c", "bash -c", "zsh -c", "python", "python3", "node -e", "perl -e", "ruby -e",
	}
	for _, b := range banned {
		if p == b || strings.HasPrefix(p, b+" ") {
			return fmt.Errorf("allow-command prefix '%s' is too broad/risky", prefix)
		}
	}
	return nil
}

func setupMCP(ctx context.Context, a *agent.Agent, cfg *config.Config) {
	if len(cfg.MCPServers) == 0 {
		return
	}
	
	color.Cyan("  [MCP] Initializing servers...")
	for name, srvCfg := range cfg.MCPServers {
		if srvCfg.Command == "" {
			continue
		}
		
		client, err := mcp.NewClient(ctx, srvCfg.Command, srvCfg.Args...)
		if err != nil {
			color.Red("  [X] Failed to start MCP server '%s': %v", name, err)
			continue
		}
		
		a.CleanupFuncs = append(a.CleanupFuncs, client.Close)
		
		for _, t := range client.Tools {
			mcpTool := tools.NewMCPTool(client, name, t)
			a.Tools.RegisterDeferred(mcpTool)
		}
		color.Green("  [OK] MCP server '%s' initialized with %d tools", name, len(client.Tools))
	}
}

func cleanupTerminal() {
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "sane")
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

// [EN] runREPL starts the interactive Read-Eval-Print Loop for the AI agent.
// [ID] runREPL memulai loop interaktif Read-Eval-Print untuk agen AI.
func runREPL(sessionID string) error {
	defer cleanupTerminal()

	config.ResetCache()
	cfg := config.Load()

	if cfg.FirstRun {
		if err := runQuickSetup(); err != nil {
			return err
		}
		config.ResetCache()
	}

	ag, err := createAgent()
	if err != nil {
		if isRuntimeConfigError(err) {
			return err
		}
		return runProviderSetup(err)
	}
	
	setupMCP(context.Background(), ag, cfg)
	defer ag.Close()

	if sessionID != "" {
		configDir := filepath.Dir(config.GetConfigPath())
		history, tasks, err := agent.LoadSession(sessionID, configDir)
		if err != nil {
			color.Red("  Error loading session '%s': %v\n", sessionID, err)
		} else {
			ag.History = history
			ag.Tasks = tasks
			color.Green("  [OK] Resumed session: %s\n", sessionID)
		}
	}

	showBanner(ag)

	cActive := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()

	toolDefs := ag.Tools.Definitions()
	toolNames := make([]string, len(toolDefs))
	for i, t := range toolDefs {
		toolNames[i] = t.Name
	}

	fmt.Printf("  %s %s\n", cActive("Provider:"), ag.Provider.Name())
	fmt.Printf("  %s %d loaded\n", cActive("Tools:"), len(toolDefs))
	fmt.Printf("  %s %s | sandbox=%s\n", cActive("Mode:"), ag.PermissionMode, ag.ShellSandboxProfile)
	if len(ag.AllowedCommandPrefix) > 0 {
		fmt.Printf("  %s %s\n", cActive("Allowed command prefixes:"), strings.Join(ag.AllowedCommandPrefix, " | "))
	}
	fmt.Printf("  %s\n", cSubtle("Type your prompt, @ to tag files, or / for commands"))
	fmt.Println()

	executor := func(input string) {
		input = strings.TrimSpace(input)
		if input == "" {
			return
		}

		// [EN] Temporarily restore terminal to cooked mode so tools and SIGINT work properly
		// [ID] Kembalikan terminal ke mode normal sementara agar input tool dan SIGINT berfungsi
		cleanupTerminal()
		

		// Output is rendered by Bubbletea natively before exit

		if strings.HasPrefix(input, "/") {
			if handleCommand(input, ag) {
				ag.ProcMgr.Cleanup()
				ag.Close()
				cleanupTerminal()
				os.Exit(0)
			}
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		
		// [EN] Setup SIGINT (Ctrl+C) handler to cancel generation gracefully (simulating ESC stop)
		// [ID] Atur handler SIGINT (Ctrl+C) untuk membatalkan proses dengan aman (sebagai ganti ESC)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			select {
			case <-sigCh:
				cancel()
			case <-ctx.Done():
			}
		}()
		
		if err := ag.ProcessPrompt(ctx, input); err != nil {
			if err.Error() != "context canceled" {
				color.Red("\n  Error: %v\n", err)
			} else {
				color.HiBlack("\n  [!] Dibatalkan (Ctrl+C)\n")
			}
		}

		signal.Stop(sigCh)
		cancel()
	}

	// [EN] Main input loop using Bubbletea TUI instead of go-prompt
	for {
		res, err := ui.RunPrompt()
		if err != nil {
			color.Red("  [!] UI Error: %v\n", err)
			break
		}
		if res.Canceled {
			fmt.Println()
			color.HiBlack("  Goodbye!")
			fmt.Println()
			break
		}
		
		executor(res.Text)
	}

	if len(ag.History) > 0 {
		configDir := filepath.Dir(config.GetConfigPath())
		id, err := agent.SaveSession(ag.History, ag.Tasks, ag.WorkDir, configDir)
		if err != nil {
			color.Red("  [!] Failed to auto-save session: %v", err)
		} else {
			fmt.Println()
			color.HiBlack("  Session auto-saved. To resume, run:")
			color.HiWhite("  curecode --resume %s", id)
			fmt.Println()
		}
	}
	ag.ProcMgr.Cleanup()

	return nil
}

// [EN] handleCommand processes slash commands (e.g., /help, /exit) in the REPL.
// [ID] handleCommand memproses perintah slash (misal: /help, /exit) di REPL.
func handleCommand(input string, ag *agent.Agent) bool {
	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/exit", "/quit", "/q":
		if len(ag.History) > 0 {
			configDir := filepath.Dir(config.GetConfigPath())
			if id, err := agent.SaveSession(ag.History, ag.Tasks, ag.WorkDir, configDir); err == nil {
				fmt.Println()
				color.HiBlack("  Session auto-saved. To resume, run:")
				color.HiWhite("  curecode --resume %s", id)
			}
		}
		fmt.Println()
		color.HiBlack("  Goodbye!")
		fmt.Println()
		return true

	case "/clear", "/cls":
		fmt.Print("\033[H\033[2J")
		showBanner(ag)

	case "/help", "/h":
		showHelp()

	case "/model":
		handleModelSwitch(ag)

	case "/mode":
		handlePermissionMode(parts, ag)

	case "/sandbox":
		handleSandboxProfile(parts, ag)

	case "/mcp":
		handleMCP(parts, ag)

	case "/learn":
		handleLearn(input, parts)

	case "/grill-me":
		ag.InterviewMode = !ag.InterviewMode
		if ag.InterviewMode {
			color.HiCyan("  [Mode] Interview mode activated. The agent will grill you with questions to clarify your requirements.")
		} else {
			color.HiCyan("  [Mode] Interview mode deactivated.")
		}
		fmt.Println()

	case "/allowcmd":
		handleAllowCommandPrefix(input, ag)

	case "/compact":
		ag.ClearHistory()
		color.Green("  [OK] Conversation history cleared")
		fmt.Println()

	case "/save":
		configDir := filepath.Dir(config.GetConfigPath())
		id, err := agent.SaveSession(ag.History, ag.Tasks, ag.WorkDir, configDir)
		if err != nil {
			color.Red("  Error saving session: %v\n", err)
		} else {
			color.Green("  [OK] Session saved as: %s\n", id)
		}
		fmt.Println()

	case "/resume":
		handleResume(ag)

	case "/ps":
		handleProcesses(ag)

	case "/usage":
		color.HiCyan("\n  %s\n\n", ag.UsageSummary())

	case "/version":
		fmt.Printf("  CuRe Code v%s\n", version.GetVersion())
		fmt.Printf("  Architecture: Agentic Memory v1.0\n\n")

	case "/doctor":
		runDoctor(ag)

	default:
		color.Yellow("  Unknown command: %s (type /help for commands)\n\n", cmd)
	}

	return false
}

func handleResume(ag *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)
	configDir := filepath.Dir(config.GetConfigPath())
	sessions, err := agent.ListSessions(configDir)
	if err != nil {
		color.Red("  Error listing sessions: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		color.Yellow("  No saved sessions found.\n")
		return
	}

	ui.PrintHeader("RESUME SESSION")
	for i, s := range sessions {
		fmt.Printf("  %d. %s\n", i+1, s)
	}
	fmt.Println("  0. Cancel")
	fmt.Print("\n  Select > ")

	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	if choice == "0" || choice == "" {
		return
	}

	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx > 0 && idx <= len(sessions) {
		sessionID := sessions[idx-1]
		history, tasks, err := agent.LoadSession(sessionID, configDir)
		if err != nil {
			color.Red("  Error loading session: %v\n", err)
			return
		}
		ag.History = history
		ag.Tasks = tasks
		color.Green("  [OK] Resumed session: %s (%d messages)\n\n", sessionID, len(history))
	}
}

func handleProcesses(ag *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)
	procs := ag.ProcMgr.List()
	if len(procs) == 0 {
		color.Yellow("  No active background processes.\n")
		return
	}

	ui.PrintHeader("BACKGROUND PROCESSES")
	for _, p := range procs {
		fmt.Printf("  %d. [%d] %s\n", p.ID, p.Cmd.Process.Pid, p.Command)
	}
	fmt.Println("  0. Back")
	fmt.Print("\n  To stop a process, enter ID (or 0 to cancel) > ")

	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	if choice == "0" || choice == "" {
		return
	}

	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx > 0 {
		err := ag.ProcMgr.Stop(idx)
		if err != nil {
			color.Red("  Error stopping process: %v\n", err)
		} else {
			color.Green("  [OK] Process %d terminated.\n", idx)
		}
	}
}

func handlePermissionMode(parts []string, ag *agent.Agent) {
	if len(parts) == 1 {
		color.HiCyan("  Permission mode: %s\n\n", ag.PermissionMode)
		return
	}
	mode := strings.TrimSpace(parts[1])
	if err := ag.SetPermissionMode(mode); err != nil {
		color.Red("  %v\n\n", err)
		return
	}
	color.Green("  [OK] Permission mode set to: %s\n\n", ag.PermissionMode)
}

func handleSandboxProfile(parts []string, ag *agent.Agent) {
	if len(parts) == 1 {
		color.HiCyan("  Sandbox profile: %s\n\n", ag.ShellSandboxProfile)
		return
	}
	profile := strings.TrimSpace(parts[1])
	if err := ag.SetShellSandboxProfile(profile); err != nil {
		color.Red("  %v\n\n", err)
		return
	}
	color.Green("  [OK] Sandbox profile set to: %s\n\n", ag.ShellSandboxProfile)
}

func handleAllowCommandPrefix(input string, ag *agent.Agent) {
	prefix := strings.TrimSpace(strings.TrimPrefix(input, "/allowcmd"))
	if prefix == "" {
		if len(ag.AllowedCommandPrefix) == 0 {
			color.HiBlack("  No allowed command prefixes configured.\n\n")
			return
		}
		color.HiCyan("  Allowed command prefixes:")
		for _, p := range ag.AllowedCommandPrefix {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
		return
	}
	if err := validateAllowCommandPrefix(prefix); err != nil {
		color.Red("  %v\n\n", err)
		return
	}
	ag.AddAllowedCommandPrefix(prefix)
	color.Green("  [OK] Added allowed command prefix: %s\n\n", prefix)
}

func handleMCP(parts []string, ag *agent.Agent) {
	cfg := config.Load()

	if len(parts) < 2 {
		color.HiYellow("  Usage: /mcp [list | add | remove]")
		color.HiBlack("    /mcp list")
		color.HiBlack("    /mcp add <name> <command> [args...]")
		color.HiBlack("    /mcp remove <name>")
		fmt.Println()
		return
	}

	action := strings.ToLower(parts[1])
	switch action {
	case "list":
		if len(cfg.MCPServers) == 0 {
			color.Yellow("  No MCP servers configured.")
		} else {
			color.Cyan("  Configured MCP Servers:")
			for name, srv := range cfg.MCPServers {
				args := strings.Join(srv.Args, " ")
				color.HiWhite("    - %s: %s %s", name, srv.Command, args)
			}
		}
		fmt.Println()

	case "add":
		if len(parts) < 4 {
			color.Red("  Error: Missing arguments for /mcp add. Usage: /mcp add <name> <command> [args...]")
			return
		}
		name := parts[2]
		cmd := parts[3]
		args := parts[4:]

		if cfg.MCPServers == nil {
			cfg.MCPServers = make(map[string]config.MCPServerConfig)
		}

		cfg.MCPServers[name] = config.MCPServerConfig{
			Command: cmd,
			Args:    args,
		}

		if err := config.Save(cfg); err != nil {
			color.Red("  [X] Failed to save config: %v", err)
			return
		}

		color.Green("  [OK] Added MCP server '%s'. Initializing...", name)
		
		// Initialize it on the fly
		ctx := context.Background()
		client, err := mcp.NewClient(ctx, cmd, args...)
		if err != nil {
			color.Red("  [X] Failed to start MCP server '%s': %v", name, err)
			return
		}

		ag.CleanupFuncs = append(ag.CleanupFuncs, client.Close)

		for _, t := range client.Tools {
			mcpTool := tools.NewMCPTool(client, name, t)
			ag.Tools.RegisterDeferred(mcpTool)
		}
		color.Green("  [OK] MCP server '%s' initialized with %d tools", name, len(client.Tools))
		fmt.Println()

	case "remove":
		if len(parts) < 3 {
			color.Red("  Error: Missing name for /mcp remove. Usage: /mcp remove <name>")
			return
		}
		name := parts[2]

		if cfg.MCPServers == nil || cfg.MCPServers[name].Command == "" {
			color.Yellow("  MCP server '%s' not found.", name)
			return
		}

		delete(cfg.MCPServers, name)
		if err := config.Save(cfg); err != nil {
			color.Red("  [X] Failed to save config: %v", err)
			return
		}

		color.Green("  [OK] Removed MCP server '%s'. Note: Please restart CuRe Code to completely remove its tools.", name)
		fmt.Println()

	default:
		color.Red("  Unknown action: %s. Valid actions: list, add, remove", action)
		fmt.Println()
	}
}

func handleLearn(input string, parts []string) {
	if len(parts) < 2 {
		store, _ := memory.Load()
		if len(store.Rules) == 0 {
			color.Yellow("  No rules learned yet.")
		} else {
			color.Cyan("  Learned Rules:")
			for i, rule := range store.Rules {
				color.HiWhite("    %d. %s", i+1, rule)
			}
			color.HiBlack("  (Use /learn clear to remove all rules)")
		}
		fmt.Println()
		return
	}
	
	if parts[1] == "clear" {
		memory.ClearRules()
		color.Green("  [OK] Cleared all learned rules.")
		fmt.Println()
		return
	}

	rule := strings.TrimSpace(input[len(parts[0]):])
	if err := memory.AddRule(rule); err != nil {
		color.Red("  [X] Failed to save rule: %v", err)
	} else {
		color.Green("  [OK] Learned new rule: %s", rule)
	}
	fmt.Println()
}

func runOneShot(prompt string, sessionID string) error {
	ag, err := createAgent()
	if err != nil {
		return err
	}
	
	cfg := config.Load()
	setupMCP(context.Background(), ag, cfg)
	defer ag.Close()
	
	if sessionID != "" {
		configDir := filepath.Dir(config.GetConfigPath())
		history, tasks, err := agent.LoadSession(sessionID, configDir)
		if err == nil {
			ag.History = history
			ag.Tasks = tasks
			color.Green("  [OK] Resumed session: %s\n", sessionID)
		}
	}

	prompt = strings.TrimSpace(prompt)
	if strings.HasPrefix(prompt, "/") {
		handleCommand(prompt, ag)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	return ag.ProcessPrompt(ctx, prompt)
}

func showBanner(ag *agent.Agent) {
	ui.ShowStartupBanner(version.Version)

	cVer := color.New(color.FgHiBlack).SprintFunc()
	cWarn := color.New(color.FgHiYellow).SprintFunc()

	if ag != nil {
		ws := agent.DetectWorkspace(ag.WorkDir)
		if !ws.HasGit {
			fmt.Printf("  %s\n", cWarn("[!] Not in a git repository. Changes cannot be easily undone."))
		} else if ws.GitDirtyCount > 0 {
			fmt.Printf("  %s %s\n", cVer("Git:"), cWarn(fmt.Sprintf("%s (%d uncommitted changes)", ws.GitBranch, ws.GitDirtyCount)))
		} else {
			fmt.Printf("  %s %s\n", cVer("Git:"), color.HiGreenString(ws.GitBranch))
		}
	}
	fmt.Println()
}

func showHelp() {
	cCmd := color.New(color.FgCyan).SprintFunc()
	cDesc := color.New(color.FgWhite).SprintFunc()
	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s\n", cTitle("Commands"))
	fmt.Printf("  %s  %s\n", cCmd("/help    "), cDesc("Show this help"))
	fmt.Printf("  %s  %s\n", cCmd("/clear   "), cDesc("Clear screen"))
	fmt.Printf("  %s  %s\n", cCmd("/compact "), cDesc("Clear conversation history"))
	fmt.Printf("  %s  %s\n", cCmd("/model   "), cDesc("Switch AI provider/model"))
	fmt.Printf("  %s  %s\n", cCmd("/mode    "), cDesc("Show/set permission mode"))
	fmt.Printf("  %s  %s\n", cCmd("/sandbox "), cDesc("Show/set shell sandbox profile"))
	fmt.Printf("  %s  %s\n", cCmd("/mcp     "), cDesc("Manage MCP servers (list/add/remove)"))
	fmt.Printf("  %s  %s\n", cCmd("/learn   "), cDesc("Teach the agent a new persistent rule"))
	fmt.Printf("  %s  %s\n", cCmd("/grill-me"), cDesc("Activate interactive interview mode"))
	fmt.Printf("  %s  %s\n", cCmd("/allowcmd"), cDesc("Add allowed shell command prefix"))
	fmt.Printf("  %s  %s\n", cCmd("/usage   "), cDesc("Show session token usage"))
	fmt.Printf("  %s  %s\n", cCmd("/save    "), cDesc("Save current session"))
	fmt.Printf("  %s  %s\n", cCmd("/resume  "), cDesc("Resume a saved session"))
	fmt.Printf("  %s  %s\n", cCmd("/ps      "), cDesc("List/stop background processes"))
	fmt.Printf("  %s  %s\n", cCmd("/version "), cDesc("Show version"))
	fmt.Printf("  %s  %s\n", cCmd("/doctor  "), cDesc("Run environment diagnostics"))
	fmt.Printf("  %s  %s\n", cCmd("/exit    "), cDesc("Exit CuRe Code"))
	fmt.Println()
}

func handleModelSwitch(ag *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)
	ui.PrintHeader("SWITCH MODEL")
	fmt.Println("  1. Gemini 2.5 Flash")
	fmt.Println("  2. Gemini 2.5 Pro")
	fmt.Println("  3. GPT-4o Mini")
	fmt.Println("  4. GPT-4o")
	fmt.Println("  5. Claude Sonnet 4")
	fmt.Println("  6. NVIDIA Nemotron (Reasoning)")
	fmt.Println("  7. xAI Grok-2")
	fmt.Println("  8. DeepSeek Coder")
	fmt.Println("  10. Together Llama 3.1 70B")
	fmt.Println("  11. Mistral Large")
	fmt.Println("  12. Ollama (Local)")
	fmt.Println("  13. Custom Provider (OpenAI Compatible)")
	fmt.Println("  0. Cancel")
	fmt.Print("\n  Select > ")

	if !scanner.Scan() {
		return
	}

	providers := map[string]struct {
		pType string
		model string
		url   string
	}{
		"1":  {"gemini", "gemini-2.5-flash", "https://aistudio.google.com/app/apikey"},
		"2":  {"gemini", "gemini-2.5-pro", "https://aistudio.google.com/app/apikey"},
		"3":  {"openai", "gpt-4o-mini", "https://platform.openai.com/api-keys"},
		"4":  {"openai", "gpt-4o", "https://platform.openai.com/api-keys"},
		"5":  {"claude", "claude-sonnet-4-20250514", "https://console.anthropic.com/settings/keys"},
		"6":  {"nvidia", "nvidia/nemotron-3-super-120b-a12b", "https://build.nvidia.com/explore/discover"},
		"7":  {"xai", "grok-2-1212", "https://console.x.ai/"},
		"8":  {"deepseek", "deepseek-coder", "https://platform.deepseek.com/api_keys"},
		"9":  {"together", "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo", "https://api.together.xyz/v1"},
		"10": {"mistral", "mistral-large-latest", "https://api.mistral.ai/v1"},
	}

	choice := strings.TrimSpace(scanner.Text())

	if choice == "11" {
		fmt.Print("  Model name (e.g., llama3): ")
		if scanner.Scan() {
			model := strings.TrimSpace(scanner.Text())
			if model == "" {
				model = "llama3"
			}
			p, err := ai.CreateFCProvider("ollama", model)
			if err != nil {
				color.Red("  Error: %v", err)
				return
			}
			ag.Provider = p
			ag.ClearHistory()
			color.Green("  [OK] Switched to %s\n\n", p.Name())
			config.SaveLastModel("ollama", model)
		}
		return
	}

	if choice == "13" {
		fmt.Print("  API Base URL (e.g., http://localhost:11434/v1): ")
		if scanner.Scan() {
			baseURL := strings.TrimSpace(scanner.Text())
			if baseURL == "" {
				return
			}
			fmt.Print("  API Key (leave blank if none): ")
			var key string
			if scanner.Scan() {
				key = strings.TrimSpace(scanner.Text())
			}
			fmt.Print("  Model Name: ")
			var model string
			if scanner.Scan() {
				model = strings.TrimSpace(scanner.Text())
			}
			if model == "" {
				model = "default"
			}
			
			fmt.Print("  Testing connection... ")
			if err := testCustomConnection(baseURL, key, model); err != nil {
				color.Red("Failed: %v\n", err)
				return
			}
			color.Green("OK!\n")
			
			config.SaveAPIKey("CUSTOM_API_URL", baseURL)
			if key != "" {
				config.SaveAPIKey("CUSTOM_API_KEY", key)
			}
			config.SaveAPIKey("CUSTOM_MODEL", model)
			godotenv.Load(config.GetEnvPath())
			
			p, err := ai.CreateFCProvider("custom", model)
			if err != nil {
				color.Red("  Error: %v", err)
				return
			}
			ag.Provider = p
			ag.ClearHistory()
			color.Green("  [OK] Switched to %s\n\n", p.Name())
			config.SaveLastModel("custom", model)
		}
		return
	}

	if info, ok := providers[choice]; ok {
		p, err := ai.CreateFCProvider(info.pType, info.model)
		if err != nil {
			color.Red("  Error: %v\n", err)
			color.Yellow("  Get key here: %s\n", info.url)
			fmt.Print("  Paste API Key to set it now (or leave empty): ")
			if scanner.Scan() {
				key := strings.TrimSpace(scanner.Text())
				if key != "" {
					envKey := strings.ToUpper(info.pType) + "_API_KEY"
					if info.pType == "claude" {
						envKey = "ANTHROPIC_API_KEY"
					}
					config.SaveAPIKey(envKey, key)
					godotenv.Load()
					godotenv.Load(config.GetEnvPath())

					p, err = ai.CreateFCProvider(info.pType, info.model)
					if err == nil {
						ag.Provider = p
						ag.ClearHistory()
						color.Green("  [OK] API key saved and switched to %s\n\n", p.Name())
						config.SaveLastModel(info.pType, info.model)
						return
					}
				}
			}
			return
		}
		ag.Provider = p
		ag.ClearHistory()
		color.Green("  [OK] Switched to %s\n\n", p.Name())
		config.SaveLastModel(info.pType, info.model)
	}
}

func getInteractiveScanner() (*bufio.Scanner, func()) {
	if runtime.GOOS == "windows" {
		f, err := os.OpenFile("CONIN$", os.O_RDWR, 0644)
		if err == nil {
			return bufio.NewScanner(f), func() { f.Close() }
		}
	} else {
		f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err == nil {
			return bufio.NewScanner(f), func() { f.Close() }
		}
	}
	return bufio.NewScanner(os.Stdin), func() {}
}

func testCustomConnection(baseURL, apiKey, modelName string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// First, try /models
	modelsURL := strings.TrimSuffix(baseURL, "/") + "/models"
	req, _ := http.NewRequest("GET", modelsURL, nil)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
	}

	// Fallback to /chat/completions
	chatURL := strings.TrimSuffix(baseURL, "/") + "/chat/completions"
	reqBody := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 5,
	}
	payload, _ := json.Marshal(reqBody)
	
	req, _ = http.NewRequest("POST", chatURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	// For 4xx errors, try to parse JSON error to see if it's a valid API rejecting us
	body, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Error interface{} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
		return fmt.Errorf("API reached but rejected request (status %d): %s", resp.StatusCode, string(body))
	}
	
	return fmt.Errorf("server returned status: %d", resp.StatusCode)
}

func runQuickSetup() error {
	fmt.Print("\033[H\033[2J")

	ui.ShowStartupBanner(version.Version)
	color.HiCyan("  Welcome to CuRe Code Setup!")
	fmt.Println()

	scanner, cleanup := getInteractiveScanner()
	defer cleanup()
	
	fmt.Println("  Set up your AI provider:")
	fmt.Println("  1. Google Gemini (recommended)")
	fmt.Println("  2. OpenAI")
	fmt.Println("  3. Anthropic Claude")
	fmt.Println("  4. NVIDIA NIM")
	fmt.Println("  5. xAI Grok")
	fmt.Println("  6. DeepSeek")
	fmt.Println("  7. Ollama (local, no API key)")
	fmt.Println("  8. Custom Provider (OpenAI Compatible)")
	fmt.Print("\n  Select > ")

	if !scanner.Scan() {
		return fmt.Errorf("setup cancelled")
	}

	choice := strings.TrimSpace(scanner.Text())
	switch choice {
	case "1":
		fmt.Println("\n  Get key: https://aistudio.google.com/app/apikey")
		fmt.Print("  Paste API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("GEMINI_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("gemini", "gemini-2.5-flash")
			}
		}
	case "2":
		fmt.Println("\n  Get key: https://platform.openai.com/api-keys")
		fmt.Print("  Paste API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("OPENAI_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("openai", "gpt-4o-mini")
			}
		}
	case "3":
		fmt.Println("\n  Get key: https://console.anthropic.com/settings/keys")
		fmt.Print("  Paste API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("ANTHROPIC_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("claude", "claude-sonnet-4-20250514")
			}
		}
	case "4":
		fmt.Println("\n  Get key: https://build.nvidia.com/explore/discover")
		fmt.Print("  Paste NVIDIA API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("NVIDIA_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("nvidia", "nvidia/nemotron-3-super-120b-a12b")
			}
		}
	case "5":
		fmt.Println("\n  Get key: https://console.x.ai/")
		fmt.Print("  Paste xAI API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("XAI_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("xai", "grok-2-1212")
			}
		}
	case "6":
		fmt.Println("\n  Get key: https://platform.deepseek.com/api_keys")
		fmt.Print("  Paste DeepSeek API Key: ")
		if scanner.Scan() {
			key := strings.TrimSpace(scanner.Text())
			if key != "" {
				config.SaveAPIKey("DEEPSEEK_API_KEY", key)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("deepseek", "deepseek-coder")
			}
		}
	case "7":
		color.Yellow("  Make sure Ollama is running (https://ollama.com)")
		config.SaveLastModel("ollama", "llama3")
	case "8":
		fmt.Print("\n  API Base URL (e.g., http://localhost:11434/v1): ")
		if scanner.Scan() {
			baseURL := strings.TrimSpace(scanner.Text())
			if baseURL != "" {
				fmt.Print("  API Key (leave blank if none): ")
				var key string
				if scanner.Scan() {
					key = strings.TrimSpace(scanner.Text())
				}
				fmt.Print("  Model Name: ")
				var model string
				if scanner.Scan() {
					model = strings.TrimSpace(scanner.Text())
				}
				if model == "" {
					model = "default"
				}
				
				fmt.Print("  Testing connection... ")
				if err := testCustomConnection(baseURL, key, model); err != nil {
					color.Red("Failed: %v\n", err)
					return fmt.Errorf("connection failed")
				}
				color.Green("OK!\n")
				
				config.SaveAPIKey("CUSTOM_API_URL", baseURL)
				if key != "" {
					config.SaveAPIKey("CUSTOM_API_KEY", key)
				}
				config.SaveAPIKey("CUSTOM_MODEL", model)
				godotenv.Load(config.GetEnvPath())
				config.SaveLastModel("custom", model)
			}
		}
	}

	config.SaveFirstRun(false)
	color.Green("\n  [OK] Setup complete!")
	time.Sleep(1 * time.Second)
	fmt.Print("\033[H\033[2J")
	return nil
}

func runProviderSetup(originalErr error) error {
	color.Yellow("\n  [!] AI Provider configuration error")
	fmt.Printf("  Error: %v\n\n", originalErr)

	fmt.Print("  Do you want to re-configure now? [Y/n]: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() && strings.ToLower(strings.TrimSpace(scanner.Text())) != "n" {
		if err := runQuickSetup(); err == nil {
			return runREPL("")
		}
	}

	fmt.Println("\n  Please set an API key (export or .env):")
	fmt.Println("  GEMINI_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY,")
	fmt.Println("  NVIDIA_API_KEY, GROQ_API_KEY, DEEPSEEK_API_KEY")
	return originalErr
}

func isRuntimeConfigError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid permission mode") ||
		strings.Contains(msg, "invalid sandbox profile") ||
		strings.Contains(msg, "allow-command prefix")
}

func runDoctor(ag *agent.Agent) {
	configPath := config.GetConfigPath()
	configDir := filepath.Dir(configPath)
	sessionDir := filepath.Join(configDir, "sessions")

	ws := agent.DetectWorkspace(ag.WorkDir)
	sessions, _ := agent.ListSessions(configDir)
	_, goErr := exec.LookPath("go")
	goInPath := goErr == nil

	providerKeys := []string{
		"GEMINI_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY",
		"NVIDIA_API_KEY", "XAI_API_KEY", "DEEPSEEK_API_KEY", "OPENROUTER_API_KEY",
	}
	availableKeys := make([]string, 0)
	for _, k := range providerKeys {
		if os.Getenv(k) != "" {
			availableKeys = append(availableKeys, k)
		}
	}

	coreTools := ag.Tools.CoreDefinitions()
	deferredTools := ag.Tools.DeferredDefinitions()

	color.HiCyan("\n  Diagnostics")
	fmt.Printf("  - Version: %s\n", version.GetVersion())
	fmt.Printf("  - Provider: %s\n", ag.Provider.Name())
	fmt.Printf("  - Permission mode: %s\n", ag.PermissionMode)
	fmt.Printf("  - Shell sandbox: %s\n", ag.ShellSandboxProfile)
	fmt.Printf("  - Allowed command prefixes: %d\n", len(ag.AllowedCommandPrefix))
	fmt.Printf("  - Config path: %s\n", configPath)
	fmt.Printf("  - Sessions dir: %s (%d sessions)\n", sessionDir, len(sessions))
	fmt.Printf("  - Workspace: %s\n", ag.WorkDir)
	if ws.HasGit {
		fmt.Printf("  - Git: %s (%d uncommitted)\n", ws.GitBranch, ws.GitDirtyCount)
	} else {
		fmt.Printf("  - Git: not a repository\n")
	}
	if goInPath {
		fmt.Printf("  - Go in PATH: yes\n")
	} else {
		fmt.Printf("  - Go in PATH: no\n")
	}
	if len(availableKeys) > 0 {
		fmt.Printf("  - API keys detected: %s\n", strings.Join(availableKeys, ", "))
	} else {
		fmt.Printf("  - API keys detected: none\n")
	}
	fmt.Printf("  - Tools: %d core, %d deferred\n\n", len(coreTools), len(deferredTools))
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
