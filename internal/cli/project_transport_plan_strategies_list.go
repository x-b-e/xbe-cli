package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlanStrategiesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	Name        string
	StepPattern string
}

type projectTransportPlanStrategyRow struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
	StepPattern string `json:"step_pattern,omitempty"`
	StepCount   int    `json:"step_count,omitempty"`
}

func newProjectTransportPlanStrategiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan strategies",
		Long: `List project transport plan strategies.

Output Columns:
  ID            Strategy identifier
  NAME          Strategy name
  ACTIVE        Whether the strategy is active
  STEP_PATTERN  Step pattern identifier
  STEPS         Step count

Filters:
  --name          Filter by strategy name
  --step-pattern  Filter by step pattern

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List strategies
  xbe view project-transport-plan-strategies list

  # Filter by name
  xbe view project-transport-plan-strategies list --name "Default"

  # Filter by step pattern
  xbe view project-transport-plan-strategies list --step-pattern "pickup-dropoff"

  # JSON output
  xbe view project-transport-plan-strategies list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStrategiesList,
	}
	initProjectTransportPlanStrategiesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategiesCmd.AddCommand(newProjectTransportPlanStrategiesListCmd())
}

func initProjectTransportPlanStrategiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("name", "", "Filter by strategy name")
	cmd.Flags().String("step-pattern", "", "Filter by step pattern")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStrategiesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-strategies]", strings.Join([]string{
		"name",
		"is-active",
		"step-pattern",
		"steps",
	}, ","))

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[step-pattern]", opts.StepPattern)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategies", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildProjectTransportPlanStrategyRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStrategiesTable(cmd, rows)
}

func parseProjectTransportPlanStrategiesListOptions(cmd *cobra.Command) (projectTransportPlanStrategiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	name, _ := cmd.Flags().GetString("name")
	stepPattern, _ := cmd.Flags().GetString("step-pattern")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategiesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		Name:        name,
		StepPattern: stepPattern,
	}, nil
}

func buildProjectTransportPlanStrategyRows(resp jsonAPIResponse) []projectTransportPlanStrategyRow {
	rows := make([]projectTransportPlanStrategyRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanStrategyRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			IsActive:    boolAttr(resource.Attributes, "is-active"),
			StepPattern: stringAttr(resource.Attributes, "step-pattern"),
			StepCount:   countStrategySteps(resource.Relationships),
		}

		rows = append(rows, row)
	}
	return rows
}

func countStrategySteps(relationships map[string]jsonAPIRelationship) int {
	if relationships == nil {
		return 0
	}
	rel, ok := relationships["steps"]
	if !ok {
		return 0
	}
	return len(relationshipIDs(rel))
}

func renderProjectTransportPlanStrategiesTable(cmd *cobra.Command, rows []projectTransportPlanStrategyRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan strategies found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tACTIVE\tSTEP_PATTERN\tSTEPS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%t\t%s\t%d\n",
			row.ID,
			row.Name,
			row.IsActive,
			row.StepPattern,
			row.StepCount,
		)
	}
	return writer.Flush()
}
