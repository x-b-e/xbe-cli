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

type doJobProductionPlanInspectorsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	JobProductionPlanID string
	UserID              string
}

func newDoJobProductionPlanInspectorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan inspector",
		Long: `Create a job production plan inspector.

Required:
  --job-production-plan-id  Job production plan ID
  --user                    User ID`,
		Example: `  # Create a job production plan inspector
  xbe do job-production-plan-inspectors create --job-production-plan-id 123 --user 456`,
		RunE: runDoJobProductionPlanInspectorsCreate,
	}
	initDoJobProductionPlanInspectorsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanInspectorsCmd.AddCommand(newDoJobProductionPlanInspectorsCreateCmd())
}

func initDoJobProductionPlanInspectorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-id", "", "Job production plan ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan-id")
	_ = cmd.MarkFlagRequired("user")
}

func runDoJobProductionPlanInspectorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanInspectorsCreateOptions(cmd)
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
		"job-production-plan-id": opts.JobProductionPlanID,
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-inspectors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-inspectors", jsonBody)
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
		row := jobProductionPlanInspectorRow{
			ID:                  resp.Data.ID,
			JobProductionPlanID: stringAttr(resp.Data.Attributes, "job-production-plan-id"),
		}
		if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan inspector %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanInspectorsCreateOptions(cmd *cobra.Command) (doJobProductionPlanInspectorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanInspectorsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		JobProductionPlanID: jobProductionPlanID,
		UserID:              userID,
	}, nil
}
