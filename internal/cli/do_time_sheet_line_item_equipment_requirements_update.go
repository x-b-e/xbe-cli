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

type doTimeSheetLineItemEquipmentRequirementsUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	IsPrimary string
}

func newDoTimeSheetLineItemEquipmentRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time sheet line item equipment requirement",
		Long: `Update a time sheet line item equipment requirement.

Updatable fields:
  --is-primary    Mark as primary (true/false)

Arguments:
  <id>    Time sheet line item equipment requirement ID (required).`,
		Example: `  # Mark a requirement as primary
  xbe do time-sheet-line-item-equipment-requirements update 123 --is-primary true

  # JSON output
  xbe do time-sheet-line-item-equipment-requirements update 123 --is-primary false --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetLineItemEquipmentRequirementsUpdate,
	}
	initDoTimeSheetLineItemEquipmentRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemEquipmentRequirementsCmd.AddCommand(newDoTimeSheetLineItemEquipmentRequirementsUpdateCmd())
}

func initDoTimeSheetLineItemEquipmentRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("is-primary", "", "Mark as primary (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemEquipmentRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetLineItemEquipmentRequirementsUpdateOptions(cmd, args)
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
	if opts.IsPrimary != "" {
		attributes["is-primary"] = opts.IsPrimary == "true"
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field is required to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         opts.ID,
			"type":       "time-sheet-line-item-equipment-requirements",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheet-line-item-equipment-requirements/"+opts.ID, jsonBody)
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

	row := buildTimeSheetLineItemEquipmentRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet line item equipment requirement %s\n", row.ID)
	return nil
}

func parseDoTimeSheetLineItemEquipmentRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doTimeSheetLineItemEquipmentRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isPrimary, _ := cmd.Flags().GetString("is-primary")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetLineItemEquipmentRequirementsUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		IsPrimary: isPrimary,
	}, nil
}
