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

type doJobProductionPlanCancellationReasonTypesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Name        string
	Description string
}

func newDoJobProductionPlanCancellationReasonTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing cancellation reason type",
		Long: `Update an existing job production plan cancellation reason type.

Provide the type ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Note: The slug cannot be changed after creation.

Updatable fields:
  --name         The type name
  --description  Type description`,
		Example: `  # Update name
  xbe do job-production-plan-cancellation-reason-types update 123 --name "Updated Name"

  # Update description
  xbe do job-production-plan-cancellation-reason-types update 123 --description "New description"

  # Get JSON output
  xbe do job-production-plan-cancellation-reason-types update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanCancellationReasonTypesUpdate,
	}
	initDoJobProductionPlanCancellationReasonTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCancellationReasonTypesCmd.AddCommand(newDoJobProductionPlanCancellationReasonTypesUpdateCmd())
}

func initDoJobProductionPlanCancellationReasonTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Type name")
	cmd.Flags().String("description", "", "Type description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanCancellationReasonTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanCancellationReasonTypesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --description")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-cancellation-reason-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-cancellation-reason-types/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanCancellationReasonTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated cancellation reason type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoJobProductionPlanCancellationReasonTypesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanCancellationReasonTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCancellationReasonTypesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Name:        name,
		Description: description,
	}, nil
}
