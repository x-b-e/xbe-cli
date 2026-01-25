package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type driverDayAdjustmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverDayAdjustmentDetails struct {
	ID                   string `json:"id"`
	DriverDayID          string `json:"driver_day_id,omitempty"`
	TruckerID            string `json:"trucker_id,omitempty"`
	DriverID             string `json:"driver_id,omitempty"`
	Amount               string `json:"amount,omitempty"`
	AmountExplicit       string `json:"amount_explicit,omitempty"`
	AmountGenerated      string `json:"amount_generated,omitempty"`
	ExpressionGenerated  string `json:"expression_generated,omitempty"`
	PlanContent          string `json:"plan_content,omitempty"`
	DriverDayDescription string `json:"driver_day_description,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

func newDriverDayAdjustmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver day adjustment details",
		Long: `Show the full details of a driver day adjustment.

Output Fields:
  ID                     Adjustment identifier
  Driver Day             Driver day ID
  Trucker                Trucker ID
  Driver                 Driver user ID
  Amount                 Final adjustment amount
  Amount Explicit        Explicit adjustment amount
  Amount Generated       Generated adjustment amount
  Expression Generated   Generated expression
  Created At             Created timestamp
  Updated At             Updated timestamp
  Plan Content           Adjustment plan content
  Driver Day Description Driver day description snapshot

Arguments:
  <id>    The adjustment ID (required). You can find IDs using the list command.`,
		Example: `  # Show an adjustment
  xbe view driver-day-adjustments show 123

  # Get JSON output
  xbe view driver-day-adjustments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverDayAdjustmentsShow,
	}
	initDriverDayAdjustmentsShowFlags(cmd)
	return cmd
}

func init() {
	driverDayAdjustmentsCmd.AddCommand(newDriverDayAdjustmentsShowCmd())
}

func initDriverDayAdjustmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayAdjustmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverDayAdjustmentsShowOptions(cmd)
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
		return fmt.Errorf("driver day adjustment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-adjustments/"+id, nil)
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

	details := buildDriverDayAdjustmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayAdjustmentDetails(cmd, details)
}

func parseDriverDayAdjustmentsShowOptions(cmd *cobra.Command) (driverDayAdjustmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayAdjustmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverDayAdjustmentDetails(resp jsonAPISingleResponse) driverDayAdjustmentDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := driverDayAdjustmentDetails{
		ID:                   resource.ID,
		Amount:               stringAttr(attrs, "amount"),
		AmountExplicit:       stringAttr(attrs, "amount-explicit"),
		AmountGenerated:      stringAttr(attrs, "amount-generated"),
		ExpressionGenerated:  stringAttr(attrs, "expression-generated"),
		PlanContent:          stringAttr(attrs, "plan-content"),
		DriverDayDescription: stringAttr(attrs, "driver-day-description"),
		CreatedAt:            formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:            formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}

	return details
}

func renderDriverDayAdjustmentDetails(cmd *cobra.Command, details driverDayAdjustmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day: %s\n", details.DriverDayID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverID)
	}
	if details.Amount != "" {
		fmt.Fprintf(out, "Amount: %s\n", details.Amount)
	}
	if details.AmountExplicit != "" {
		fmt.Fprintf(out, "Amount Explicit: %s\n", details.AmountExplicit)
	}
	if details.AmountGenerated != "" {
		fmt.Fprintf(out, "Amount Generated: %s\n", details.AmountGenerated)
	}
	if details.ExpressionGenerated != "" {
		fmt.Fprintf(out, "Expression Generated: %s\n", details.ExpressionGenerated)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.PlanContent != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Plan Content:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.PlanContent)
	}
	if details.DriverDayDescription != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Driver Day Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.DriverDayDescription)
	}

	return nil
}
