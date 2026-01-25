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

type doLineupDispatchFulfillmentClerksCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	LineupDispatch string
}

type lineupDispatchFulfillmentClerkRow struct {
	ID               string `json:"id"`
	LineupDispatchID string `json:"lineup_dispatch_id,omitempty"`
}

func newDoLineupDispatchFulfillmentClerksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup dispatch fulfillment clerk",
		Long: `Create a lineup dispatch fulfillment clerk.

Triggers the asynchronous fulfillment workflow for a specific lineup dispatch.

Required flags:
  --lineup-dispatch   Lineup dispatch ID (required)`,
		Example: `  # Run fulfillment for a dispatch
  xbe do lineup-dispatch-fulfillment-clerks create --lineup-dispatch 123`,
		Args: cobra.NoArgs,
		RunE: runDoLineupDispatchFulfillmentClerksCreate,
	}
	initDoLineupDispatchFulfillmentClerksCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupDispatchFulfillmentClerksCmd.AddCommand(newDoLineupDispatchFulfillmentClerksCreateCmd())
}

func initDoLineupDispatchFulfillmentClerksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-dispatch", "", "Lineup dispatch ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupDispatchFulfillmentClerksCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupDispatchFulfillmentClerksCreateOptions(cmd)
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

	if opts.LineupDispatch == "" {
		err := fmt.Errorf("--lineup-dispatch is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"lineup-dispatch": map[string]any{
			"data": map[string]any{
				"type": "lineup-dispatches",
				"id":   opts.LineupDispatch,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-dispatch-fulfillment-clerks",
			"attributes":    map[string]any{},
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-dispatch-fulfillment-clerks", jsonBody)
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

	row := buildLineupDispatchFulfillmentClerkRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup dispatch fulfillment clerk %s\n", row.ID)
	return nil
}

func parseDoLineupDispatchFulfillmentClerksCreateOptions(cmd *cobra.Command) (doLineupDispatchFulfillmentClerksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupDispatch, _ := cmd.Flags().GetString("lineup-dispatch")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupDispatchFulfillmentClerksCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		LineupDispatch: lineupDispatch,
	}, nil
}

func buildLineupDispatchFulfillmentClerkRowFromSingle(resp jsonAPISingleResponse) lineupDispatchFulfillmentClerkRow {
	row := lineupDispatchFulfillmentClerkRow{ID: resp.Data.ID}

	if rel, ok := resp.Data.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
		row.LineupDispatchID = rel.Data.ID
	}

	return row
}
