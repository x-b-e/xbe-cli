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

type doWorkOrdersUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	ResponsiblePartyID    string
	ServiceSiteID         string
	CustomWorkOrderStatus string
	ServiceCodeID         string
	Priority              string
	Status                string
	EstimatedLaborHours   float64
	EstimatedPartCost     float64
	DueDate               string
	SafetyTagStatus       string
	Note                  string
}

func newDoWorkOrdersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a work order",
		Long: `Update a work order.

Optional flags:
  --responsible-party         Responsible party (business unit) ID
  --service-site              Service site ID
  --custom-work-order-status  Custom work order status ID
  --service-code              Service code ID
  --priority                  Priority level
  --status                    Status
  --estimated-labor-hours     Estimated labor hours
  --estimated-part-cost       Estimated part cost
  --due-date                  Due date (ISO 8601)
  --safety-tag-status         Safety tag status
  --note                      Note`,
		Example: `  # Update status
  xbe do work-orders update 123 --status completed

  # Update priority
  xbe do work-orders update 123 --priority high

  # Update estimated hours
  xbe do work-orders update 123 --estimated-labor-hours 8.5`,
		Args: cobra.ExactArgs(1),
		RunE: runDoWorkOrdersUpdate,
	}
	initDoWorkOrdersUpdateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrdersCmd.AddCommand(newDoWorkOrdersUpdateCmd())
}

func initDoWorkOrdersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("responsible-party", "", "Responsible party (business unit) ID")
	cmd.Flags().String("service-site", "", "Service site ID")
	cmd.Flags().String("custom-work-order-status", "", "Custom work order status ID")
	cmd.Flags().String("service-code", "", "Service code ID")
	cmd.Flags().String("priority", "", "Priority level")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().Float64("estimated-labor-hours", 0, "Estimated labor hours")
	cmd.Flags().Float64("estimated-part-cost", 0, "Estimated part cost")
	cmd.Flags().String("due-date", "", "Due date (ISO 8601)")
	cmd.Flags().String("safety-tag-status", "", "Safety tag status")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoWorkOrdersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoWorkOrdersUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("priority") {
		attributes["priority"] = opts.Priority
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("estimated-labor-hours") {
		attributes["estimated-labor-hours"] = opts.EstimatedLaborHours
	}
	if cmd.Flags().Changed("estimated-part-cost") {
		attributes["estimated-part-cost"] = opts.EstimatedPartCost
	}
	if cmd.Flags().Changed("due-date") {
		attributes["due-date"] = opts.DueDate
	}
	if cmd.Flags().Changed("safety-tag-status") {
		attributes["safety-tag-status"] = opts.SafetyTagStatus
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	if cmd.Flags().Changed("responsible-party") {
		if opts.ResponsiblePartyID == "" {
			relationships["responsible-party"] = map[string]any{"data": nil}
		} else {
			relationships["responsible-party"] = map[string]any{
				"data": map[string]any{
					"type": "business-units",
					"id":   opts.ResponsiblePartyID,
				},
			}
		}
	}
	if cmd.Flags().Changed("service-site") {
		if opts.ServiceSiteID == "" {
			relationships["service-site"] = map[string]any{"data": nil}
		} else {
			relationships["service-site"] = map[string]any{
				"data": map[string]any{
					"type": "service-sites",
					"id":   opts.ServiceSiteID,
				},
			}
		}
	}
	if cmd.Flags().Changed("custom-work-order-status") {
		if opts.CustomWorkOrderStatus == "" {
			relationships["custom-work-order-status"] = map[string]any{"data": nil}
		} else {
			relationships["custom-work-order-status"] = map[string]any{
				"data": map[string]any{
					"type": "custom-work-order-statuses",
					"id":   opts.CustomWorkOrderStatus,
				},
			}
		}
	}
	if cmd.Flags().Changed("service-code") {
		if opts.ServiceCodeID == "" {
			relationships["service-code"] = map[string]any{"data": nil}
		} else {
			relationships["service-code"] = map[string]any{
				"data": map[string]any{
					"type": "work-order-service-codes",
					"id":   opts.ServiceCodeID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "work-orders",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/work-orders/"+opts.ID, jsonBody)
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

	row := buildWorkOrderRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated work order %s\n", row.ID)
	return nil
}

func parseDoWorkOrdersUpdateOptions(cmd *cobra.Command, args []string) (doWorkOrdersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	responsiblePartyID, _ := cmd.Flags().GetString("responsible-party")
	serviceSiteID, _ := cmd.Flags().GetString("service-site")
	customWorkOrderStatus, _ := cmd.Flags().GetString("custom-work-order-status")
	serviceCodeID, _ := cmd.Flags().GetString("service-code")
	priority, _ := cmd.Flags().GetString("priority")
	status, _ := cmd.Flags().GetString("status")
	estimatedLaborHours, _ := cmd.Flags().GetFloat64("estimated-labor-hours")
	estimatedPartCost, _ := cmd.Flags().GetFloat64("estimated-part-cost")
	dueDate, _ := cmd.Flags().GetString("due-date")
	safetyTagStatus, _ := cmd.Flags().GetString("safety-tag-status")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doWorkOrdersUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		ResponsiblePartyID:    responsiblePartyID,
		ServiceSiteID:         serviceSiteID,
		CustomWorkOrderStatus: customWorkOrderStatus,
		ServiceCodeID:         serviceCodeID,
		Priority:              priority,
		Status:                status,
		EstimatedLaborHours:   estimatedLaborHours,
		EstimatedPartCost:     estimatedPartCost,
		DueDate:               dueDate,
		SafetyTagStatus:       safetyTagStatus,
		Note:                  note,
	}, nil
}
