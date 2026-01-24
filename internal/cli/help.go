package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command group annotations
const (
	GroupCore    = "core"
	GroupAuth    = "auth"
	GroupUtility = "utility"
)

// Resource categories for view/do subcommands
const (
	CatOrganizations   = "organizations"
	CatContent         = "content"
	CatProjects        = "projects"
	CatFleet           = "fleet"
	CatMaterials       = "materials"
	CatCertifications  = "certifications"
	CatClassifications = "classifications"
	CatReference       = "reference"
)

// resourceCategories maps resource names to their categories
var resourceCategories = map[string][]string{
	// Organizations
	"bidders":        {CatOrganizations},
	"brokers":        {CatOrganizations},
	"business-units": {CatOrganizations},
	"customers":      {CatOrganizations},
	"developers":     {CatOrganizations},
	"memberships":    {CatOrganizations},
	"truckers":       {CatOrganizations, CatFleet}, // appears in both
	"users":          {CatOrganizations},

	// Content & Publishing
	"features":          {CatContent},
	"glossary-terms":    {CatContent},
	"newsletters":       {CatContent},
	"platform-statuses": {CatContent},
	"posts":             {CatContent},
	"press-releases":    {CatContent},
	"release-notes":     {CatContent},

	// Projects & Jobs
	"action-items":                                             {CatProjects},
	"crew-assignment-confirmations":                            {CatProjects},
	"job-production-plan-cancellation-reason-types":            {CatProjects},
	"job-production-plan-duplication-works":                    {CatProjects},
	"job-production-plan-material-types":                       {CatProjects},
	"job-production-plan-service-type-unit-of-measure-cohorts": {CatProjects},
	"job-production-plan-time-card-approvers":                  {CatProjects},
	"job-production-plans":                                     {CatProjects},
	"job-schedule-shift-start-site-changes":                    {CatProjects},
	"invoice-revisionizing-works":                              {CatProjects},
	"job-sites":                                                {CatProjects},
	"lineup-dispatch-shifts":                                   {CatProjects},
	"time-sheet-line-item-equipment-requirements":              {CatProjects},
	"time-card-pre-approvals":                                  {CatProjects},
	"time-card-unscrappages":                                   {CatProjects},
	"time-sheet-rejections":                                    {CatProjects},
	"service-sites":                                            {CatProjects},
	"project-categories":                                       {CatProjects},
	"project-divisions":                                        {CatProjects},
	"project-estimate-file-imports":                            {CatProjects},
	"project-offices":                                          {CatProjects},
	"projects":                                                 {CatProjects},

	// Fleet & Transport
	"driver-day-adjustment-plans":              {CatFleet},
	"driver-day-shortfall-calculations":        {CatFleet},
	"shift-counters":                           {CatFleet},
	"equipment-utilization-readings":           {CatFleet},
	"equipment-movement-requirement-locations": {CatFleet},
	"maintenance-requirement-sets":             {CatFleet},
	"hos-events":                               {CatFleet},
	"tractors":                                 {CatFleet},
	"trailers":                                 {CatFleet},
	"transport-orders":                         {CatFleet},
	"transport-routes":                         {CatFleet},

	// Materials
	"inventory-estimates":              {CatMaterials},
	"material-site-measures":           {CatMaterials},
	"material-sites":                   {CatMaterials},
	"material-supplier-memberships":    {CatMaterials},
	"material-suppliers":               {CatMaterials},
	"material-transaction-inspections": {CatMaterials},
	"material-transactions":            {CatMaterials},
	"material-type-unavailabilities":   {CatMaterials},
	"material-types":                   {CatMaterials},

	// Certifications & Credentials
	"certification-requirements": {CatCertifications},
	"certification-types":        {CatCertifications},
	"certifications":             {CatCertifications},
	"tractor-credentials":        {CatCertifications},
	"trailer-credentials":        {CatCertifications},
	"user-credentials":           {CatCertifications},

	// Classifications
	"cost-codes":                 {CatClassifications},
	"cost-index-entries":         {CatClassifications},
	"cost-indexes":               {CatClassifications},
	"craft-classes":              {CatClassifications},
	"crafts":                     {CatClassifications},
	"custom-work-order-statuses": {CatClassifications},
	"developer-reference-types":  {CatClassifications},
	"developer-trucker-certification-classifications": {CatClassifications},
	"equipment-classifications":                       {CatClassifications},
	"incident-tags":                                   {CatClassifications},
	"labor-classifications":                           {CatClassifications},
	"profit-improvement-categories":                   {CatClassifications},
	"project-cost-classifications":                    {CatClassifications},
	"project-labor-classifications":                   {CatClassifications},
	"project-resource-classifications":                {CatClassifications},
	"project-revenue-classifications":                 {CatClassifications},
	"project-transport-event-types":                   {CatClassifications},
	"project-transport-location-event-types":          {CatClassifications},
	"quality-control-classifications":                 {CatClassifications},
	"reaction-classifications":                        {CatClassifications},
	"shift-feedback-reasons":                          {CatClassifications},
	"stakeholder-classifications":                     {CatClassifications},
	"time-sheet-line-item-classifications":            {CatClassifications},
	"tractor-trailer-credential-classifications":      {CatClassifications},
	"trailer-classifications":                         {CatClassifications},
	"truck-scopes":                                    {CatClassifications},
	"user-credential-classifications":                 {CatClassifications},

	// Reference Data
	"application-settings":          {CatReference},
	"culture-values":                {CatReference},
	"external-identification-types": {CatReference},
	"languages":                     {CatReference},
	"service-types":                 {CatReference},
	"tag-categories":                {CatReference},
	"tags":                          {CatReference},
	"unit-of-measures":              {CatReference},
}

