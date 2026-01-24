package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlanStrategiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStrategyDetails struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	IsActive    bool     `json:"is_active,omitempty"`
	StepPattern string   `json:"step_pattern,omitempty"`
	StepIDs     []string `json:"step_ids,omitempty"`
}

func newProjectTransportPlanStrategiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan strategy details",
		Long: `Show the full details of a project transport plan strategy.

Output Fields:
  ID
  Name
  Active
  Step Pattern
  Step IDs

Arguments:
  <id>    Strategy ID (required).`,
		Example: `  # Show strategy details
  xbe view project-transport-plan-strategies show 123

  # JSON output
  xbe view project-transport-plan-strategies show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStrategiesShow,
	}
	initProjectTransportPlanStrategiesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategiesCmd.AddCommand(newProjectTransportPlanStrategiesShowCmd())
}

func initProjectTransportPlanStrategiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanStrategiesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan strategy id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-strategies]", strings.Join([]string{
		"name",
		"is-active",
		"step-pattern",
		"steps",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategies/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildProjectTransportPlanStrategyDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStrategyDetails(cmd, details)
}

func parseProjectTransportPlanStrategiesShowOptions(cmd *cobra.Command) (projectTransportPlanStrategiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStrategyDetails(resp jsonAPISingleResponse) projectTransportPlanStrategyDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanStrategyDetails{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		IsActive:    boolAttr(attrs, "is-active"),
		StepPattern: stringAttr(attrs, "step-pattern"),
	}

	if rel, ok := resp.Data.Relationships["steps"]; ok {
		details.StepIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectTransportPlanStrategyDetails(cmd *cobra.Command, details projectTransportPlanStrategyDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	} else {
		fmt.Fprintln(out, "Name: (none)")
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)
	if details.StepPattern != "" {
		fmt.Fprintf(out, "Step Pattern: %s\n", details.StepPattern)
	} else {
		fmt.Fprintln(out, "Step Pattern: (none)")
	}
	if len(details.StepIDs) > 0 {
		fmt.Fprintf(out, "Step IDs: %s\n", strings.Join(details.StepIDs, ", "))
	} else {
		fmt.Fprintln(out, "Step IDs: (none)")
	}

	return nil
}
