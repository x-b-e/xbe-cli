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

type crewAssignmentConfirmationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type crewAssignmentConfirmationDetails struct {
	ID                         string `json:"id"`
	AssignmentConfirmationUUID string `json:"assignment_confirmation_uuid,omitempty"`
	CrewRequirementID          string `json:"crew_requirement_id,omitempty"`
	ResourceType               string `json:"resource_type,omitempty"`
	ResourceID                 string `json:"resource_id,omitempty"`
	ConfirmedByID              string `json:"confirmed_by_id,omitempty"`
	StartAt                    string `json:"start_at,omitempty"`
	ConfirmedAt                string `json:"confirmed_at,omitempty"`
	Note                       string `json:"note,omitempty"`
	IsExplicit                 bool   `json:"is_explicit"`
}

func newCrewAssignmentConfirmationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show crew assignment confirmation details",
		Long: `Show the full details of a crew assignment confirmation.

Output Fields:
  ID                          Confirmation identifier
  Assignment Confirmation UUID
  Crew Requirement ID
  Resource Type / Resource ID
  Confirmed By (user ID)
  Start At
  Confirmed At
  Note
  Is Explicit

Arguments:
  <id>    The confirmation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a confirmation
  xbe view crew-assignment-confirmations show 123

  # Get JSON output
  xbe view crew-assignment-confirmations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCrewAssignmentConfirmationsShow,
	}
	initCrewAssignmentConfirmationsShowFlags(cmd)
	return cmd
}

func init() {
	crewAssignmentConfirmationsCmd.AddCommand(newCrewAssignmentConfirmationsShowCmd())
}

func initCrewAssignmentConfirmationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewAssignmentConfirmationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCrewAssignmentConfirmationsShowOptions(cmd)
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
		return fmt.Errorf("crew assignment confirmation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/crew-assignment-confirmations/"+id, nil)
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

	details := buildCrewAssignmentConfirmationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCrewAssignmentConfirmationDetails(cmd, details)
}

func parseCrewAssignmentConfirmationsShowOptions(cmd *cobra.Command) (crewAssignmentConfirmationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewAssignmentConfirmationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCrewAssignmentConfirmationDetails(resp jsonAPISingleResponse) crewAssignmentConfirmationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := crewAssignmentConfirmationDetails{
		ID:                         resource.ID,
		AssignmentConfirmationUUID: stringAttr(attrs, "assignment-confirmation-uuid"),
		StartAt:                    formatDateTime(stringAttr(attrs, "start-at")),
		ConfirmedAt:                formatDateTime(stringAttr(attrs, "confirmed-at")),
		Note:                       stringAttr(attrs, "note"),
		IsExplicit:                 boolAttr(attrs, "is-explicit"),
	}

	if rel, ok := resource.Relationships["crew-requirement"]; ok && rel.Data != nil {
		details.CrewRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		details.ResourceType = rel.Data.Type
		details.ResourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["confirmed-by"]; ok && rel.Data != nil {
		details.ConfirmedByID = rel.Data.ID
	}

	return details
}

func renderCrewAssignmentConfirmationDetails(cmd *cobra.Command, details crewAssignmentConfirmationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.AssignmentConfirmationUUID != "" {
		fmt.Fprintf(out, "Assignment Confirmation UUID: %s\n", details.AssignmentConfirmationUUID)
	}
	if details.CrewRequirementID != "" {
		fmt.Fprintf(out, "Crew Requirement ID: %s\n", details.CrewRequirementID)
	}
	if details.ResourceType != "" || details.ResourceID != "" {
		resource := details.ResourceType
		if details.ResourceID != "" {
			if resource != "" {
				resource += "/"
			}
			resource += details.ResourceID
		}
		fmt.Fprintf(out, "Resource: %s\n", resource)
	}
	if details.ConfirmedByID != "" {
		fmt.Fprintf(out, "Confirmed By: %s\n", details.ConfirmedByID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.ConfirmedAt != "" {
		fmt.Fprintf(out, "Confirmed At: %s\n", details.ConfirmedAt)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	fmt.Fprintf(out, "Is Explicit: %t\n", details.IsExplicit)

	return nil
}