// categoryOrder defines the display order and titles for categories
var categoryOrder = []struct {
	key   string
	title string
}{
	{CatOrganizations, "Organizations"},
	{CatContent, "Content & Publishing"},
	{CatProjects, "Projects & Jobs"},
	{CatFleet, "Fleet & Transport"},
	{CatMaterials, "Materials"},
	{CatCertifications, "Certifications & Credentials"},
	{CatClassifications, "Classifications"},
	{CatReference, "Reference Data"},
}

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
	fmt.Fprintln(out, "  # View: browse and read data")
	fmt.Fprintln(out, "  xbe view brokers list                      # List all brokers")
	fmt.Fprintln(out, "  xbe view projects list --status active     # Filter by status")
	fmt.Fprintln(out, "  xbe view newsletters show 123              # Show specific record")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  # Do: create, update, delete data")
	fmt.Fprintln(out, "  xbe do customers create --name \"Acme Corp\"")
	fmt.Fprintln(out, "  xbe do projects update 456 --status complete")
	fmt.Fprintln(out, "  xbe do posts delete 789 --confirm")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  # Summarize: aggregate data for analysis")
	fmt.Fprintln(out, "  xbe summarize shift-summary create --project 123 --start-on 2025-01-01")
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

		// Check if we have both view and do - if so, consolidate their resources
		var viewCmd, doCmd *cobra.Command
		var otherCmds []*cobra.Command
		for _, cmd := range cmds {
			switch cmd.Name() {
			case "view":
				viewCmd = cmd
			case "do":
				doCmd = cmd
			default:
				otherCmds = append(otherCmds, cmd)
			}
		}

		// Print view and do with consolidated resources
		if viewCmd != nil && doCmd != nil {
			fmt.Fprintf(out, "    %-20s %s\n", "view", viewCmd.Short)
			fmt.Fprintf(out, "    %-20s %s\n", "do", doCmd.Short)
			fmt.Fprintln(out)
			printConsolidatedResources(out, viewCmd, doCmd)
		} else {
			if viewCmd != nil {
				printCommandCompact(out, viewCmd, "    ")
			}
			if doCmd != nil {
				printCommandCompact(out, doCmd, "    ")
			}
		}

		// Print other commands normally
		for _, cmd := range otherCmds {
			printCommandCompact(out, cmd, "    ")
		}
	}
}

