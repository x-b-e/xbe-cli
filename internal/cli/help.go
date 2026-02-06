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
	GroupKnowledge = "knowledge"
	GroupCore      = "core"
	GroupAuth      = "auth"
	GroupUtility   = "utility"
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
	"bidders":                 {CatOrganizations},
	"brokers":                 {CatOrganizations},
	"broker-vendors":          {CatOrganizations},
	"business-unit-customers": {CatOrganizations},
	"business-unit-laborers":  {CatOrganizations},
	"business-units":          {CatOrganizations},
	"contractors":             {CatOrganizations},
	"customer-vendors":        {CatOrganizations},
	"customers":               {CatOrganizations},
	"developers":              {CatOrganizations},
	"memberships":             {CatOrganizations},
	"trucker-memberships":     {CatOrganizations, CatFleet},
	"trucker-settings":        {CatOrganizations, CatFleet},
	"open-door-issues":        {CatOrganizations},
	"organization-invoices-batch-invoice-batchings": {CatOrganizations},
	"organization-invoices-batch-invoice-failures":  {CatOrganizations},
	"organization-invoices-batch-processes":         {CatOrganizations},
	"organization-invoices-batch-pdf-files":         {CatOrganizations},
	"organization-invoices-batch-status-changes":    {CatOrganizations},
	"trucker-invoice-payments":                      {CatOrganizations, CatFleet},
	"truckers":                                      {CatOrganizations, CatFleet}, // appears in both
	"users":                                         {CatOrganizations},
	"user-languages":                                {CatOrganizations},

	// Content & Publishing
	"features":                {CatContent},
	"glossary-terms":          {CatContent},
	"newsletters":             {CatContent},
	"platform-statuses":       {CatContent},
	"posts":                   {CatContent},
	"press-releases":          {CatContent},
	"release-notes":           {CatContent},
	"public-praise-reactions": {CatContent},

	// Projects & Jobs
	"action-items":                                             {CatProjects},
	"action-item-key-results":                                  {CatProjects},
	"key-result-scrappages":                                    {CatProjects},
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
	"rate-agreements":                                          {CatProjects},
	"customer-retainers":                                       {CatProjects},
	"retainer-payment-deductions":                              {CatProjects},
	"tender-acceptances":                                       {CatProjects},
	"tender-offers":                                            {CatProjects},
	"tender-re-rates":                                          {CatProjects},
	"time-sheet-line-item-equipment-requirements":              {CatProjects},
	"time-card-pre-approvals":                                  {CatProjects},
	"time-card-unscrappages":                                   {CatProjects},
	"time-sheet-rejections":                                    {CatProjects},
	"incident-request-approvals":                               {CatProjects},
	"incident-request-rejections":                              {CatProjects},
	"incident-unit-of-measure-quantities":                      {CatProjects},
	"liability-incidents":                                      {CatProjects},
	"production-incidents":                                     {CatProjects},
	"predictions":                                              {CatProjects},
	"prediction-subject-bids":                                  {CatProjects},
	"prediction-subject-gaps":                                  {CatProjects},
	"service-sites":                                            {CatProjects},
	"project-categories":                                       {CatProjects},
	"project-divisions":                                        {CatProjects},
	"project-estimate-file-imports":                            {CatProjects},
	"projects-file-imports":                                    {CatProjects},
	"project-offices":                                          {CatProjects},
	"projects":                                                 {CatProjects},

	// Fleet & Transport
	"device-location-events":                   {CatFleet},
	"user-location-events":                     {CatFleet},
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
	"raw-material-transactions":        {CatMaterials},
	"material-type-unavailabilities":   {CatMaterials},
	"material-types":                   {CatMaterials},
	"pave-frame-actual-hours":          {CatMaterials},

	// Certifications & Credentials
	"certification-requirements":                  {CatCertifications},
	"certification-types":                         {CatCertifications},
	"certifications":                              {CatCertifications},
	"developer-trucker-certification-multipliers": {CatCertifications},
	"tractor-credentials":                         {CatCertifications},
	"trailer-credentials":                         {CatCertifications},
	"user-credentials":                            {CatCertifications},

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
	"broker-project-transport-event-types":            {CatClassifications},
	"project-transport-event-types":                   {CatClassifications},
	"project-transport-location-event-types":          {CatClassifications},
	"quality-control-classifications":                 {CatClassifications},
	"reaction-classifications":                        {CatClassifications},
	"shift-feedback-reasons":                          {CatClassifications},
	"objective-stakeholder-classification-quotes":     {CatClassifications},
	"stakeholder-classifications":                     {CatClassifications},
	"time-sheet-line-item-classifications":            {CatClassifications},
	"tractor-trailer-credential-classifications":      {CatClassifications},
	"trailer-classifications":                         {CatClassifications},
	"truck-scopes":                                    {CatClassifications},
	"user-credential-classifications":                 {CatClassifications},
	"work-order-service-codes":                        {CatClassifications},

	// Reference Data
	"application-settings":          {CatReference},
	"base-summary-templates":        {CatReference},
	"culture-values":                {CatReference},
	"email-address-statuses":        {CatReference},
	"external-identification-types": {CatReference},
	"languages":                     {CatReference},
	"model-filter-infos":            {CatReference},
	"search-catalog-entries":        {CatReference},
	"service-types":                 {CatReference},
	"tag-categories":                {CatReference},
	"taggings":                      {CatReference},
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
		fmt.Fprintln(out, formatCommandHelpText(cmd, cmd.Long))
	} else if cmd.Short != "" {
		fmt.Fprintln(out, formatCommandHelpText(cmd, cmd.Short))
	}
	if fieldsHelp := fieldsHelpForCommand(cmd); fieldsHelp != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, fieldsHelp)
	}

	// If this is the root command, print the full command reference
	if cmd.Parent() == nil {
		fmt.Fprintln(out)
		printBootstrapLoop(out)
		fmt.Fprintln(out)
		printCommandGrammar(out)
		fmt.Fprintln(out)
		printQuickStart(out)
		fmt.Fprintln(out)
		printCommandTree(out, cmd)
		fmt.Fprintln(out)
		printResourceDiscoveryExample(out)
		fmt.Fprintln(out)
		printKnowledgeTools(out)
		fmt.Fprintln(out)
		printMetadataFlags(out)
		fmt.Fprintln(out)
		printGlobalFlags(out)
		fmt.Fprintln(out)
		printTimeNotes(out)
		fmt.Fprintln(out)
		printAuthOverview(out)
		fmt.Fprintln(out)
		printRunHelp(out)
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

		if tip := clientURLHelpLine(cmd); tip != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, tip)
		}

		if fieldsExample := fieldsExampleForCommand(cmd); fieldsExample != "" {
			if cmd.Example == "" {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "EXAMPLES:")
			}
			fmt.Fprintln(out, "  # Include additional fields in output")
			fmt.Fprintln(out, fieldsExample)
		}

		if resource, ok := relatedDiscoveryResource(cmd); ok {
			fmt.Fprintln(out)
			printRelatedDiscovery(out, resource)
		}

		if intel := commandIntelHint(cmd); intel != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "COMMAND INTEL:")
			fmt.Fprintln(out, intel)
		}
	}
}

