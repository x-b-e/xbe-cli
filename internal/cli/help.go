package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Command group annotations
const (
	GroupCore    = "core"
	GroupAuth    = "auth"
	GroupUtility = "utility"
)

// initHelp configures the custom help system for the root command.
func initHelp(cmd *cobra.Command) {
	cmd.SetHelpFunc(customHelpFunc)
	cmd.SetUsageFunc(customUsageFunc)
}

func customHelpFunc(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStdout()

	// Print Long description if available, otherwise Short
	if cmd.Long != "" {
		fmt.Fprintln(out, cmd.Long)
	} else if cmd.Short != "" {
		fmt.Fprintln(out, cmd.Short)
	}

	// If this is the root command, print the full command reference
	if cmd.Parent() == nil {
		fmt.Fprintln(out)
		printQuickStart(out)
		fmt.Fprintln(out)
		printDataAnalysis(out)
		fmt.Fprintln(out)
		printCommandTree(out, cmd)
		fmt.Fprintln(out)
		printGlobalFlags(out)
		fmt.Fprintln(out)
		printConfiguration(out)
		fmt.Fprintln(out)
		printLearnMore(out, cmd)
	} else {
		// For subcommands, print standard help
		fmt.Fprintln(out)
		printUsage(out, cmd)

		if cmd.HasAvailableSubCommands() {
			fmt.Fprintln(out)
			printSubcommands(out, cmd)
		}

		if cmd.HasAvailableLocalFlags() {
			fmt.Fprintln(out)
			printFlags(out, cmd)
		}

		if cmd.Example != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "EXAMPLES:")
			fmt.Fprintln(out, cmd.Example)
		}
	}
}

func customUsageFunc(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	printUsage(out, cmd)

	if cmd.HasAvailableSubCommands() {
		fmt.Fprintln(out)
		printSubcommands(out, cmd)
	}

	if cmd.HasAvailableLocalFlags() {
		fmt.Fprintln(out)
		printFlags(out, cmd)
	}

	return nil
}

func printQuickStart(out io.Writer) {
	fmt.Fprintln(out, "QUICK START:")
	fmt.Fprintln(out, "  # Authenticate (token stored securely in system keychain)")
	fmt.Fprintln(out, "  xbe auth login")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  # List recent newsletters")
	fmt.Fprintln(out, "  xbe view newsletters list")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  # View a specific newsletter")
	fmt.Fprintln(out, "  xbe view newsletters show 123")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  # List brokers")
	fmt.Fprintln(out, "  xbe view brokers list")
}

func printDataAnalysis(out io.Writer) {
	fmt.Fprintln(out, "DATA ANALYSIS:")
	fmt.Fprintln(out, "  Summary commands aggregate large datasets (like pivot tables) for analysis.")
	fmt.Fprintln(out, "  Use these when you need totals, averages, or grouped statistics:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  xbe do lane-summary create                   Aggregate hauling/cycle data")
	fmt.Fprintln(out, "  xbe do material-transaction-summary create   Aggregate material transactions")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Example: Summarize tons by material site for a date range")
	fmt.Fprintln(out, "    xbe do material-transaction-summary create \\")
	fmt.Fprintln(out, "      --group-by material_site \\")
	fmt.Fprintln(out, "      --filter broker=123 --filter date_min=2025-01-01")
}

func printCommandTree(out io.Writer, root *cobra.Command) {
	fmt.Fprintln(out, "COMMANDS:")

	// Group commands by annotation
	groups := map[string][]*cobra.Command{
		GroupCore:    {},
		GroupAuth:    {},
		GroupUtility: {},
	}

	for _, cmd := range root.Commands() {
		if cmd.Hidden || !cmd.IsAvailableCommand() {
			continue
		}
		group := cmd.Annotations["group"]
		if group == "" {
			group = GroupCore
		}
		groups[group] = append(groups[group], cmd)
	}

	// Print each group
	groupOrder := []struct {
		key   string
		title string
	}{
		{GroupCore, "Core Commands"},
		{GroupAuth, "Authentication"},
		{GroupUtility, "Utility"},
	}

	for _, g := range groupOrder {
		cmds := groups[g.key]
		if len(cmds) == 0 {
			continue
		}

		fmt.Fprintf(out, "\n  %s:\n", g.title)

		// Sort commands alphabetically within group
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		for _, cmd := range cmds {
			printCommandWithSubcommands(out, cmd, "    ")
		}
	}
}

