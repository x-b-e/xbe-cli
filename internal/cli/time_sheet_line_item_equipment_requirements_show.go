package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeSheetLineItemEquipmentRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetLineItemEquipmentRequirementDetails struct {
	ID                     string `json:"id"`
	TimeSheetLineItemID    string `json:"time_sheet_line_item_id,omitempty"`
	EquipmentRequirementID string `json:"equipment_requirement_id,omitempty"`
	IsPrimary              bool   `json:"is_primary"`
}

func newTimeSheetLineItemEquipmentRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet line item equipment requirement details",
		Long: `Show the full details of a time sheet line item equipment requirement.

Output Fields:
  ID
  Time Sheet Line Item ID
  Equipment Requirement ID
  Is Primary

Arguments:
  <id>    Time sheet line item equipment requirement ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet line item equipment requirement
  xbe view time-sheet-line-item-equipment-requirements show 123

  # JSON output
  xbe view time-sheet-line-item-equipment-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetLineItemEquipmentRequirementsShow,
	}
	initTimeSheetLineItemEquipmentRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetLineItemEquipmentRequirementsCmd.AddCommand(newTimeSheetLineItemEquipmentRequirementsShowCmd())
}

func initTimeSheetLineItemEquipmentRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetLineItemEquipmentRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetLineItemEquipmentRequirementsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("time sheet line item equipment requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-line-item-equipment-requirements]", "is-primary,time-sheet-line-item,equipment-requirement")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-line-item-equipment-requirements/"+id, query)
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

	details := buildTimeSheetLineItemEquipmentRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetLineItemEquipmentRequirementDetails(cmd, details)
}

func parseTimeSheetLineItemEquipmentRequirementsShowOptions(cmd *cobra.Command) (timeSheetLineItemEquipmentRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetLineItemEquipmentRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetLineItemEquipmentRequirementDetails(resp jsonAPISingleResponse) timeSheetLineItemEquipmentRequirementDetails {
	details := timeSheetLineItemEquipmentRequirementDetails{
		ID:        resp.Data.ID,
		IsPrimary: boolAttr(resp.Data.Attributes, "is-primary"),
	}

	if rel, ok := resp.Data.Relationships["time-sheet-line-item"]; ok && rel.Data != nil {
		details.TimeSheetLineItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-requirement"]; ok && rel.Data != nil {
		details.EquipmentRequirementID = rel.Data.ID
	}

	return details
}

func renderTimeSheetLineItemEquipmentRequirementDetails(cmd *cobra.Command, details timeSheetLineItemEquipmentRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetLineItemID != "" {
		fmt.Fprintf(out, "Time Sheet Line Item ID: %s\n", details.TimeSheetLineItemID)
	}
	if details.EquipmentRequirementID != "" {
		fmt.Fprintf(out, "Equipment Requirement ID: %s\n", details.EquipmentRequirementID)
	}
	if details.IsPrimary {
		fmt.Fprintln(out, "Is Primary: yes")
	} else {
		fmt.Fprintln(out, "Is Primary: no")
	}

	return nil
}
