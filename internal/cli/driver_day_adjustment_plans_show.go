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

type driverDayAdjustmentPlansShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverDayAdjustmentPlanDetails struct {
	ID               string `json:"id"`
	TruckerID        string `json:"trucker_id,omitempty"`
	TruckerName      string `json:"trucker_name,omitempty"`
	Content          string `json:"content,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	StartAtEffective string `json:"start_at_effective,omitempty"`
}

func newDriverDayAdjustmentPlansShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver day adjustment plan details",
		Long: `Show the full details of a driver day adjustment plan.

Output Fields:
  ID             Plan identifier
  Trucker        Trucker name and ID (if available)
  Start At       Plan start timestamp
  Effective At   Effective start timestamp
  Content        Full plan content

Arguments:
  <id>           The plan ID (required).`,
		Example: `  # View a plan by ID
  xbe view driver-day-adjustment-plans show 123

  # Get plan as JSON
  xbe view driver-day-adjustment-plans show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverDayAdjustmentPlansShow,
	}
	initDriverDayAdjustmentPlansShowFlags(cmd)
	return cmd
}

func init() {
	driverDayAdjustmentPlansCmd.AddCommand(newDriverDayAdjustmentPlansShowCmd())
}

func initDriverDayAdjustmentPlansShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayAdjustmentPlansShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverDayAdjustmentPlansShowOptions(cmd)
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
		return fmt.Errorf("driver day adjustment plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "trucker")
	query.Set("fields[driver-day-adjustment-plans]", "content,start-at,start-at-effective,trucker")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-adjustment-plans/"+id, query)
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

	details := buildDriverDayAdjustmentPlanDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayAdjustmentPlanDetails(cmd, details)
}

func parseDriverDayAdjustmentPlansShowOptions(cmd *cobra.Command) (driverDayAdjustmentPlansShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayAdjustmentPlansShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverDayAdjustmentPlanDetails(resp jsonAPISingleResponse) driverDayAdjustmentPlanDetails {
	attrs := resp.Data.Attributes
	details := driverDayAdjustmentPlanDetails{
		ID:               resp.Data.ID,
		Content:          strings.TrimSpace(stringAttr(attrs, "content")),
		StartAt:          formatDateTime(stringAttr(attrs, "start-at")),
		StartAtEffective: formatDateTime(stringAttr(attrs, "start-at-effective")),
	}

	truckerType := ""
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		truckerType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.TruckerID != "" && truckerType != "" {
		if trucker, ok := included[resourceKey(truckerType, details.TruckerID)]; ok {
			details.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	return details
}

func renderDriverDayAdjustmentPlanDetails(cmd *cobra.Command, details driverDayAdjustmentPlanDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		if details.TruckerName != "" {
			fmt.Fprintf(out, "Trucker: %s (%s)\n", details.TruckerName, details.TruckerID)
		} else {
			fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
		}
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.StartAtEffective != "" {
		fmt.Fprintf(out, "Effective At: %s\n", details.StartAtEffective)
	}
	if details.Content != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Content:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Content)
	}

	return nil
}