// printConsolidatedResources prints a single resource list for both view and do commands
func printConsolidatedResources(out io.Writer, viewCmd, doCmd *cobra.Command) {
	// Collect resources from both commands
	viewResources := make(map[string]bool)
	doResources := make(map[string]bool)

	for _, sub := range viewCmd.Commands() {
		if !sub.Hidden && sub.IsAvailableCommand() {
			viewResources[sub.Name()] = true
		}
	}
	for _, sub := range doCmd.Commands() {
		if !sub.Hidden && sub.IsAvailableCommand() {
			doResources[sub.Name()] = true
		}
	}

	// Find resources in both, view-only, and do-only
	allResources := make(map[string]bool)
	for name := range viewResources {
		allResources[name] = true
	}
	for name := range doResources {
		allResources[name] = true
	}

	var viewOnly, doOnly []string
	for name := range allResources {
		inView := viewResources[name]
		inDo := doResources[name]
		if inView && !inDo {
			viewOnly = append(viewOnly, name)
		} else if inDo && !inView {
			doOnly = append(doOnly, name)
		}
	}
	sort.Strings(viewOnly)
	sort.Strings(doOnly)

	// Bucket by category
	buckets := make(map[string][]string)
	var uncategorized []string

	for name := range allResources {
		categories, ok := resourceCategories[name]
		if !ok {
			uncategorized = append(uncategorized, name)
			continue
		}
		for _, cat := range categories {
			buckets[cat] = append(buckets[cat], name)
		}
	}

	// Print header
	fmt.Fprintln(out, "    Resources (use with 'view' or 'do'):")

	// Print each category
	for _, catInfo := range categoryOrder {
		names := buckets[catInfo.key]
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)

		fmt.Fprintf(out, "      [%s]\n", catInfo.title)
		for _, name := range names {
			fmt.Fprintf(out, "        %s\n", name)
		}
	}

	// Print uncategorized
	if len(uncategorized) > 0 {
		sort.Strings(uncategorized)
		fmt.Fprintln(out, "      [Other]")
		for _, name := range uncategorized {
			fmt.Fprintf(out, "        %s\n", name)
		}
	}

	// Note view-only or do-only resources
	if len(viewOnly) > 0 {
		fmt.Fprintf(out, "      View-only: %s\n", strings.Join(viewOnly, ", "))
	}
	if len(doOnly) > 0 {
		fmt.Fprintf(out, "      Do-only: %s\n", strings.Join(doOnly, ", "))
	}

	fmt.Fprintln(out, "      Use 'xbe view --help' or 'xbe do --help' for descriptions")
}

func printCommandCompact(out io.Writer, cmd *cobra.Command, indent string) {
	// Print the command itself
	fmt.Fprintf(out, "%s%-20s %s\n", indent, cmd.Name(), cmd.Short)

	// For view and do commands, use consolidated resource list
	if cmd.Name() == "view" || cmd.Name() == "do" {
		printResourcesCompact(out, cmd, indent+"  ")
		return
	}

	// For summarize, show compact list
	if cmd.Name() == "summarize" {
		printSummarizeCompact(out, cmd, indent+"  ")
		return
	}

	// Print subcommands recursively for other commands
	subCmds := cmd.Commands()
	sort.Slice(subCmds, func(i, j int) bool {
		return subCmds[i].Name() < subCmds[j].Name()
	})

	for _, sub := range subCmds {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		printCommandCompact(out, sub, indent+"  ")
	}
}

