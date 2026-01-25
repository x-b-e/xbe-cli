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

type timeCardCostCodeAllocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardCostCodeAllocationDetails struct {
	ID          string   `json:"id"`
	TimeCardID  string   `json:"time_card_id,omitempty"`
	CostCodeIDs []string `json:"cost_code_ids,omitempty"`
	Details     any      `json:"details,omitempty"`
}

func newTimeCardCostCodeAllocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card cost code allocation details",
		Long: `Show the full details of a time card cost code allocation.

Output Fields:
  ID
  Time Card ID
  Cost Code IDs
  Details

Arguments:
  <id>    The allocation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time card cost code allocation
  xbe view time-card-cost-code-allocations show 123

  # Output as JSON
  xbe view time-card-cost-code-allocations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardCostCodeAllocationsShow,
	}
	initTimeCardCostCodeAllocationsShowFlags(cmd)
	return cmd
}

func init() {
	timeCardCostCodeAllocationsCmd.AddCommand(newTimeCardCostCodeAllocationsShowCmd())
}

func initTimeCardCostCodeAllocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardCostCodeAllocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardCostCodeAllocationsShowOptions(cmd)
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
		return fmt.Errorf("time card cost code allocation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-cost-code-allocations]", "details,time-card,cost-codes")

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-cost-code-allocations/"+id, query)
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

	details := buildTimeCardCostCodeAllocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardCostCodeAllocationDetails(cmd, details)
}

func parseTimeCardCostCodeAllocationsShowOptions(cmd *cobra.Command) (timeCardCostCodeAllocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardCostCodeAllocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardCostCodeAllocationDetails(resp jsonAPISingleResponse) timeCardCostCodeAllocationDetails {
	resource := resp.Data
	details := allocationDetailsValue(resource.Attributes)
	costCodeIDs := costCodeIDsFromResource(resource, details)

	out := timeCardCostCodeAllocationDetails{
		ID:          resource.ID,
		CostCodeIDs: costCodeIDs,
		Details:     details,
	}
	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		out.TimeCardID = rel.Data.ID
	}
	return out
}

func renderTimeCardCostCodeAllocationDetails(cmd *cobra.Command, details timeCardCostCodeAllocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card ID: %s\n", details.TimeCardID)
	}
	if len(details.CostCodeIDs) > 0 {
		fmt.Fprintf(out, "Cost Code IDs: %s\n", strings.Join(details.CostCodeIDs, ", "))
	}
	if details.Details != nil {
		prettyDetails := formatJSONValue(details.Details)
		if prettyDetails != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, prettyDetails)
		}
	}

	return nil
}
