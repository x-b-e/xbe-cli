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

type doProjectTransportPlanDriversUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	Status                        string
	ConfirmNote                   string
	ConfirmAtMax                  string
	SkipAssignmentRulesValidation bool
	AssignmentRuleOverrideReason  string
	SegmentStart                  string
	SegmentEnd                    string
	Driver                        string
	InboundProjectOfficeExplicit  string
}

func newDoProjectTransportPlanDriversUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan driver assignment",
		Long: `Update an existing project transport plan driver assignment.

Provide the assignment ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --status                          Assignment status (editing/pending/active)
  --confirm-note                    Confirmation note (empty to clear)
  --confirm-at-max                  Max confirmation time (RFC3339, empty to clear)
  --skip-assignment-rules-validation Skip assignment rules validation
  --assignment-rule-override-reason Reason for assignment rule override (empty to clear)
  --segment-start                   Segment start ID
  --segment-end                     Segment end ID
  --driver                          Driver (user) ID (empty to clear)
  --inbound-project-office-explicit Explicit inbound project office ID (empty to clear)`,
		Example: `  # Update status and confirmation note
  xbe do project-transport-plan-drivers update 123 --status pending --confirm-note "Awaiting confirmation"

  # Clear driver assignment
  xbe do project-transport-plan-drivers update 123 --driver ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanDriversUpdate,
	}
	initDoProjectTransportPlanDriversUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanDriversCmd.AddCommand(newDoProjectTransportPlanDriversUpdateCmd())
}

func initDoProjectTransportPlanDriversUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Assignment status (editing/pending/active)")
	cmd.Flags().String("confirm-note", "", "Confirmation note")
	cmd.Flags().String("confirm-at-max", "", "Max confirmation time (RFC3339)")
	cmd.Flags().Bool("skip-assignment-rules-validation", false, "Skip assignment rules validation")
	cmd.Flags().String("assignment-rule-override-reason", "", "Reason for assignment rule override")
	cmd.Flags().String("segment-start", "", "Segment start ID")
	cmd.Flags().String("segment-end", "", "Segment end ID")
	cmd.Flags().String("driver", "", "Driver (user) ID")
	cmd.Flags().String("inbound-project-office-explicit", "", "Explicit inbound project office ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanDriversUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanDriversUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("confirm-note") {
		attributes["confirm-note"] = opts.ConfirmNote
	}
	if cmd.Flags().Changed("confirm-at-max") {
		attributes["confirm-at-max"] = opts.ConfirmAtMax
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if cmd.Flags().Changed("assignment-rule-override-reason") {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("segment-start") {
		if strings.TrimSpace(opts.SegmentStart) == "" {
			err := fmt.Errorf("--segment-start cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["segment-start"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentStart,
			},
		}
	}
	if cmd.Flags().Changed("segment-end") {
		if strings.TrimSpace(opts.SegmentEnd) == "" {
			err := fmt.Errorf("--segment-end cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["segment-end"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentEnd,
			},
		}
	}
	if cmd.Flags().Changed("driver") {
		if strings.TrimSpace(opts.Driver) == "" {
			relationships["driver"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["driver"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.Driver,
				},
			}
		}
	}
	if cmd.Flags().Changed("inbound-project-office-explicit") {
		if strings.TrimSpace(opts.InboundProjectOfficeExplicit) == "" {
			relationships["inbound-project-office-explicit"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["inbound-project-office-explicit"] = map[string]any{
				"data": map[string]any{
					"type": "project-offices",
					"id":   opts.InboundProjectOfficeExplicit,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-drivers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-drivers/"+opts.ID, jsonBody)
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

	row := projectTransportPlanDriverRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan driver %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanDriversUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanDriversUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	confirmNote, _ := cmd.Flags().GetString("confirm-note")
	confirmAtMax, _ := cmd.Flags().GetString("confirm-at-max")
	skipAssignmentRulesValidation, _ := cmd.Flags().GetBool("skip-assignment-rules-validation")
	assignmentRuleOverrideReason, _ := cmd.Flags().GetString("assignment-rule-override-reason")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	driver, _ := cmd.Flags().GetString("driver")
	inboundProjectOfficeExplicit, _ := cmd.Flags().GetString("inbound-project-office-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanDriversUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            strings.TrimSpace(args[0]),
		Status:                        status,
		ConfirmNote:                   confirmNote,
		ConfirmAtMax:                  confirmAtMax,
		SkipAssignmentRulesValidation: skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:  assignmentRuleOverrideReason,
		SegmentStart:                  segmentStart,
		SegmentEnd:                    segmentEnd,
		Driver:                        driver,
		InboundProjectOfficeExplicit:  inboundProjectOfficeExplicit,
	}, nil
}
