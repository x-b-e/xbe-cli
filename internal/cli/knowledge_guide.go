package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type knowledgeGuideSection struct {
	Title    string   `json:"title"`
	Notes    []string `json:"notes,omitempty"`
	Commands []string `json:"commands,omitempty"`
}

type knowledgeGuidePayload struct {
	Sections []knowledgeGuideSection `json:"sections"`
}

func newKnowledgeGuideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "First-run playbook for AI agents",
		Long: `Show a practical first-run playbook for exploring the CLI.

This command captures the non-obvious rules that typically slow down new
agents: naming conventions, summary/resource mapping, filter discovery, and
how to move from discovery to execution safely.`,
		Example: `  # Show first-run playbook
  xbe knowledge guide

  # Machine-readable playbook
  xbe knowledge guide --json`,
		RunE: runKnowledgeGuide,
	}
}

func runKnowledgeGuide(cmd *cobra.Command, _ []string) error {
	payload := knowledgeGuidePayload{
		Sections: []knowledgeGuideSection{
			{
				Title: "Unknown Task Loop",
				Notes: []string{
					"Start with search, then inspect a resource, then inspect commands tied to that resource.",
					"Use neighbors/relations/filters when you need adjacent resources or indirect filter paths.",
				},
				Commands: []string{
					"xbe knowledge search <term>",
					"xbe knowledge resource <resource>",
					"xbe knowledge commands --resource <resource>",
					"xbe knowledge neighbors <resource>",
					"xbe knowledge filters --resource <resource>",
				},
			},
			{
				Title: "Naming Rules",
				Notes: []string{
					"Resource names are usually plural (jobs, customers, projects).",
					"The knowledge commands resolve common singular/plural mismatches automatically when possible.",
					"Use 'xbe knowledge resources --query <term>' when unsure.",
				},
			},
			{
				Title: "Summary Name Mapping",
				Notes: []string{
					"Summarize command names often differ from summary resource names.",
					"Example: 'transport-summary' command maps to 'transport-summaries' resource.",
					"'xbe knowledge summaries --summary' accepts command names or summary resource names.",
				},
				Commands: []string{
					"xbe summarize transport-summary create --help",
					"xbe knowledge summaries --summary transport-summary --details",
				},
			},
			{
				Title: "Filters And Flags",
				Notes: []string{
					"Use knowledge flags to see which CLI flags map to model fields.",
					"Use knowledge filters to see multi-hop filter paths (for example, when a flag filters through relationships).",
					"When using --command with filters, provide a specific substring to avoid ambiguous matches.",
				},
				Commands: []string{
					"xbe knowledge flags --command \"view jobs list\"",
					"xbe knowledge filters --command \"view jobs list\"",
				},
			},
			{
				Title: "Client URL Discovery",
				Notes: []string{
					"'--client-url' works on view list/show and emits client routes only.",
					"Some resources are not directly route-addressable; use client-routes --query for broader discovery.",
					"The jump-to route is documented in knowledge output, including required query params.",
				},
				Commands: []string{
					"xbe view jobs list --client-url",
					"xbe knowledge client-routes --query job",
					"xbe knowledge client-routes --query jump-to",
				},
			},
			{
				Title: "Safe Execution",
				Notes: []string{
					"Before write commands, inspect required flags and metadata.",
					"Use command metadata to check permissions, side effects, and validation rules.",
					"Prefer view/list plus --limit first when exploring unknown resources.",
				},
				Commands: []string{
					"xbe do <resource> <action> --help",
					"xbe do <resource> <action> --metadata",
					"xbe view <resource> list --limit 5",
				},
			},
			{
				Title: "Output For Agents",
				Notes: []string{
					"Use --json for stable machine output.",
					"Use --output yaml and --jq for filtered structured output without extra tooling.",
				},
				Commands: []string{
					"xbe knowledge commands --resource jobs --json",
					"xbe knowledge commands --resource jobs --output yaml --jq \".[] | .path\"",
				},
			},
		},
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, payload)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "FIRST-RUN PLAYBOOK")
	for _, section := range payload.Sections {
		fmt.Fprintf(out, "\n%s:\n", section.Title)
		for _, note := range section.Notes {
			fmt.Fprintf(out, "  - %s\n", note)
		}
		for _, command := range section.Commands {
			fmt.Fprintf(out, "  $ %s\n", command)
		}
	}
	return nil
}
