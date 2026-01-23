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

type doTimeSheetLineItemEquipmentRequirementsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	TimeSheetLineItem    string
	EquipmentRequirement string
	IsPrimary            string
}

func newDoTimeSheetLineItemEquipmentRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheet line item equipment requirement",
		Long: `Create a time sheet line item equipment requirement.

Required flags:
  --time-sheet-line-item     Time sheet line item ID (required)
  --equipment-requirement    Equipment requirement ID (required)

Optional flags:
  --is-primary               Mark as primary (true/false)`,
		Example: `  # Create a requirement link
  xbe do time-sheet-line-item-equipment-requirements create \\
    --time-sheet-line-item 123 \\
    --equipment-requirement 456

  # Create and mark as primary
  xbe do time-sheet-line-item-equipment-requirements create \\
    --time-sheet-line-item 123 \\
    --equipment-requirement 456 \\
    --is-primary true

  # JSON output
  xbe do time-sheet-line-item-equipment-requirements create \\
    --time-sheet-line-item 123 \\
    --equipment-requirement 456 \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetLineItemEquipmentRequirementsCreate,
	}
	initDoTimeSheetLineItemEquipmentRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemEquipmentRequirementsCmd.AddCommand(newDoTimeSheetLineItemEquipmentRequirementsCreateCmd())
}

func initDoTimeSheetLineItemEquipmentRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet-line-item", "", "Time sheet line item ID (required)")
	cmd.Flags().String("equipment-requirement", "", "Equipment requirement ID (required)")
	cmd.Flags().String("is-primary", "", "Mark as primary (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemEquipmentRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetLineItemEquipmentRequirementsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TimeSheetLineItem) == "" {
		err := fmt.Errorf("--time-sheet-line-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.EquipmentRequirement) == "" {
		err := fmt.Errorf("--equipment-requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.IsPrimary != "" {
		attributes["is-primary"] = opts.IsPrimary == "true"
	}

	relationships := map[string]any{
		"time-sheet-line-item": map[string]any{
			"data": map[string]any{
				"type": "time-sheet-line-items",
				"id":   opts.TimeSheetLineItem,
			},
		},
		"equipment-requirement": map[string]any{
			"data": map[string]any{
				"type": "equipment-requirements",
				"id":   opts.EquipmentRequirement,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-sheet-line-item-equipment-requirements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-line-item-equipment-requirements", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet line item equipment requirement %s\n", row.ID)
	return nil
}

func parseDoTimeSheetLineItemEquipmentRequirementsCreateOptions(cmd *cobra.Command) (doTimeSheetLineItemEquipmentRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheetLineItem, _ := cmd.Flags().GetString("time-sheet-line-item")
	equipmentRequirement, _ := cmd.Flags().GetString("equipment-requirement")
	isPrimary, _ := cmd.Flags().GetString("is-primary")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetLineItemEquipmentRequirementsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		TimeSheetLineItem:    timeSheetLineItem,
		EquipmentRequirement: equipmentRequirement,
		IsPrimary:            isPrimary,
	}, nil
}

func buildTimeSheetLineItemEquipmentRequirementRowFromSingle(resp jsonAPISingleResponse) timeSheetLineItemEquipmentRequirementRow {
	attrs := resp.Data.Attributes

	row := timeSheetLineItemEquipmentRequirementRow{
		ID:        resp.Data.ID,
		IsPrimary: boolAttr(attrs, "is-primary"),
	}

	if rel, ok := resp.Data.Relationships["time-sheet-line-item"]; ok && rel.Data != nil {
		row.TimeSheetLineItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-requirement"]; ok && rel.Data != nil {
		row.EquipmentRequirementID = rel.Data.ID
	}

	return row
}
