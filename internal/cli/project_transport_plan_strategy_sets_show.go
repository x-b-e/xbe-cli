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

type projectTransportPlanStrategySetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStrategySetDetails struct {
	ID              string   `json:"id"`
	Name            string   `json:"name,omitempty"`
	StrategyPattern string   `json:"strategy_pattern,omitempty"`
	IsActive        bool     `json:"is_active,omitempty"`
	StrategyIDs     []string `json:"strategy_ids,omitempty"`
}

func newProjectTransportPlanStrategySetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan strategy set details",
		Long: `Show the full details of a project transport plan strategy set.

Output Fields:
  ID        Strategy set identifier
  NAME      Strategy set name
  PATTERN   Strategy pattern derived from strategies
  ACTIVE    Whether the set is active
  STRATEGIES  Strategy IDs associated with the set

Arguments:
  <id>  Project transport plan strategy set ID (required).`,
		Example: `  # Show a project transport plan strategy set
  xbe view project-transport-plan-strategy-sets show 123

  # Output as JSON
  xbe view project-transport-plan-strategy-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStrategySetsShow,
	}
	initProjectTransportPlanStrategySetsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategySetsCmd.AddCommand(newProjectTransportPlanStrategySetsShowCmd())
}

func initProjectTransportPlanStrategySetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategySetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanStrategySetsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan strategy set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-strategy-sets]", "name,strategy-pattern,is-active,strategies")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategy-sets/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectTransportPlanStrategySetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStrategySetDetails(cmd, details)
}

func parseProjectTransportPlanStrategySetsShowOptions(cmd *cobra.Command) (projectTransportPlanStrategySetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategySetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStrategySetDetails(resp jsonAPISingleResponse) projectTransportPlanStrategySetDetails {
	attrs := resp.Data.Attributes

	details := projectTransportPlanStrategySetDetails{
		ID:              resp.Data.ID,
		Name:            stringAttr(attrs, "name"),
		StrategyPattern: stringAttr(attrs, "strategy-pattern"),
		IsActive:        boolAttr(attrs, "is-active"),
	}

	if rel, ok := resp.Data.Relationships["strategies"]; ok {
		details.StrategyIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectTransportPlanStrategySetDetails(cmd *cobra.Command, details projectTransportPlanStrategySetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.StrategyPattern != "" {
		fmt.Fprintf(out, "Strategy Pattern: %s\n", details.StrategyPattern)
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)
	if len(details.StrategyIDs) > 0 {
		fmt.Fprintf(out, "Strategies: %s\n", strings.Join(details.StrategyIDs, ", "))
	}

	return nil
}
