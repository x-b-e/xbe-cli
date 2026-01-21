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

type doCustomWorkOrderStatusesUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Label         string
	Description   string
	ColorHex      string
	PrimaryStatus string
}

func newDoCustomWorkOrderStatusesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing custom work order status",
		Long: `Update an existing custom work order status.

Provide the status ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --label           The display label
  --description     Description of the status
  --color-hex       Hex color code
  --primary-status  The underlying status`,
		Example: `  # Update label
  xbe do custom-work-order-statuses update 123 --label "Parts Ordered"

  # Update color
  xbe do custom-work-order-statuses update 123 --color-hex "#FF5500"

  # Update multiple fields
  xbe do custom-work-order-statuses update 123 --label "In Review" --primary-status in_progress

  # Get JSON output
  xbe do custom-work-order-statuses update 123 --label "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomWorkOrderStatusesUpdate,
	}
	initDoCustomWorkOrderStatusesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomWorkOrderStatusesCmd.AddCommand(newDoCustomWorkOrderStatusesUpdateCmd())
}

func initDoCustomWorkOrderStatusesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("label", "", "Display label")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("color-hex", "", "Hex color code")
	cmd.Flags().String("primary-status", "", "Primary status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomWorkOrderStatusesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomWorkOrderStatusesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("label") {
		attributes["label"] = opts.Label
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("color-hex") {
		attributes["color-hex"] = opts.ColorHex
	}
	if cmd.Flags().Changed("primary-status") {
		attributes["primary-status"] = opts.PrimaryStatus
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --label, --description, --color-hex, --primary-status")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "custom-work-order-statuses",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/custom-work-order-statuses/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated custom work order status %s (%s)\n", row.ID, row.Label)
	return nil
}

func parseDoCustomWorkOrderStatusesUpdateOptions(cmd *cobra.Command, args []string) (doCustomWorkOrderStatusesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	label, _ := cmd.Flags().GetString("label")
	description, _ := cmd.Flags().GetString("description")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	primaryStatus, _ := cmd.Flags().GetString("primary-status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomWorkOrderStatusesUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Label:         label,
		Description:   description,
		ColorHex:      colorHex,
		PrimaryStatus: primaryStatus,
	}, nil
}