func clientURLHelpLine(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	if cmd.Name() == "view" {
		return "Tip: Use --client-url with view list/show to emit client app URL(s) only (see xbe --help)."
	}
	if _, ok := resourceForSparseList(cmd); ok {
		return "Tip: Use --client-url to emit client app URL(s) only (see xbe --help)."
	}
	if _, ok := resourceForSparseShow(cmd); ok {
		return "Tip: Use --client-url to emit client app URL(s) only (see xbe --help)."
	}
	return ""
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
	fmt.Fprintln(out, "  xbe knowledge guide")
	fmt.Fprintln(out, "  xbe knowledge search job")
	fmt.Fprintln(out, "  xbe knowledge resource jobs")
	fmt.Fprintln(out, "  xbe knowledge commands --resource jobs")
	fmt.Fprintln(out, "  xbe view jobs list --help")
	fmt.Fprintln(out, "  xbe view jobs list --limit 5")
}

func printBootstrapLoop(out io.Writer) {
	fmt.Fprintln(out, "BOOTSTRAP LOOP (for unknown tasks):")
	fmt.Fprintln(out, "  0) Orient:   xbe knowledge guide")
	fmt.Fprintln(out, "  1) Find:     xbe knowledge search <term>")
	fmt.Fprintln(out, "  2) Inspect:  xbe knowledge resource <resource>")
	fmt.Fprintln(out, "  3) Choose:   xbe knowledge commands --resource <resource> [--kind view|do|summarize]")
	fmt.Fprintln(out, "  4) Verify:   xbe <view|do|summarize> <resource> <action> --help")
	fmt.Fprintln(out, "  5) Explore:  xbe knowledge relations|neighbors|filters --resource <resource>")
	fmt.Fprintln(out, "  Note: 'knowledge' also has alias 'kb' (example: xbe kb search job).")
	fmt.Fprintln(out, "  Note: xbe knowledge commands shows permissions, side effects, and validation notes.")
}

