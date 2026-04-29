package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func runMCPInstall(args []string) error {
	// Parse flags manually
	var (
		client string
		force  bool
		dryRun bool
		list   bool
	)

	client = "claude-desktop" // default

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--client":
			if i+1 >= len(args) {
				return fmt.Errorf("--client requires a value")
			}
			i++
			client = args[i]
		case "--force":
			force = true
		case "--dry-run":
			dryRun = true
		case "--list":
			list = true
		case "-h", "--help":
			printMCPInstallHelp()
			return nil
		}
	}

	if list {
		fmt.Println("Supported MCP clients:")
		fmt.Println("  claude-desktop  Claude Desktop (default)")
		fmt.Println("  claude-code     Claude Code")
		fmt.Println("  cursor          Cursor")
		fmt.Println("  windsurf        Windsurf")
		fmt.Println("  codex           OpenAI Codex CLI")
		fmt.Println("  vscode          VS Code Copilot")
		fmt.Println("  gemini          Gemini CLI")
		fmt.Println("  amazon-q        Amazon Q Developer")
		fmt.Println("  zed             Zed")
		fmt.Println("  lm-studio       LM Studio")
		return nil
	}

	return runInstall(client, force, dryRun)
}

func printMCPInstallHelp() {
	fmt.Fprintf(os.Stderr, `Usage: trading-cli mcp install [flags]

Install trading-cli as an MCP server in your AI client's config.

Flags:
  --client <name>   Target client (default: claude-desktop)
  --force           Overwrite existing entry
  --dry-run         Print planned change without writing
  --list            Show all supported clients

Clients:
  claude-desktop  Claude Desktop (default)
  claude-code     Claude Code
  cursor          Cursor
  windsurf        Windsurf
  codex           OpenAI Codex CLI
  vscode          VS Code Copilot
  gemini          Gemini CLI
  amazon-q        Amazon Q Developer
  zed             Zed
  lm-studio       LM Studio

Examples:
  trading-cli mcp install
  trading-cli mcp install --client cursor
  trading-cli mcp install --client claude-code --force
  trading-cli mcp install --dry-run
`)
}

func clientConfigPath(client string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	switch strings.ToLower(client) {
	case "claude-desktop", "claude":
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
		case "linux":
			return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
		case "windows":
			appdata := os.Getenv("APPDATA")
			if appdata == "" {
				appdata = filepath.Join(home, "AppData", "Roaming")
			}
			return filepath.Join(appdata, "Claude", "claude_desktop_config.json"), nil
		}
	case "cursor":
		return filepath.Join(home, ".cursor", "mcp.json"), nil
	case "claude-code":
		return filepath.Join(home, ".claude.json"), nil
	case "windsurf":
		return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"), nil
	case "codex":
		return filepath.Join(home, ".codex", "config.toml"), nil
	case "vscode", "vs-code", "copilot":
		return filepath.Join(".vscode", "mcp.json"), nil
	case "gemini":
		return filepath.Join(home, ".gemini", "settings.json"), nil
	case "amazon-q", "q":
		return filepath.Join(home, ".aws", "amazonq", "mcp.json"), nil
	case "zed":
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Application Support", "Zed", "settings.json"), nil
		default:
			return filepath.Join(home, ".config", "zed", "settings.json"), nil
		}
	case "lm-studio":
		return filepath.Join(home, ".lm-studio", "mcp.json"), nil
	}
	return "", fmt.Errorf("unknown client %q\nUse --list to see supported clients", client)
}

func binaryPath() (string, error) {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs, nil
		}
	}
	if path, err := exec.LookPath("trading-cli"); err == nil {
		return filepath.Abs(path)
	}
	return "", fmt.Errorf("cannot locate trading-cli binary")
}

func mcpConfigKey(client string) string {
	switch strings.ToLower(client) {
	case "vscode", "vs-code", "copilot":
		return "servers"
	case "zed":
		return "context_servers"
	default:
		return "mcpServers"
	}
}

func isCodexTOML(client string) bool {
	return strings.ToLower(client) == "codex"
}

func runInstall(client string, force, dryRun bool) error {
	cfgPath, err := clientConfigPath(client)
	if err != nil {
		return err
	}
	bin, err := binaryPath()
	if err != nil {
		return err
	}

	if isCodexTOML(client) {
		return runInstallCodexTOML(cfgPath, bin, force, dryRun)
	}

	cfg, existingData, err := loadJSONConfig(cfgPath, force)
	if err != nil {
		return err
	}

	key := mcpConfigKey(client)
	servers, _ := cfg[key].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}

	if existing, ok := servers["trading-cli"]; ok && !force {
		if dryRun {
			fmt.Printf("trading-cli is already installed in %s\n  existing: %v\n  (use --force to overwrite)\n", cfgPath, existing)
			return nil
		}
		fmt.Printf("trading-cli is already installed in %s\nUse --force to overwrite.\n", cfgPath)
		return nil
	}

	servers["trading-cli"] = map[string]any{
		"command": bin,
		"args":    []string{"mcp"},
	}
	cfg[key] = servers

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if dryRun {
		fmt.Printf("Would write to %s:\n\n%s\n", cfgPath, out)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if len(existingData) > 0 {
		_ = os.WriteFile(cfgPath+".trading-cli.bak", existingData, 0o644)
	}

	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", cfgPath, err)
	}

	fmt.Printf("Installed trading-cli as MCP server for %s.\n", client)
	fmt.Printf("  config: %s\n", cfgPath)
	fmt.Printf("  binary: %s\n", bin)
	fmt.Println()
	fmt.Println("Restart your AI client to pick up the change.")
	return nil
}

func loadJSONConfig(cfgPath string, force bool) (map[string]any, []byte, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil, nil
		}
		return nil, nil, fmt.Errorf("read config %s: %w", cfgPath, err)
	}
	if len(data) == 0 {
		return map[string]any{}, data, nil
	}

	cfg := map[string]any{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		if force {
			return map[string]any{}, data, nil
		}
		return nil, data, fmt.Errorf("parse existing config %s: %w (fix the file or use --force to overwrite)", cfgPath, err)
	}
	return cfg, data, nil
}

func runInstallCodexTOML(cfgPath, bin string, force, dryRun bool) error {
	entry := fmt.Sprintf("\n[mcp_servers.trading-cli]\ncommand = %q\nargs = [\"mcp\"]\n", bin)

	existing, _ := os.ReadFile(cfgPath)
	content := string(existing)

	if strings.Contains(content, "[mcp_servers.trading-cli]") && !force {
		if dryRun {
			fmt.Printf("trading-cli is already in %s (use --force to overwrite)\n", cfgPath)
			return nil
		}
		fmt.Printf("trading-cli is already installed in %s\nUse --force to overwrite.\n", cfgPath)
		return nil
	}

	if dryRun {
		fmt.Printf("Would append to %s:\n%s\n", cfgPath, entry)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if len(existing) > 0 {
		_ = os.WriteFile(cfgPath+".trading-cli.bak", existing, 0o644)
	}

	f, err := os.OpenFile(cfgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open config %s: %w", cfgPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write config %s: %w", cfgPath, err)
	}

	fmt.Printf("Installed trading-cli as MCP server for Codex.\n")
	fmt.Printf("  config: %s\n", cfgPath)
	fmt.Printf("  binary: %s\n", bin)
	return nil
}
