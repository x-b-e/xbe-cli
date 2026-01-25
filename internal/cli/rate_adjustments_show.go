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

type rateAdjustmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rateAdjustmentDetails struct {
	ID                                 string `json:"id"`
	RateID                             string `json:"rate_id,omitempty"`
	CostIndexID                        string `json:"cost_index_id,omitempty"`
	ParentRateAdjustmentID             string `json:"parent_rate_adjustment_id,omitempty"`
	ZeroInterceptValue                 string `json:"zero_intercept_value,omitempty"`
	ZeroInterceptRatio                 string `json:"zero_intercept_ratio,omitempty"`
	AdjustmentMin                      string `json:"adjustment_min,omitempty"`
	AdjustmentMax                      string `json:"adjustment_max,omitempty"`
	PreventRatingWhenIndexValueMissing bool   `json:"prevent_rating_when_index_value_missing"`
}

func newRateAdjustmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show rate adjustment details",
		Long: `Show the full details of a rate adjustment.

Output Fields:
  ID
  Rate ID
  Cost Index ID
  Parent Rate Adjustment ID
  Zero Intercept Value
  Zero Intercept Ratio
  Adjustment Min
  Adjustment Max
  Prevent Rating When Index Value Missing

Arguments:
  <id>    The rate adjustment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a rate adjustment
  xbe view rate-adjustments show 123

  # Output as JSON
  xbe view rate-adjustments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRateAdjustmentsShow,
	}
	initRateAdjustmentsShowFlags(cmd)
	return cmd
}

func init() {
	rateAdjustmentsCmd.AddCommand(newRateAdjustmentsShowCmd())
}

func initRateAdjustmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAdjustmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRateAdjustmentsShowOptions(cmd)
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
		return fmt.Errorf("rate adjustment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[rate-adjustments]", "zero-intercept-value,zero-intercept-ratio,adjustment-min,adjustment-max,prevent-rating-when-index-value-missing,rate,cost-index,parent-rate-adjustment")

	body, _, err := client.Get(cmd.Context(), "/v1/rate-adjustments/"+id, query)
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

	details := buildRateAdjustmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRateAdjustmentDetails(cmd, details)
}

func parseRateAdjustmentsShowOptions(cmd *cobra.Command) (rateAdjustmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAdjustmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRateAdjustmentDetails(resp jsonAPISingleResponse) rateAdjustmentDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := rateAdjustmentDetails{
		ID:                                 resource.ID,
		ZeroInterceptValue:                 stringAttr(attrs, "zero-intercept-value"),
		ZeroInterceptRatio:                 stringAttr(attrs, "zero-intercept-ratio"),
		AdjustmentMin:                      stringAttr(attrs, "adjustment-min"),
		AdjustmentMax:                      stringAttr(attrs, "adjustment-max"),
		PreventRatingWhenIndexValueMissing: boolAttr(attrs, "prevent-rating-when-index-value-missing"),
	}

	if rel, ok := resource.Relationships["rate"]; ok && rel.Data != nil {
		details.RateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["cost-index"]; ok && rel.Data != nil {
		details.CostIndexID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["parent-rate-adjustment"]; ok && rel.Data != nil {
		details.ParentRateAdjustmentID = rel.Data.ID
	}

	return details
}

func renderRateAdjustmentDetails(cmd *cobra.Command, details rateAdjustmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RateID != "" {
		fmt.Fprintf(out, "Rate ID: %s\n", details.RateID)
	}
	if details.CostIndexID != "" {
		fmt.Fprintf(out, "Cost Index ID: %s\n", details.CostIndexID)
	}
	if details.ParentRateAdjustmentID != "" {
		fmt.Fprintf(out, "Parent Rate Adjustment ID: %s\n", details.ParentRateAdjustmentID)
	}
	if details.ZeroInterceptValue != "" {
		fmt.Fprintf(out, "Zero Intercept Value: %s\n", details.ZeroInterceptValue)
	}
	if details.ZeroInterceptRatio != "" {
		fmt.Fprintf(out, "Zero Intercept Ratio: %s\n", details.ZeroInterceptRatio)
	}
	if details.AdjustmentMin != "" {
		fmt.Fprintf(out, "Adjustment Min: %s\n", details.AdjustmentMin)
	}
	if details.AdjustmentMax != "" {
		fmt.Fprintf(out, "Adjustment Max: %s\n", details.AdjustmentMax)
	}
	fmt.Fprintf(out, "Prevent Rating When Index Value Missing: %t\n", details.PreventRatingWhenIndexValueMissing)

	return nil
}