func printCommandWithSubcommands(out io.Writer, cmd *cobra.Command, indent string) {
	// Print the command itself
	fmt.Fprintf(out, "%s%-20s %s\n", indent, cmd.Name(), cmd.Short)

	// Print subcommands recursively
	subCmds := cmd.Commands()
	sort.Slice(subCmds, func(i, j int) bool {
		return subCmds[i].Name() < subCmds[j].Name()
	})

	for _, sub := range subCmds {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		printCommandWithSubcommands(out, sub, indent+"  ")
	}
}

func printGlobalFlags(out io.Writer) {
	fmt.Fprintln(out, "GLOBAL FLAGS:")
	fmt.Fprintln(out, "  These flags are available on most commands:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  --base-url string    API base URL (default: https://app.x-b-e.com)")
	fmt.Fprintln(out, "  --token string       API token (overrides stored token)")
	fmt.Fprintln(out, "  --no-auth            Disable automatic token lookup")
	fmt.Fprintln(out, "  --json               Output in JSON format (machine-readable)")
	fmt.Fprintln(out, "  -h, --help           Show help for any command")
}

func printConfiguration(out io.Writer) {
	fmt.Fprintln(out, "CONFIGURATION:")
	fmt.Fprintln(out, "  Token Resolution (in order of precedence):")
	fmt.Fprintln(out, "    1. --token flag")
	fmt.Fprintln(out, "    2. XBE_TOKEN or XBE_API_TOKEN environment variable")
	fmt.Fprintln(out, "    3. System keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)")
	fmt.Fprintln(out, "    4. Config file at ~/.config/xbe/config.json")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Environment Variables:")
	fmt.Fprintln(out, "    XBE_TOKEN          API access token")
	fmt.Fprintln(out, "    XBE_API_TOKEN      API access token (alternative)")
	fmt.Fprintln(out, "    XBE_BASE_URL       API base URL")
	fmt.Fprintln(out, "    XDG_CONFIG_HOME    Config directory (default: ~/.config)")
}

func printLearnMore(out io.Writer, root *cobra.Command) {
	fmt.Fprintln(out, "LEARN MORE:")
	fmt.Fprintln(out, "  Use 'xbe <command> --help' for detailed information about a command.")
	fmt.Fprintln(out, "  Use 'xbe <command> <subcommand> --help' for subcommand details.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Examples:")
	fmt.Fprintln(out, "    xbe auth --help              Learn about authentication")
	fmt.Fprintln(out, "    xbe view newsletters --help  Learn about newsletter commands")
	fmt.Fprintln(out, "    xbe view brokers list --help See all filtering options for brokers")
}

func printUsage(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "USAGE:")
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(out, "  %s [command]\n", cmd.CommandPath())
	} else {
		fmt.Fprintf(out, "  %s\n", cmd.UseLine())
	}
}

func printSubcommands(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "AVAILABLE COMMANDS:")
	for _, sub := range cmd.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		fmt.Fprintf(out, "  %-18s %s\n", sub.Name(), sub.Short)
	}
}

func printFlags(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "FLAGS:")
	fmt.Fprint(out, cmd.LocalFlags().FlagUsages())
}

// wrapText wraps text at the specified width, preserving existing newlines.
func wrapText(text string, width int) string {
	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(wrapLine(line, width))
	}

	return result.String()
}

func wrapLine(line string, width int) string {
	if len(line) <= width {
		return line
	}

	var result strings.Builder
	words := strings.Fields(line)
	currentLen := 0

	for i, word := range words {
		if currentLen+len(word)+1 > width && currentLen > 0 {
			result.WriteString("\n")
			currentLen = 0
		} else if i > 0 {
			result.WriteString(" ")
			currentLen++
		}
		result.WriteString(word)
		currentLen += len(word)
	}

	return result.String()
}