// printResourcesCompact prints resources grouped by category, one per line, no descriptions
func printResourcesCompact(out io.Writer, cmd *cobra.Command, indent string) {
	// Collect available subcommands
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		subcommands[sub.Name()] = sub
	}

	// Bucket commands by category
	buckets := make(map[string][]*cobra.Command)
	uncategorized := []*cobra.Command{}

	for name, sub := range subcommands {
		categories, ok := resourceCategories[name]
		if !ok {
			uncategorized = append(uncategorized, sub)
			continue
		}
		for _, cat := range categories {
			buckets[cat] = append(buckets[cat], sub)
		}
	}

	// Print each category in order
	for _, catInfo := range categoryOrder {
		cmds := buckets[catInfo.key]
		if len(cmds) == 0 {
			continue
		}

		// Sort commands alphabetically within category
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		fmt.Fprintf(out, "%s[%s]\n", indent, catInfo.title)
		for _, sub := range cmds {
			fmt.Fprintf(out, "%s  %s\n", indent, sub.Name())
		}
	}

	// Print any uncategorized commands
	if len(uncategorized) > 0 {
		sort.Slice(uncategorized, func(i, j int) bool {
			return uncategorized[i].Name() < uncategorized[j].Name()
		})

		fmt.Fprintf(out, "%s[Other]\n", indent)
		for _, sub := range uncategorized {
			fmt.Fprintf(out, "%s  %s\n", indent, sub.Name())
		}
	}

	// Add affordance for full details
	fmt.Fprintf(out, "%sUse 'xbe %s --help' for descriptions\n", indent, cmd.Name())
}

// printSummarizeCompact prints summarize subcommands compactly
func printSummarizeCompact(out io.Writer, cmd *cobra.Command, indent string) {
	subCmds := cmd.Commands()
	sort.Slice(subCmds, func(i, j int) bool {
		return subCmds[i].Name() < subCmds[j].Name()
	})

	for _, sub := range subCmds {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		fmt.Fprintf(out, "%s%s\n", indent, sub.Name())
	}
	fmt.Fprintf(out, "%sUse 'xbe summarize --help' for descriptions\n", indent)
}

func printCommandWithSubcommands(out io.Writer, cmd *cobra.Command, indent string) {
	// Print the command itself
	fmt.Fprintf(out, "%s%-20s %s\n", indent, cmd.Name(), cmd.Short)

	// For view and do commands, use category grouping instead of full tree
	if cmd.Name() == "view" || cmd.Name() == "do" {
		printSubcommandsInTree(out, cmd, indent+"  ")
		return
	}

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

// printSubcommandsInTree prints subcommands grouped by category for use in root help tree
func printSubcommandsInTree(out io.Writer, cmd *cobra.Command, indent string) {
	// Collect available subcommands
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		subcommands[sub.Name()] = sub
	}

	// Bucket commands by category (only first category to avoid duplicates in tree)
	buckets := make(map[string][]*cobra.Command)
	uncategorized := []*cobra.Command{}

	for name, sub := range subcommands {
		categories, ok := resourceCategories[name]
		if !ok {
			uncategorized = append(uncategorized, sub)
			continue
		}
		for _, cat := range categories {
			buckets[cat] = append(buckets[cat], sub)
		}
	}

	// Print each category in order
	for _, catInfo := range categoryOrder {
		cmds := buckets[catInfo.key]
		if len(cmds) == 0 {
			continue
		}

		// Sort commands alphabetically within category
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		fmt.Fprintf(out, "%s[%s]\n", indent, catInfo.title)
		for _, sub := range cmds {
			fmt.Fprintf(out, "%s  %-40s %s\n", indent, sub.Name(), sub.Short)
		}
	}

	// Print any uncategorized commands
	if len(uncategorized) > 0 {
		sort.Slice(uncategorized, func(i, j int) bool {
			return uncategorized[i].Name() < uncategorized[j].Name()
		})

		fmt.Fprintf(out, "%s[Other]\n", indent)
		for _, sub := range uncategorized {
			fmt.Fprintf(out, "%s  %-40s %s\n", indent, sub.Name(), sub.Short)
		}
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
	// Use grouped output for view and do commands
	if cmd.Name() == "view" || cmd.Name() == "do" {
		printSubcommandsGrouped(out, cmd)
		return
	}

	// Flat list for other commands (auth, etc.)
	fmt.Fprintln(out, "AVAILABLE COMMANDS:")
	for _, sub := range cmd.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		fmt.Fprintf(out, "  %-18s %s\n", sub.Name(), sub.Short)
	}
}

