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

type doDriverDayAdjustmentPlansCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Trucker string
	Content string
	StartAt string
}

func newDoDriverDayAdjustmentPlansCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day adjustment plan",
		Long: `Create a driver day adjustment plan.

Required:
  --trucker    Trucker ID
  --content    Plan content

Optional:
  --start-at   Plan start timestamp (ISO 8601)`,
		Example: `  # Create a plan with required fields
  xbe do driver-day-adjustment-plans create --trucker 123 --content "Adjusted schedule"

  # Create a plan with a start timestamp
  xbe do driver-day-adjustment-plans create --trucker 123 --content "Adjusted schedule" --start-at "2025-01-15T08:00:00Z"`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayAdjustmentPlansCreate,
	}
	initDoDriverDayAdjustmentPlansCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayAdjustmentPlansCmd.AddCommand(newDoDriverDayAdjustmentPlansCreateCmd())
}

func initDoDriverDayAdjustmentPlansCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("content", "", "Plan content")
	cmd.Flags().String("start-at", "", "Plan start timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("trucker")
	_ = cmd.MarkFlagRequired("content")
}

func runDoDriverDayAdjustmentPlansCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayAdjustmentPlansCreateOptions(cmd)
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

	attributes := map[string]any{
		"content": opts.Content,
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-day-adjustment-plans",
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

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-adjustment-plans", jsonBody)
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

	if opts.JSON {
		row := driverDayAdjustmentPlanRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver day adjustment plan %s\n", resp.Data.ID)
	return nil
}

func parseDoDriverDayAdjustmentPlansCreateOptions(cmd *cobra.Command) (doDriverDayAdjustmentPlansCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	content, _ := cmd.Flags().GetString("content")
	startAt, _ := cmd.Flags().GetString("start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayAdjustmentPlansCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Trucker: trucker,
		Content: content,
		StartAt: startAt,
	}, nil
}
