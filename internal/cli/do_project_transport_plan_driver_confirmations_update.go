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

type doProjectTransportPlanDriverConfirmationsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Status       string
	Note         string
	ConfirmAtMax string
}

func newDoProjectTransportPlanDriverConfirmationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan driver confirmation",
		Long: `Update a project transport plan driver confirmation.

Optional flags:
  --status          Update status (pending, confirmed, rejected, expired, superseded)
  --note            Append a note
  --confirm-at-max  Update confirm-at max timestamp (ISO 8601)`,
		Example: `  # Confirm a driver assignment
  xbe do project-transport-plan-driver-confirmations update 123 --status confirmed

  # Add a note
  xbe do project-transport-plan-driver-confirmations update 123 --note "Reviewed"

  # Update confirmation deadline
  xbe do project-transport-plan-driver-confirmations update 123 --confirm-at-max "2025-01-01T12:00:00Z"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanDriverConfirmationsUpdate,
	}
	initDoProjectTransportPlanDriverConfirmationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanDriverConfirmationsCmd.AddCommand(newDoProjectTransportPlanDriverConfirmationsUpdateCmd())
}

func initDoProjectTransportPlanDriverConfirmationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Status (pending, confirmed, rejected, expired, superseded)")
	cmd.Flags().String("note", "", "Note to append")
	cmd.Flags().String("confirm-at-max", "", "Confirm-at max timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanDriverConfirmationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanDriverConfirmationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("confirm-at-max") {
		attributes["confirm-at-max"] = opts.ConfirmAtMax
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-transport-plan-driver-confirmations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-driver-confirmations/"+opts.ID, jsonBody)
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

	row := buildProjectTransportPlanDriverConfirmationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan driver confirmation %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanDriverConfirmationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanDriverConfirmationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	note, _ := cmd.Flags().GetString("note")
	confirmAtMax, _ := cmd.Flags().GetString("confirm-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanDriverConfirmationsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Status:       status,
		Note:         note,
		ConfirmAtMax: confirmAtMax,
	}, nil
}
