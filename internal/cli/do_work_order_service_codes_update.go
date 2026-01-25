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

type doWorkOrderServiceCodesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Code        string
	Description string
}

func newDoWorkOrderServiceCodesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing work order service code",
		Long: `Update an existing work order service code.

Provide the service code ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --code         Service code value
  --description  Description of the service code`,
		Example: `  # Update code
  xbe do work-order-service-codes update 123 --code "HAUL-1"

  # Update description
  xbe do work-order-service-codes update 123 --description "Updated description"

  # Update multiple fields
  xbe do work-order-service-codes update 123 --code "SPREAD" --description "Spreading service"

  # Get JSON output
  xbe do work-order-service-codes update 123 --code "UPDATED" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoWorkOrderServiceCodesUpdate,
	}
	initDoWorkOrderServiceCodesUpdateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrderServiceCodesCmd.AddCommand(newDoWorkOrderServiceCodesUpdateCmd())
}

func initDoWorkOrderServiceCodesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Service code value")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoWorkOrderServiceCodesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoWorkOrderServiceCodesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("code") {
		attributes["code"] = opts.Code
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --code, --description")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "work-order-service-codes",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/work-order-service-codes/"+opts.ID, jsonBody)
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

	row := buildWorkOrderServiceCodeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated work order service code %s (%s)\n", row.ID, row.Code)
	return nil
}

func parseDoWorkOrderServiceCodesUpdateOptions(cmd *cobra.Command, args []string) (doWorkOrderServiceCodesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doWorkOrderServiceCodesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Code:        code,
		Description: description,
	}, nil
}