func printSubcommandsGrouped(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "RESOURCES:")

	// Collect available subcommands
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		subcommands[sub.Name()] = sub
	}

	// Bucket commands by category
	buckets := make(map[string][]*cobra.Command)
	uncategorized := []*cobra.Command{}

	for name, sub := range subcommands {
		categories, ok := resourceCategories[name]
		if !ok {
			uncategorized = append(uncategorized, sub)
			continue
		}
		for _, cat := range categories {
			buckets[cat] = append(buckets[cat], sub)
		}
	}

	// Print each category in order
	for _, catInfo := range categoryOrder {
		cmds := buckets[catInfo.key]
		if len(cmds) == 0 {
			continue
		}

		// Sort commands alphabetically within category
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		fmt.Fprintf(out, "  [%s]\n", catInfo.title)
		for _, sub := range cmds {
			fmt.Fprintf(out, "    %s\n", sub.Name())
		}
	}

	// Print any uncategorized commands
	if len(uncategorized) > 0 {
		sort.Slice(uncategorized, func(i, j int) bool {
			return uncategorized[i].Name() < uncategorized[j].Name()
		})

		fmt.Fprintln(out, "  [Other]")
		for _, sub := range uncategorized {
			fmt.Fprintf(out, "    %s\n", sub.Name())
		}
	}

	// Affordance for detailed descriptions
	fmt.Fprintf(out, "\nUse 'xbe %s <resource> --help' for resource details and available operations.\n", cmd.Name())
}

// Global flags that appear on most commands (documented in root help)
var (
	paginationFlags = map[string]bool{"limit": true, "offset": true, "sort": true}
	outputFlags     = map[string]bool{"json": true}
	connectionFlags = map[string]bool{"base-url": true, "token": true, "no-auth": true}
)

func printFlags(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "FLAGS:")

	// Collect command-specific flags (filters), skip global flags
	filters := []string{}
	hasGlobalFlags := false

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}

		// Check if this is a global flag
		if paginationFlags[f.Name] || outputFlags[f.Name] || connectionFlags[f.Name] {
			hasGlobalFlags = true
			return
		}

		filters = append(filters, formatFlag(f))
	})

	// Print command-specific flags
	for _, usage := range filters {
		fmt.Fprint(out, usage)
	}

	// Reference global flags if any were present
	if hasGlobalFlags {
		if len(filters) > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, "  Use 'xbe --help' for global flags (--json, --limit, --base-url, etc.)")
	}
}

// formatFlag formats a single flag for display, matching Cobra's format
func formatFlag(f *pflag.Flag) string {
	var buf strings.Builder

	if f.Shorthand != "" {
		buf.WriteString(fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name))
	} else {
		buf.WriteString(fmt.Sprintf("      --%s", f.Name))
	}

	varType, usage := pflag.UnquoteUsage(f)
	if varType != "" {
		buf.WriteString(" ")
		buf.WriteString(varType)
	}

	// Calculate padding for alignment (similar to Cobra's default)
	padding := 28 - buf.Len()
	if padding < 1 {
		padding = 1
	}
	buf.WriteString(strings.Repeat(" ", padding))

	buf.WriteString(usage)

	// Add default value if not zero value
	if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {
		buf.WriteString(fmt.Sprintf(" (default %q)", f.DefValue))
	}

	buf.WriteString("\n")
	return buf.String()
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
