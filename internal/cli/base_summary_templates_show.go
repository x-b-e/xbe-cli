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

type baseSummaryTemplatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type baseSummaryTemplateDetails struct {
	ID              string   `json:"id"`
	Label           string   `json:"label"`
	GroupBys        []string `json:"group_bys,omitempty"`
	Filters         any      `json:"filters,omitempty"`
	ExplicitMetrics []string `json:"explicit_metrics,omitempty"`
	StartDate       string   `json:"start_date,omitempty"`
	EndDate         string   `json:"end_date,omitempty"`
	BrokerID        string   `json:"broker_id,omitempty"`
	CreatedByID     string   `json:"created_by_id,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}

func newBaseSummaryTemplatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show base summary template details",
		Long: `Show the full details of a base summary template.

Output Fields:
  ID               Template identifier
  Label            Template label
  Group Bys        Group-by fields
  Explicit Metrics Explicit metrics list
  Filters          Filter JSON
  Start Date       Optional start date
  End Date         Optional end date
  Broker ID        Broker ID (if scoped)
  Created By       Creator user ID
  Created At       Created timestamp
  Updated At       Updated timestamp

Arguments:
  <id>  The base summary template ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a base summary template
  xbe view base-summary-templates show 123

  # JSON output
  xbe view base-summary-templates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBaseSummaryTemplatesShow,
	}
	initBaseSummaryTemplatesShowFlags(cmd)
	return cmd
}

func init() {
	baseSummaryTemplatesCmd.AddCommand(newBaseSummaryTemplatesShowCmd())
}

func initBaseSummaryTemplatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBaseSummaryTemplatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBaseSummaryTemplatesShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("base summary template id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[base-summary-templates]", "label,group-bys,filters,explicit-metrics,start-date,end-date,broker,created-by,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/base-summary-templates/"+id, query)
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

	details := buildBaseSummaryTemplateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBaseSummaryTemplateDetails(cmd, details)
}

func parseBaseSummaryTemplatesShowOptions(cmd *cobra.Command) (baseSummaryTemplatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return baseSummaryTemplatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBaseSummaryTemplateDetails(resp jsonAPISingleResponse) baseSummaryTemplateDetails {
	attrs := resp.Data.Attributes
	details := baseSummaryTemplateDetails{
		ID:              resp.Data.ID,
		Label:           strings.TrimSpace(stringAttr(attrs, "label")),
		GroupBys:        stringSliceAttr(attrs, "group-bys"),
		ExplicitMetrics: stringSliceAttr(attrs, "explicit-metrics"),
		Filters:         attrs["filters"],
		StartDate:       formatDateTime(stringAttr(attrs, "start-date")),
		EndDate:         formatDateTime(stringAttr(attrs, "end-date")),
		CreatedAt:       formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:       formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderBaseSummaryTemplateDetails(cmd *cobra.Command, details baseSummaryTemplateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Label != "" {
		fmt.Fprintf(out, "Label: %s\n", details.Label)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.StartDate != "" {
		fmt.Fprintf(out, "Start Date: %s\n", details.StartDate)
	}
	if details.EndDate != "" {
		fmt.Fprintf(out, "End Date: %s\n", details.EndDate)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	groupBys := strings.Join(details.GroupBys, ", ")
	if groupBys == "" {
		groupBys = "(none)"
	}
	explicitMetrics := strings.Join(details.ExplicitMetrics, ", ")
	if explicitMetrics == "" {
		explicitMetrics = "(none)"
	}

	fmt.Fprintf(out, "Group Bys: %s\n", groupBys)
	fmt.Fprintf(out, "Explicit Metrics: %s\n", explicitMetrics)

	filters := formatBaseSummaryTemplateFilters(details.Filters)
	fmt.Fprintln(out, "Filters:")
	if filters == "" {
		fmt.Fprintln(out, "  (none)")
	} else {
		fmt.Fprintln(out, indentBaseSummaryTemplateFilters(filters, "  "))
	}

	return nil
}

func formatBaseSummaryTemplateFilters(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}

func indentBaseSummaryTemplateFilters(value, prefix string) string {
	if value == "" {
		return ""
	}
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}
