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

type doShiftCountersCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	StartAtMin string
}

type shiftCounterDetails struct {
	ID         string `json:"id"`
	StartAtMin string `json:"start_at_min,omitempty"`
	Count      *int   `json:"count,omitempty"`
}

func newDoShiftCountersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a shift counter",
		Long: `Create a shift counter.

Optional flags:
  --start-at-min  Minimum shift start timestamp (ISO 8601)`,
		Example: `  # Count accepted shifts (default start)
  xbe do shift-counters create

  # Count accepted shifts after a date
  xbe do shift-counters create --start-at-min 2025-01-01T00:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoShiftCountersCreate,
	}
	initDoShiftCountersCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftCountersCmd.AddCommand(newDoShiftCountersCreateCmd())
}

func initDoShiftCountersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at-min", "", "Minimum shift start timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftCountersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftCountersCreateOptions(cmd)
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.StartAtMin) != "" {
		attributes["start-at-min"] = opts.StartAtMin
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "shift-counters",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/shift-counters", jsonBody)
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

	details := buildShiftCounterDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderShiftCounterDetails(cmd, details)
}

func parseDoShiftCountersCreateOptions(cmd *cobra.Command) (doShiftCountersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftCountersCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		StartAtMin: startAtMin,
	}, nil
}

func buildShiftCounterDetails(resp jsonAPISingleResponse) shiftCounterDetails {
	attrs := resp.Data.Attributes
	details := shiftCounterDetails{
		ID:         resp.Data.ID,
		StartAtMin: formatDateTime(stringAttr(attrs, "start-at-min")),
	}

	if value, ok := attrs["count"]; ok && value != nil {
		count := intAttr(attrs, "count")
		details.Count = &count
	}

	return details
}

func renderShiftCounterDetails(cmd *cobra.Command, details shiftCounterDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StartAtMin != "" {
		fmt.Fprintf(out, "Start At Min: %s\n", details.StartAtMin)
	} else {
		fmt.Fprintln(out, "Start At Min: (default)")
	}
	if details.Count != nil {
		fmt.Fprintf(out, "Count: %d\n", *details.Count)
	} else {
		fmt.Fprintln(out, "Count: (none)")
	}

	return nil
}
