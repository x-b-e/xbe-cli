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

type doWorkOrderAssignmentsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	WorkOrderID string
	UserID      string
}

func newDoWorkOrderAssignmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a work order assignment",
		Long: `Update an existing work order assignment.

Arguments:
  <id>    The assignment ID (required)

Flags:
  --work-order   Work order ID
  --user         User ID`,
		Example: `  # Update assignment relationships
  xbe do work-order-assignments update 123 --work-order 456 --user 789

  # Get JSON output
  xbe do work-order-assignments update 123 --user 789 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoWorkOrderAssignmentsUpdate,
	}
	initDoWorkOrderAssignmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrderAssignmentsCmd.AddCommand(newDoWorkOrderAssignmentsUpdateCmd())
}

func initDoWorkOrderAssignmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("work-order", "", "Work order ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoWorkOrderAssignmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoWorkOrderAssignmentsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("work-order") {
		if opts.WorkOrderID == "" {
			relationships["work-order"] = map[string]any{"data": nil}
		} else {
			relationships["work-order"] = map[string]any{
				"data": map[string]any{
					"type": "work-orders",
					"id":   opts.WorkOrderID,
				},
			}
		}
	}
	if cmd.Flags().Changed("user") {
		if opts.UserID == "" {
			relationships["user"] = map[string]any{"data": nil}
		} else {
			relationships["user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.UserID,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "work-order-assignments",
		"id":            opts.ID,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/work-order-assignments/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated work order assignment %s\n", details.ID)
	return nil
}

func parseDoWorkOrderAssignmentsUpdateOptions(cmd *cobra.Command, args []string) (doWorkOrderAssignmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	workOrderID, _ := cmd.Flags().GetString("work-order")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doWorkOrderAssignmentsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		WorkOrderID: workOrderID,
		UserID:      userID,
	}, nil
}
