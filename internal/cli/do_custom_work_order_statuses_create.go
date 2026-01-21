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

type doCustomWorkOrderStatusesCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Label         string
	Description   string
	ColorHex      string
	PrimaryStatus string
	Broker        string
}

func newDoCustomWorkOrderStatusesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new custom work order status",
		Long: `Create a new custom work order status.

Required flags:
  --label           The display label (required)
  --primary-status  The underlying status (required)
  --broker          The broker ID (required)

Optional flags:
  --description  Description of the status
  --color-hex    Hex color code (e.g., #FF5500)`,
		Example: `  # Create a basic custom work order status
  xbe do custom-work-order-statuses create --label "Awaiting Parts" --primary-status pending --broker 123

  # Create with color and description
  xbe do custom-work-order-statuses create --label "In Review" --primary-status in_progress --broker 123 --color-hex "#3366CC" --description "Work order is being reviewed"

  # Get JSON output
  xbe do custom-work-order-statuses create --label "Test" --primary-status pending --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCustomWorkOrderStatusesCreate,
	}
	initDoCustomWorkOrderStatusesCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomWorkOrderStatusesCmd.AddCommand(newDoCustomWorkOrderStatusesCreateCmd())
}

func initDoCustomWorkOrderStatusesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("label", "", "Display label (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("color-hex", "", "Hex color code (e.g., #FF5500)")
	cmd.Flags().String("primary-status", "", "Primary status (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomWorkOrderStatusesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomWorkOrderStatusesCreateOptions(cmd)
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

	if opts.Label == "" {
		err := fmt.Errorf("--label is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.PrimaryStatus == "" {
		err := fmt.Errorf("--primary-status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"label":          opts.Label,
		"primary-status": opts.PrimaryStatus,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "custom-work-order-statuses",
			"attributes": attributes,
			"relationships": map[string]any{
				"broker": map[string]any{
					"data": map[string]any{
						"type": "brokers",
						"id":   opts.Broker,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/custom-work-order-statuses", jsonBody)
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

	row := buildCustomWorkOrderStatusRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created custom work order status %s (%s)\n", row.ID, row.Label)
	return nil
}

func parseDoCustomWorkOrderStatusesCreateOptions(cmd *cobra.Command) (doCustomWorkOrderStatusesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	label, _ := cmd.Flags().GetString("label")
	description, _ := cmd.Flags().GetString("description")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	primaryStatus, _ := cmd.Flags().GetString("primary-status")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomWorkOrderStatusesCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Label:         label,
		Description:   description,
		ColorHex:      colorHex,
		PrimaryStatus: primaryStatus,
		Broker:        broker,
	}, nil
}

func buildCustomWorkOrderStatusRowFromSingle(resp jsonAPISingleResponse) customWorkOrderStatusRow {
	attrs := resp.Data.Attributes

	row := customWorkOrderStatusRow{
		ID:            resp.Data.ID,
		Label:         stringAttr(attrs, "label"),
		Description:   stringAttr(attrs, "description"),
		ColorHex:      stringAttr(attrs, "color-hex"),
		PrimaryStatus: stringAttr(attrs, "primary-status"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
