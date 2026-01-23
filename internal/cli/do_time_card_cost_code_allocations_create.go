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

type doTimeCardCostCodeAllocationsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	TimeCard string
	Details  string
}

func newDoTimeCardCostCodeAllocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card cost code allocation",
		Long: `Create a time card cost code allocation.

Required flags:
  --time-card  Time card ID
  --details    Allocation details as JSON array

Details JSON format:
  [
    {"cost_code_id":123,"percentage":0.5},
    {"cost_code_id":456,"percentage":0.5,"project_cost_classification_id":789}
  ]

Percentages must sum to 1.0 (100%).

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a time card cost code allocation
  xbe do time-card-cost-code-allocations create \\
    --time-card 123 \\
    --details '[{"cost_code_id":1,"percentage":0.6},{"cost_code_id":2,"percentage":0.4}]'`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardCostCodeAllocationsCreate,
	}
	initDoTimeCardCostCodeAllocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardCostCodeAllocationsCmd.AddCommand(newDoTimeCardCostCodeAllocationsCreateCmd())
}

func initDoTimeCardCostCodeAllocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID")
	cmd.Flags().String("details", "", "Allocation details JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardCostCodeAllocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardCostCodeAllocationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TimeCard) == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Details) == "" {
		err := fmt.Errorf("--details is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var details []map[string]any
	if err := json.Unmarshal([]byte(opts.Details), &details); err != nil {
		err = fmt.Errorf("invalid details JSON: %w", err)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"details": details,
	}
	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   opts.TimeCard,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-card-cost-code-allocations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-cost-code-allocations", jsonBody)
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

	row := timeCardCostCodeAllocationRow{ID: resp.Data.ID}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card cost code allocation %s\n", row.ID)
	return nil
}

func parseDoTimeCardCostCodeAllocationsCreateOptions(cmd *cobra.Command) (doTimeCardCostCodeAllocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	details, _ := cmd.Flags().GetString("details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardCostCodeAllocationsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		TimeCard: timeCard,
		Details:  details,
	}, nil
}
