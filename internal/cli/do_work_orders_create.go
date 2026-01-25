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

type doWorkOrdersCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	BrokerID              string
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

func newDoWorkOrdersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new work order",
		Long: `Create a new work order.

Required flags:
  --broker                    Broker ID (required)
  --responsible-party         Responsible party (business unit) ID (required)

Optional flags:
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
		Example: `  # Create a work order
  xbe do work-orders create \
    --broker 123 \
    --responsible-party 456

  # Create with priority and due date
  xbe do work-orders create \
    --broker 123 \
    --responsible-party 456 \
    --priority high \
    --due-date 2024-12-31

  # Get JSON output
  xbe do work-orders create \
    --broker 123 \
    --responsible-party 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoWorkOrdersCreate,
	}
	initDoWorkOrdersCreateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrdersCmd.AddCommand(newDoWorkOrdersCreateCmd())
}

func initDoWorkOrdersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("responsible-party", "", "Responsible party (business unit) ID (required)")
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

func runDoWorkOrdersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoWorkOrdersCreateOptions(cmd)
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

	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ResponsiblePartyID == "" {
		err := fmt.Errorf("--responsible-party is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.Priority != "" {
		attributes["priority"] = opts.Priority
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("estimated-labor-hours") {
		attributes["estimated-labor-hours"] = opts.EstimatedLaborHours
	}
	if cmd.Flags().Changed("estimated-part-cost") {
		attributes["estimated-part-cost"] = opts.EstimatedPartCost
	}
	if opts.DueDate != "" {
		attributes["due-date"] = opts.DueDate
	}
	if opts.SafetyTagStatus != "" {
		attributes["safety-tag-status"] = opts.SafetyTagStatus
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"responsible-party": map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.ResponsiblePartyID,
			},
		},
	}

	if opts.ServiceSiteID != "" {
		relationships["service-site"] = map[string]any{
			"data": map[string]any{
				"type": "service-sites",
				"id":   opts.ServiceSiteID,
			},
		}
	}
	if opts.CustomWorkOrderStatus != "" {
		relationships["custom-work-order-status"] = map[string]any{
			"data": map[string]any{
				"type": "custom-work-order-statuses",
				"id":   opts.CustomWorkOrderStatus,
			},
		}
	}
	if opts.ServiceCodeID != "" {
		relationships["service-code"] = map[string]any{
			"data": map[string]any{
				"type": "work-order-service-codes",
				"id":   opts.ServiceCodeID,
			},
		}
	}

	data := map[string]any{
		"type":          "work-orders",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/work-orders", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created work order %s\n", row.ID)
	return nil
}

func parseDoWorkOrdersCreateOptions(cmd *cobra.Command) (doWorkOrdersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
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

	return doWorkOrdersCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		BrokerID:              brokerID,
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

func buildWorkOrderRowFromSingle(resp jsonAPISingleResponse) workOrderRow {
	attrs := resp.Data.Attributes

	row := workOrderRow{
		ID:                  resp.Data.ID,
		Priority:            stringAttr(attrs, "priority"),
		Status:              stringAttr(attrs, "status"),
		ActualStatus:        stringAttr(attrs, "actual-status"),
		EstimatedLaborHours: floatAttr(attrs, "estimated-labor-hours"),
		EstimatedPartCost:   floatAttr(attrs, "estimated-part-cost"),
		DueDate:             stringAttr(attrs, "due-date"),
		SafetyTagStatus:     stringAttr(attrs, "safety-tag-status"),
		Note:                stringAttr(attrs, "note"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["responsible-party"]; ok && rel.Data != nil {
		row.ResponsiblePartyID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["service-site"]; ok && rel.Data != nil {
		row.ServiceSiteID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["custom-work-order-status"]; ok && rel.Data != nil {
		row.CustomWorkOrderStatusID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["service-code"]; ok && rel.Data != nil {
		row.ServiceCodeID = rel.Data.ID
	}

	return row
}
