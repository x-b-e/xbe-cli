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

type costCodeTruckingCostSummariesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type costCodeTruckingCostSummaryDetails struct {
	ID            string `json:"id"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
	CreatedByID   string `json:"created_by_id,omitempty"`
	CreatedByName string `json:"created_by_name,omitempty"`
	StartOn       string `json:"start_on,omitempty"`
	EndOn         string `json:"end_on,omitempty"`
	Results       any    `json:"results,omitempty"`
}

func newCostCodeTruckingCostSummariesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show cost code trucking cost summary details",
		Long: `Show the full details of a cost code trucking cost summary.

Includes the summary date window, relationships, and computed results.

Arguments:
  <id>  Summary ID (required). Find IDs using the list command.`,
		Example: `  # Show a summary
  xbe view cost-code-trucking-cost-summaries show 123

  # Output JSON
  xbe view cost-code-trucking-cost-summaries show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCostCodeTruckingCostSummariesShow,
	}
	initCostCodeTruckingCostSummariesShowFlags(cmd)
	return cmd
}

func init() {
	costCodeTruckingCostSummariesCmd.AddCommand(newCostCodeTruckingCostSummariesShowCmd())
}

func initCostCodeTruckingCostSummariesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCostCodeTruckingCostSummariesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCostCodeTruckingCostSummariesShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("cost code trucking cost summary id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[cost-code-trucking-cost-summaries]", "start-on,end-on,results,broker,created-by")
	query.Set("include", "broker,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/cost-code-trucking-cost-summaries/"+id, query)
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

	details := buildCostCodeTruckingCostSummaryDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCostCodeTruckingCostSummaryDetails(cmd, details)
}

func parseCostCodeTruckingCostSummariesShowOptions(cmd *cobra.Command) (costCodeTruckingCostSummariesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return costCodeTruckingCostSummariesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCostCodeTruckingCostSummaryDetails(resp jsonAPISingleResponse) costCodeTruckingCostSummaryDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := costCodeTruckingCostSummaryDetails{
		ID:      resource.ID,
		StartOn: stringAttr(attrs, "start-on"),
		EndOn:   stringAttr(attrs, "end-on"),
		Results: anyAttr(attrs, "results"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(inc.Attributes, "company-name")
		}
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(inc.Attributes, "name")
		}
	}

	return details
}

func renderCostCodeTruckingCostSummaryDetails(cmd *cobra.Command, details costCodeTruckingCostSummaryDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}
	fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	fmt.Fprintf(out, "End On: %s\n", details.EndOn)

	if details.Results != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Results:")
		if formatted := formatCostCodeTruckingCostSummaryResults(details.Results); formatted != "" {
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}

func formatCostCodeTruckingCostSummaryResults(results any) string {
	if results == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", results)
	}
	return string(pretty)
}
