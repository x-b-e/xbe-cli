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

type doLineupJobProductionPlansCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	LineupID            string
	JobProductionPlanID string
}

func newDoLineupJobProductionPlansCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup job production plan",
		Long: `Create a lineup job production plan.

Required:
  --lineup                Lineup ID
  --job-production-plan   Job production plan ID`,
		Example: `  # Create a lineup job production plan
  xbe do lineup-job-production-plans create --lineup 123 --job-production-plan 456`,
		RunE: runDoLineupJobProductionPlansCreate,
	}
	initDoLineupJobProductionPlansCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupJobProductionPlansCmd.AddCommand(newDoLineupJobProductionPlansCreateCmd())
}

func initDoLineupJobProductionPlansCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup", "", "Lineup ID")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("lineup")
	_ = cmd.MarkFlagRequired("job-production-plan")
}

func runDoLineupJobProductionPlansCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupJobProductionPlansCreateOptions(cmd)
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

	relationships := map[string]any{
		"lineup": map[string]any{
			"data": map[string]any{
				"type": "lineups",
				"id":   opts.LineupID,
			},
		},
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-job-production-plans",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-job-production-plans", jsonBody)
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
		row := lineupJobProductionPlanRow{
			ID:          resp.Data.ID,
			IsDeletable: boolAttr(resp.Data.Attributes, "is-deletable"),
		}
		if rel, ok := resp.Data.Relationships["lineup"]; ok && rel.Data != nil {
			row.LineupID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup job production plan %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupJobProductionPlansCreateOptions(cmd *cobra.Command) (doLineupJobProductionPlansCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupID, _ := cmd.Flags().GetString("lineup")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupJobProductionPlansCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		LineupID:            lineupID,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}
