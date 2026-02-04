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

type projectTransportPlanDriverConfirmationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanDriverConfirmationDetails struct {
	ID                           string `json:"id"`
	Status                       string `json:"status,omitempty"`
	Note                         string `json:"note,omitempty"`
	ConfirmAtMax                 string `json:"confirm_at_max,omitempty"`
	ConfirmedAt                  string `json:"confirmed_at,omitempty"`
	Notes                        any    `json:"notes,omitempty"`
	ProjectTransportPlanID       string `json:"project_transport_plan_id,omitempty"`
	ProjectTransportPlanDriverID string `json:"project_transport_plan_driver_id,omitempty"`
	DriverID                     string `json:"driver_id,omitempty"`
	ConfirmedByID                string `json:"confirmed_by_id,omitempty"`
}

func newProjectTransportPlanDriverConfirmationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan driver confirmation details",
		Long: `Show the full details of a project transport plan driver confirmation.

Output Fields:
  ID
  Status
  Note
  Confirm At Max
  Confirmed At
  Notes
  Project Transport Plan ID
  Project Transport Plan Driver ID
  Driver ID
  Confirmed By ID

Arguments:
  <id>    The confirmation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a confirmation
  xbe view project-transport-plan-driver-confirmations show 123

  # JSON output
  xbe view project-transport-plan-driver-confirmations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanDriverConfirmationsShow,
	}
	initProjectTransportPlanDriverConfirmationsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriverConfirmationsCmd.AddCommand(newProjectTransportPlanDriverConfirmationsShowCmd())
}

func initProjectTransportPlanDriverConfirmationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriverConfirmationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanDriverConfirmationsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan driver confirmation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-driver-confirmations/"+id, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectTransportPlanDriverConfirmationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanDriverConfirmationDetails(cmd, details)
}

func parseProjectTransportPlanDriverConfirmationsShowOptions(cmd *cobra.Command) (projectTransportPlanDriverConfirmationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriverConfirmationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanDriverConfirmationDetails(resp jsonAPISingleResponse) projectTransportPlanDriverConfirmationDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanDriverConfirmationDetails{
		ID:           resp.Data.ID,
		Status:       stringAttr(attrs, "status"),
		Note:         stringAttr(attrs, "note"),
		ConfirmAtMax: formatDateTime(stringAttr(attrs, "confirm-at-max")),
		ConfirmedAt:  formatDateTime(stringAttr(attrs, "confirmed-at")),
	}

	if notes, ok := attrs["notes"]; ok {
		details.Notes = notes
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-driver"]; ok && rel.Data != nil {
		details.ProjectTransportPlanDriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["confirmed-by"]; ok && rel.Data != nil {
		details.ConfirmedByID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanDriverConfirmationDetails(cmd *cobra.Command, details projectTransportPlanDriverConfirmationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.ConfirmAtMax != "" {
		fmt.Fprintf(out, "Confirm At Max: %s\n", details.ConfirmAtMax)
	}
	if details.ConfirmedAt != "" {
		fmt.Fprintf(out, "Confirmed At: %s\n", details.ConfirmedAt)
	}
	if formatted := formatJSONValue(details.Notes); formatted != "" {
		fmt.Fprintf(out, "Notes: %s\n", formatted)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.ProjectTransportPlanDriverID != "" {
		fmt.Fprintf(out, "Project Transport Plan Driver ID: %s\n", details.ProjectTransportPlanDriverID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.ConfirmedByID != "" {
		fmt.Fprintf(out, "Confirmed By ID: %s\n", details.ConfirmedByID)
	}

	return nil
}