func printCommandGrammar(out io.Writer) {
	fmt.Fprintln(out, "COMMAND GRAMMAR:")
	fmt.Fprintln(out, "  knowledge  xbe knowledge|kb <guide|search|resources|resource|commands|fields|flags|relations|neighbors|metapath|filters|summaries|client-routes> [filters]")
	fmt.Fprintln(out, "  read       xbe view <resource> <list|show> [flags]")
	fmt.Fprintln(out, "  write      xbe do <resource> <create|update|delete|action> [flags]")
	fmt.Fprintln(out, "  analyze    xbe summarize <summary> create [flags]")
	fmt.Fprintln(out, "  resource = the noun in view/do commands (e.g., jobs, customers)")
}

func printKnowledgeTools(out io.Writer) {
	fmt.Fprintln(out, "KNOWLEDGE TOOLS (what they answer):")
	fmt.Fprintln(out, "  guide      first-run playbook + non-obvious naming rules")
	fmt.Fprintln(out, "  search     find resources/commands/fields/summaries by term")
	fmt.Fprintln(out, "  resource   see fields, relationships, summaries, commands for one resource")
	fmt.Fprintln(out, "  commands   list CLI commands + permissions/side effects/validation")
	fmt.Fprintln(out, "  flags      map flags to field semantics (filter vs setter)")
	fmt.Fprintln(out, "  relations  discover related resources")
	fmt.Fprintln(out, "  neighbors  rank next-best resources to explore")
	fmt.Fprintln(out, "  filters    infer multi-hop filter paths from commands")
	fmt.Fprintln(out, "  metapath   similarity via shared features")
	fmt.Fprintln(out, "  fields     list fields + owning resources")
	fmt.Fprintln(out, "  summaries  list summary resources + group-by/metrics")
	fmt.Fprintln(out, "  client-routes  list client app routes, params, and curated docs (e.g. jump-to)")
	fmt.Fprintln(out, "  Note: summarize command names (transport-summary) often map to summary resources (transport-summaries).")
}

