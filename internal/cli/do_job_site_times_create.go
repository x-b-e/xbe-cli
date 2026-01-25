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

type doJobSiteTimesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	JobProductionPlanID string
	UserID              string
	StartAt             string
	EndAt               string
	Description         string
}

func newDoJobSiteTimesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job site time",
		Long: `Create a job site time.

Required flags:
  --job-production-plan  Job production plan ID (required)
  --user                 User ID (required)

Optional flags:
  --start-at              Start timestamp (RFC3339)
  --end-at                End timestamp (RFC3339)
  --description           Description`,
		Example: `  # Create a job site time
  xbe do job-site-times create --job-production-plan 123 --user 456 --start-at 2026-01-23T08:00:00Z --end-at 2026-01-23T10:00:00Z

  # Create with a description
  xbe do job-site-times create --job-production-plan 123 --user 456 --start-at 2026-01-23T08:00:00Z --description "On site"

  # Output as JSON
  xbe do job-site-times create --job-production-plan 123 --user 456 --start-at 2026-01-23T08:00:00Z --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobSiteTimesCreate,
	}
	initDoJobSiteTimesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobSiteTimesCmd.AddCommand(newDoJobSiteTimesCreateCmd())
}

func initDoJobSiteTimesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (RFC3339)")
	cmd.Flags().String("end-at", "", "End timestamp (RFC3339)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("user")
}

func runDoJobSiteTimesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobSiteTimesCreateOptions(cmd)
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

	if opts.JobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-site-times",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-site-times", jsonBody)
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

	row := buildJobSiteTimeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job site time %s\n", row.ID)
	return nil
}

func parseDoJobSiteTimesCreateOptions(cmd *cobra.Command) (doJobSiteTimesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	userID, _ := cmd.Flags().GetString("user")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobSiteTimesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		JobProductionPlanID: jobProductionPlanID,
		UserID:              userID,
		StartAt:             startAt,
		EndAt:               endAt,
		Description:         description,
	}, nil
}
