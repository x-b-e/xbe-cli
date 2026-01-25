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

type doJobProductionPlanProjectPhaseRevenueItemsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Quantity     string
	ShouldUpdate bool
}

func newDoJobProductionPlanProjectPhaseRevenueItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan project phase revenue item",
		Long: `Update a job production plan project phase revenue item.

Optional:
  --quantity       Planned quantity
  --should-update  Force update on the item

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity
  xbe do job-production-plan-project-phase-revenue-items update 123 --quantity 30

  # Force update
  xbe do job-production-plan-project-phase-revenue-items update 123 --should-update`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanProjectPhaseRevenueItemsUpdate,
	}
	initDoJobProductionPlanProjectPhaseRevenueItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanProjectPhaseRevenueItemsCmd.AddCommand(newDoJobProductionPlanProjectPhaseRevenueItemsUpdateCmd())
}

func initDoJobProductionPlanProjectPhaseRevenueItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().Bool("should-update", false, "Force update on the item")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanProjectPhaseRevenueItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanProjectPhaseRevenueItemsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("should-update") {
		attributes["should-update"] = opts.ShouldUpdate
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-project-phase-revenue-items",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-project-phase-revenue-items/"+opts.ID, jsonBody)
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
		row := buildJobProductionPlanProjectPhaseRevenueItemRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan project phase revenue item %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanProjectPhaseRevenueItemsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanProjectPhaseRevenueItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	shouldUpdate, _ := cmd.Flags().GetBool("should-update")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanProjectPhaseRevenueItemsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Quantity:     quantity,
		ShouldUpdate: shouldUpdate,
	}, nil
}
