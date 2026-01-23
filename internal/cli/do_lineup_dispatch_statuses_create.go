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

type doLineupDispatchStatusesCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	BrokerID string
	Window   string
	Date     string
}

type lineupDispatchStatusRow struct {
	ID         string   `json:"id"`
	BrokerID   string   `json:"broker_id,omitempty"`
	Window     string   `json:"window,omitempty"`
	Date       string   `json:"date,omitempty"`
	OfferedPct *float64 `json:"offered_pct,omitempty"`
}

func newDoLineupDispatchStatusesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup dispatch status",
		Long: `Create a lineup dispatch status.

Computes the offered tender percentage for a broker and lineup window.

Required flags:
  --broker   Broker ID
  --window   Lineup window (day or night)
  --date     Lineup date (YYYY-MM-DD)`,
		Example: `  # Check lineup dispatch status for a day window
  xbe do lineup-dispatch-statuses create --broker 123 --window day --date 2025-01-23

  # JSON output
  xbe do lineup-dispatch-statuses create --broker 123 --window night --date 2025-01-23 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupDispatchStatusesCreate,
	}
	initDoLineupDispatchStatusesCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupDispatchStatusesCmd.AddCommand(newDoLineupDispatchStatusesCreateCmd())
}

func initDoLineupDispatchStatusesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("window", "", "Lineup window (day or night) (required)")
	cmd.Flags().String("date", "", "Lineup date (YYYY-MM-DD) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupDispatchStatusesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupDispatchStatusesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Window == "" {
		err := fmt.Errorf("--window is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Date == "" {
		err := fmt.Errorf("--date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"window": opts.Window,
		"date":   opts.Date,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-dispatch-statuses",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-dispatch-statuses", jsonBody)
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

	row := buildLineupDispatchStatusRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderLineupDispatchStatus(cmd, row)
}

func parseDoLineupDispatchStatusesCreateOptions(cmd *cobra.Command) (doLineupDispatchStatusesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	window, _ := cmd.Flags().GetString("window")
	date, _ := cmd.Flags().GetString("date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupDispatchStatusesCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		BrokerID: brokerID,
		Window:   window,
		Date:     date,
	}, nil
}

func buildLineupDispatchStatusRowFromSingle(resp jsonAPISingleResponse) lineupDispatchStatusRow {
	row := lineupDispatchStatusRow{
		ID:         resp.Data.ID,
		Window:     stringAttr(resp.Data.Attributes, "window"),
		Date:       stringAttr(resp.Data.Attributes, "date"),
		OfferedPct: floatAttrPointer(resp.Data.Attributes, "offered-pct"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}

func renderLineupDispatchStatus(cmd *cobra.Command, row lineupDispatchStatusRow) error {
	out := cmd.OutOrStdout()
	offered := formatPct(row.OfferedPct)
	if offered == "" {
		offered = "n/a"
	}

	fmt.Fprintf(out, "Lineup dispatch status %s\n", row.ID)
	fmt.Fprintf(out, "Broker: %s\n", row.BrokerID)
	fmt.Fprintf(out, "Window: %s\n", row.Window)
	fmt.Fprintf(out, "Date: %s\n", row.Date)
	fmt.Fprintf(out, "Offered: %s\n", offered)
	return nil
}