func printCommandTree(out io.Writer, root *cobra.Command) {
	fmt.Fprintln(out, "COMMANDS:")

	// Group commands by annotation
	groups := map[string][]*cobra.Command{
		GroupKnowledge: {},
		GroupCore:      {},
		GroupAuth:      {},
		GroupUtility:   {},
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
		{GroupKnowledge, "Knowledge & Exploration"},
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

		// Print view and do with a knowledge discovery hint
		if viewCmd != nil && doCmd != nil {
			printCommandLine(out, viewCmd, "    ")
			printCommandLine(out, doCmd, "    ")
			fmt.Fprintln(out)
			printResourceDiscoveryHint(out, "    ")
		} else {
			if viewCmd != nil {
				printCommandLine(out, viewCmd, "    ")
				printResourceDiscoveryHint(out, "    ")
			}
			if doCmd != nil {
				printCommandLine(out, doCmd, "    ")
				printResourceDiscoveryHint(out, "    ")
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

	// Avoid printing the full resource list in root help.
	if cmd.Name() == "view" || cmd.Name() == "do" {
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

func printResourceDiscoveryExample(out io.Writer) {
	fmt.Fprintln(out, "RESOURCE DISCOVERY (EXAMPLE OUTPUT):")
	fmt.Fprintln(out, "  $ xbe knowledge resources --query project")
	fmt.Fprintln(out, "  RESOURCE           LABEL_FIELDS   SERVER_TYPES")
	fmt.Fprintln(out, "  projects           name           Project")
	fmt.Fprintln(out, "  project-phases     name           ProjectPhase")
	fmt.Fprintln(out, "  project-offices    name           ProjectOffice")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  $ xbe knowledge commands --resource projects")
	fmt.Fprintln(out, "  COMMAND              KIND   VERB   RESOURCE   DESCRIPTION")
	fmt.Fprintln(out, "  view projects list   view   list   projects   List projects")
	fmt.Fprintln(out, "  view projects show   view   show   projects   Show project details")
	fmt.Fprintln(out, "  do projects update   do     update projects   Update a project")
}

func printCommandLine(out io.Writer, cmd *cobra.Command, indent string) {
	fmt.Fprintf(out, "%s%-20s %s\n", indent, cmd.Name(), cmd.Short)
}

func printResourceDiscoveryHint(out io.Writer, indent string) {
	fmt.Fprintf(out, "%sResource discovery:\n", indent)
	fmt.Fprintf(out, "%s  xbe knowledge resources\n", indent)
	fmt.Fprintf(out, "%s  xbe knowledge resource <name>\n", indent)
	fmt.Fprintf(out, "%s  xbe knowledge commands --resource <name>\n", indent)
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
	fmt.Fprintln(out, "  --json               machine-readable output")
	fmt.Fprintln(out, "  --output             output format: table (default), json, yaml")
	fmt.Fprintln(out, "  --jq                 jq-style filter for JSON/YAML output (--jq implies JSON if --output is unset)")
	fmt.Fprintln(out, "  --client-url         output client app URL(s) for view list/show")
	fmt.Fprintln(out, "  --limit/--offset/--sort  pagination for list commands")
	fmt.Fprintln(out, "  --fields             sparse fieldsets for list/show")
	fmt.Fprintln(out, "  --base-url/--token/--no-auth  auth/targeting")
	fmt.Fprintln(out, "  -h, --help           show help for any command")
}

func printMetadataFlags(out io.Writer) {
	fmt.Fprintln(out, "COMMAND METADATA FLAGS (view/do/summarize):")
	fmt.Fprintln(out, "  --metadata           print permissions + side effects + validation notes")
	fmt.Fprintln(out, "  --permissions        print only permission requirements")
	fmt.Fprintln(out, "  --side-effects       print only side effects")
	fmt.Fprintln(out, "  --validation-notes   print only validation notes")
}

func printTimeNotes(out io.Writer) {
	fmt.Fprintln(out, "TIMEZONES:")
	fmt.Fprintln(out, "  Timestamps are UTC unless explicitly labeled as local.")
}

func printAuthOverview(out io.Writer) {
	fmt.Fprintln(out, "AUTH:")
	fmt.Fprintln(out, "  xbe auth status | login | logout | whoami")
	fmt.Fprintln(out, "  Token precedence: --token > XBE_TOKEN/XBE_API_TOKEN > keychain > config")
}

func printRunHelp(out io.Writer) {
	fmt.Fprintln(out, "RUN HELP:")
	fmt.Fprintln(out, "  xbe <command> --help")
	fmt.Fprintln(out, "  xbe <command> <subcommand> --help")
	fmt.Fprintln(out, "  Tip: use --output yaml or --jq '<filter>' for filtered machine output")
}

func printUsage(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "USAGE:")
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(out, "  %s [command]\n", cmd.CommandPath())
	} else {
		fmt.Fprintf(out, "  %s\n", cmd.UseLine())
	}
}

func formatCommandHelpText(cmd *cobra.Command, text string) string {
	if text == "" {
		return text
	}
	if _, ok := resourceForSparseFields(cmd); ok {
		if strings.Contains(text, "Output Columns:") {
			text = strings.Replace(text, "Output Columns:", "Output Columns (default - use --fields for more):", 1)
		}
	}
	return text
}

func commandIntelHint(cmd *cobra.Command) string {
	if cmd == nil || cmd.HasAvailableSubCommands() {
		return ""
	}
	path := cmd.CommandPath()
	if path == "" {
		return ""
	}

	if rootCmd != nil {
		prefix := rootCmd.Name() + " "
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
		}
	}

	root := commandRootName(cmd)
	if root != "view" && root != "do" && root != "summarize" {
		return ""
	}
	return fmt.Sprintf("  xbe knowledge commands --query %q\n  xbe knowledge commands --resource <resource>", path)
}

func relatedDiscoveryResource(cmd *cobra.Command) (string, bool) {
	root := commandRootName(cmd)
	if root != "view" && root != "do" && root != "summarize" {
		return "", false
	}
	path := cmd.CommandPath()
	if rootCmd != nil {
		prefix := rootCmd.Name() + " "
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
		}
	}
	parts := strings.Fields(path)
	if len(parts) >= 2 {
		return parts[1], true
	}
	return "<resource>", true
}

func printRelatedDiscovery(out io.Writer, resource string) {
	if strings.TrimSpace(resource) == "" {
		resource = "<resource>"
	}
	fmt.Fprintln(out, "RELATED DISCOVERY:")
	fmt.Fprintf(out, "  xbe knowledge neighbors %s    Nearby resources to explore next\n", resource)
	fmt.Fprintf(out, "  xbe knowledge relations %s    Direct relationships for this resource\n", resource)
	fmt.Fprintf(out, "  xbe knowledge commands --resource %s    Commands that operate on this resource\n", resource)
}

func commandRootName(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	current := cmd
	for current.Parent() != nil && current.Parent().Parent() != nil {
		current = current.Parent()
	}
	if current == nil {
		return ""
	}
	return current.Name()
}

func fieldsExampleForCommand(cmd *cobra.Command) string {
	resource, ok := resourceForSparseFields(cmd)
	if !ok {
		return ""
	}
	attributeExample, _ := fieldsExamples(resource)
	if attributeExample == "" {
		return ""
	}
	path := cmd.CommandPath()
	if rootCmd != nil {
		prefix := rootCmd.Name() + " "
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
		}
	}
	arg := ""
	if cmd.Name() == "show" {
		arg = " 123"
	}
	return fmt.Sprintf("  xbe %s%s --fields %s", path, arg, attributeExample)
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
	outputFlags     = map[string]bool{"json": true, "output": true, "jq": true, "client-url": true}
	connectionFlags = map[string]bool{"base-url": true, "token": true, "no-auth": true}
	sparseFlags     = map[string]bool{"fields": true}
)

func printFlags(out io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(out, "FLAGS:")

	// Collect command-specific flags (filters), skip global flags
	filters := []string{}
	hasGlobalFlags := false

	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}

		// Check if this is a global flag
		if paginationFlags[f.Name] || outputFlags[f.Name] || connectionFlags[f.Name] || sparseFlags[f.Name] {
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
	if !hasGlobalFlags {
		cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden || f.Name == "help" {
				return
			}
			if paginationFlags[f.Name] || outputFlags[f.Name] || connectionFlags[f.Name] || sparseFlags[f.Name] {
				hasGlobalFlags = true
			}
		})
	}
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
