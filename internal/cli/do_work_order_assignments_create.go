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

type doWorkOrderAssignmentsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	WorkOrderID string
	UserID      string
}

func newDoWorkOrderAssignmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work order assignment",
		Long: `Create a work order assignment.

Required flags:
  --work-order   Work order ID
  --user         User ID`,
		Example: `  # Assign a user to a work order
  xbe do work-order-assignments create --work-order 123 --user 456

  # Get JSON output
  xbe do work-order-assignments create --work-order 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoWorkOrderAssignmentsCreate,
	}
	initDoWorkOrderAssignmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrderAssignmentsCmd.AddCommand(newDoWorkOrderAssignmentsCreateCmd())
}

func initDoWorkOrderAssignmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("work-order", "", "Work order ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoWorkOrderAssignmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoWorkOrderAssignmentsCreateOptions(cmd)
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

	workOrderID := strings.TrimSpace(opts.WorkOrderID)
	if workOrderID == "" {
		err := fmt.Errorf("--work-order is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	userID := strings.TrimSpace(opts.UserID)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"work-order": map[string]any{
			"data": map[string]any{
				"type": "work-orders",
				"id":   workOrderID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "work-order-assignments",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/work-order-assignments", jsonBody)
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

	details := buildWorkOrderAssignmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created work order assignment %s\n", details.ID)
	return nil
}

func parseDoWorkOrderAssignmentsCreateOptions(cmd *cobra.Command) (doWorkOrderAssignmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	workOrderID, _ := cmd.Flags().GetString("work-order")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doWorkOrderAssignmentsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		WorkOrderID: workOrderID,
		UserID:      userID,
	}, nil
}
