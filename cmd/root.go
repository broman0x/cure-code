package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/broman0x/cure-code/internal/agent"
	"github.com/broman0x/cure-code/internal/ai"
	"github.com/broman0x/cure-code/internal/config"
	"github.com/broman0x/cure-code/internal/ui"
	"github.com/broman0x/cure-code/internal/version"
)


var (
	cfgFile     string
	doInstall   bool
	doUninstall bool
	showVersion bool
	resumeID    string
	yoloMode    bool
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
	}

	for _, p := range apiKeyProviders {
		if os.Getenv(p.envKey) != "" {
			provider, err = ai.CreateFCProvider(p.name, p.model)
			if err == nil {
				color.Green("  [OK] Using %s (via %s)\n", p.name, p.envKey)
				a := agent.NewAgent(provider, mustGetwd())
				a.YOLO = yoloMode
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
		return a, nil
	}

	// [EN] Fall back to config's last provider only if no API keys available
	// [ID] Fallback ke last provider dari config hanya jika tidak ada API key
	if cfg.LastProvider != "" && cfg.LastModel != "" {
		provider, err = ai.CreateFCProvider(cfg.LastProvider, cfg.LastModel)
		if err == nil {
			a := agent.NewAgent(provider, mustGetwd())
			a.YOLO = yoloMode
			return a, nil
		}
		color.Yellow("  [!] Failed to load last provider (%s): %v. Falling back...\n", cfg.LastProvider, err)
	}

	return nil, fmt.Errorf("no AI provider available. Set GEMINI_API_KEY, OPENAI_API_KEY, or ensure Ollama is running")
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
		return runProviderSetup(err)
	}

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
	fmt.Printf("  %s %s\n", cActive("Tools:"), strings.Join(toolNames, ", "))
	fmt.Printf("  %s\n", cSubtle("Type your prompt, @ to tag files, or / for commands"))
	fmt.Println()

	completer := func(d prompt.Document) []prompt.Suggest {
		word := d.GetWordBeforeCursor()
		if strings.HasPrefix(word, "/") {
			return []prompt.Suggest{
				{Text: "/help", Description: "Show help message"},
				{Text: "/exit", Description: "Exit the agent"},
				{Text: "/clear", Description: "Clear the screen"},
				{Text: "/compact", Description: "Clear conversation history"},
				{Text: "/save", Description: "Save the current session"},
				{Text: "/resume", Description: "Resume a previous session"},
				{Text: "/model", Description: "Switch AI model"},
				{Text: "/usage", Description: "Show token usage stats"},
				{Text: "/version", Description: "Show version info"},
			}
		}
		if strings.HasPrefix(word, "@") {
			path := strings.TrimPrefix(word, "@")
			dir := "."
			if idx := strings.LastIndex(path, "/"); idx != -1 {
				dir = path[:idx]
			}
			fullDir := filepath.Join(ag.WorkDir, dir)
			entries, _ := os.ReadDir(fullDir)
			var suggestions []prompt.Suggest
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() {
					name += "/"
				}
				p := filepath.Join(dir, name)
				if dir == "." {
					p = name
				}
				if strings.HasPrefix(p, path) {
					suggestions = append(suggestions, prompt.Suggest{Text: "@" + p})
				}
			}
			return suggestions
		}
		return nil
	}

	executor := func(input string) {
		input = strings.TrimSpace(input)
		if input == "" {
			return
		}

		// [EN] Temporarily restore terminal to cooked mode so tools and SIGINT work properly
		// [ID] Kembalikan terminal ke mode normal sementara agar input tool dan SIGINT berfungsi
		cleanupTerminal()

		if strings.HasPrefix(input, "/") {
			if handleCommand(input, ag) {
				cleanupTerminal()
				os.Exit(0)
			}
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		
		// Stub for signal.Notify - simplified for build compatibility
		// In production this would use proper signal handling
		
		if err := ag.ProcessPrompt(ctx, input); err != nil {
			if err.Error() != "context canceled" {
				color.Red("\n  Error: %v\n", err)
			}
		}

		cancel()
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("  cure > "),
		prompt.OptionPrefixTextColor(prompt.Cyan),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionBGColor(prompt.Cyan),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.Black),
		prompt.OptionMaxSuggestion(10),
		prompt.OptionSuggestionTextColor(prompt.White),
	)

	p.Run()

	if len(ag.History) > 0 {
		configDir := filepath.Dir(config.GetConfigPath())
		id, err := agent.SaveSession(ag.History, ag.Tasks, ag.WorkDir, configDir)
		if err != nil {
			color.Red("  [!] Failed to auto-save session: %v", err)
		} else {
			color.HiBlack("  Session auto-saved as %s", id)
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

func runOneShot(prompt string, sessionID string) error {
	ag, err := createAgent()
	if err != nil {
		return err
	}
	
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
	fmt.Printf("  %s  %s\n", cCmd("/usage   "), cDesc("Show session token usage"))
	fmt.Printf("  %s  %s\n", cCmd("/save    "), cDesc("Save current session"))
	fmt.Printf("  %s  %s\n", cCmd("/resume  "), cDesc("Resume a saved session"))
	fmt.Printf("  %s  %s\n", cCmd("/ps      "), cDesc("List/stop background processes"))
	fmt.Printf("  %s  %s\n", cCmd("/version "), cDesc("Show version"))
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

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
