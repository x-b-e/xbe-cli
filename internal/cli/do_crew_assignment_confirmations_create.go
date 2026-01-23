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

type doCrewAssignmentConfirmationsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	AssignmentConfirmationUUID string
	CrewRequirement            string
	ResourceType               string
	ResourceID                 string
	StartAt                    string
	Note                       string
	IsExplicit                 bool
	ConfirmedBy                string
}

func newDoCrewAssignmentConfirmationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a crew assignment confirmation",
		Long: `Create a crew assignment confirmation.

Provide either an assignment confirmation UUID or the crew requirement + resource
+ start-at values. The API will set confirmed-by to the current user by default.

Required (one of the following):
  --assignment-confirmation-uuid

OR
  --crew-requirement
  --resource-type
  --resource-id
  --start-at

Optional flags:
  --note          Confirmation note
  --is-explicit   Mark confirmation as explicit
  --confirmed-by  Confirmed-by user ID (defaults to current user)`,
		Example: `  # Confirm using assignment confirmation UUID
  xbe do crew-assignment-confirmations create \
    --assignment-confirmation-uuid "uuid-here" \
    --note "Confirmed" \
    --is-explicit

  # Confirm using crew requirement + resource + start time
  xbe do crew-assignment-confirmations create \
    --crew-requirement 123 \
    --resource-type laborers \
    --resource-id 456 \
    --start-at "2025-01-01T08:00:00Z" \
    --note "Assigned"`,
		Args: cobra.NoArgs,
		RunE: runDoCrewAssignmentConfirmationsCreate,
	}
	initDoCrewAssignmentConfirmationsCreateFlags(cmd)
	return cmd
}

func init() {
	doCrewAssignmentConfirmationsCmd.AddCommand(newDoCrewAssignmentConfirmationsCreateCmd())
}

func initDoCrewAssignmentConfirmationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("assignment-confirmation-uuid", "", "Assignment confirmation UUID")
	cmd.Flags().String("crew-requirement", "", "Crew requirement ID")
	cmd.Flags().String("resource-type", "", "Resource type (e.g., laborers)")
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.Flags().String("start-at", "", "Assignment start time (ISO 8601)")
	cmd.Flags().String("note", "", "Confirmation note")
	cmd.Flags().Bool("is-explicit", false, "Mark confirmation as explicit")
	cmd.Flags().String("confirmed-by", "", "Confirmed-by user ID (optional)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCrewAssignmentConfirmationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCrewAssignmentConfirmationsCreateOptions(cmd)
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

	if opts.AssignmentConfirmationUUID == "" {
		if opts.CrewRequirement == "" || opts.ResourceType == "" || opts.ResourceID == "" || opts.StartAt == "" {
			err := fmt.Errorf("--assignment-confirmation-uuid or --crew-requirement/--resource-type/--resource-id/--start-at is required")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if opts.ResourceType != "" && opts.ResourceID == "" {
		err := fmt.Errorf("--resource-id is required when --resource-type is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ResourceID != "" && opts.ResourceType == "" {
		err := fmt.Errorf("--resource-type is required when --resource-id is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.AssignmentConfirmationUUID != "" {
		attributes["assignment-confirmation-uuid"] = opts.AssignmentConfirmationUUID
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("is-explicit") {
		attributes["is-explicit"] = opts.IsExplicit
	}

	relationships := map[string]any{}
	if opts.CrewRequirement != "" {
		relationships["crew-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "crew-requirements",
				"id":   opts.CrewRequirement,
			},
		}
	}
	if opts.ResourceType != "" && opts.ResourceID != "" {
		relationships["resource"] = map[string]any{
			"data": map[string]any{
				"type": opts.ResourceType,
				"id":   opts.ResourceID,
			},
		}
	}
	if opts.ConfirmedBy != "" {
		relationships["confirmed-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ConfirmedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "crew-assignment-confirmations",
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/crew-assignment-confirmations", jsonBody)
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

	row := buildCrewAssignmentConfirmationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created crew assignment confirmation %s\n", row.ID)
	return nil
}

func parseDoCrewAssignmentConfirmationsCreateOptions(cmd *cobra.Command) (doCrewAssignmentConfirmationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	assignmentConfirmationUUID, _ := cmd.Flags().GetString("assignment-confirmation-uuid")
	crewRequirement, _ := cmd.Flags().GetString("crew-requirement")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	startAt, _ := cmd.Flags().GetString("start-at")
	note, _ := cmd.Flags().GetString("note")
	isExplicit, _ := cmd.Flags().GetBool("is-explicit")
	confirmedBy, _ := cmd.Flags().GetString("confirmed-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCrewAssignmentConfirmationsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		AssignmentConfirmationUUID: assignmentConfirmationUUID,
		CrewRequirement:            crewRequirement,
		ResourceType:               resourceType,
		ResourceID:                 resourceID,
		StartAt:                    startAt,
		Note:                       note,
		IsExplicit:                 isExplicit,
		ConfirmedBy:                confirmedBy,
	}, nil
}
