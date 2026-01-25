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

type doProjectTransportPlanDriversCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ProjectTransportPlan          string
	SegmentStart                  string
	SegmentEnd                    string
	Driver                        string
	Status                        string
	ConfirmNote                   string
	ConfirmAtMax                  string
	SkipAssignmentRulesValidation bool
	AssignmentRuleOverrideReason  string
	InboundProjectOfficeExplicit  string
}

func newDoProjectTransportPlanDriversCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan driver assignment",
		Long: `Create a project transport plan driver assignment.

Required:
  --project-transport-plan  Project transport plan ID
  --segment-start           Segment start ID
  --segment-end             Segment end ID

Optional:
  --driver                          Driver (user) ID
  --status                          Assignment status (editing/pending/active)
  --confirm-note                    Confirmation note
  --confirm-at-max                  Max confirmation time (RFC3339)
  --skip-assignment-rules-validation Skip assignment rules validation
  --assignment-rule-override-reason Reason for assignment rule override
  --inbound-project-office-explicit Explicit inbound project office ID`,
		Example: `  # Create a driver assignment
  xbe do project-transport-plan-drivers create \\
    --project-transport-plan 123 \\
    --segment-start 456 \\
    --segment-end 789

  # Create with driver and status
  xbe do project-transport-plan-drivers create \\
    --project-transport-plan 123 \\
    --segment-start 456 \\
    --segment-end 789 \\
    --driver 321 \\
    --status pending`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanDriversCreate,
	}
	initDoProjectTransportPlanDriversCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanDriversCmd.AddCommand(newDoProjectTransportPlanDriversCreateCmd())
}

func initDoProjectTransportPlanDriversCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID (required)")
	cmd.Flags().String("segment-start", "", "Segment start ID (required)")
	cmd.Flags().String("segment-end", "", "Segment end ID (required)")
	cmd.Flags().String("driver", "", "Driver (user) ID")
	cmd.Flags().String("status", "", "Assignment status (editing/pending/active)")
	cmd.Flags().String("confirm-note", "", "Confirmation note")
	cmd.Flags().String("confirm-at-max", "", "Max confirmation time (RFC3339)")
	cmd.Flags().Bool("skip-assignment-rules-validation", false, "Skip assignment rules validation")
	cmd.Flags().String("assignment-rule-override-reason", "", "Reason for assignment rule override")
	cmd.Flags().String("inbound-project-office-explicit", "", "Explicit inbound project office ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanDriversCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanDriversCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectTransportPlan) == "" {
		err := fmt.Errorf("--project-transport-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentStart) == "" {
		err := fmt.Errorf("--segment-start is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentEnd) == "" {
		err := fmt.Errorf("--segment-end is required")
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
	if strings.TrimSpace(opts.Status) != "" {
		attributes["status"] = opts.Status
	}
	if strings.TrimSpace(opts.ConfirmNote) != "" {
		attributes["confirm-note"] = opts.ConfirmNote
	}
	if strings.TrimSpace(opts.ConfirmAtMax) != "" {
		attributes["confirm-at-max"] = opts.ConfirmAtMax
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if strings.TrimSpace(opts.AssignmentRuleOverrideReason) != "" {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		},
		"segment-start": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentStart,
			},
		},
		"segment-end": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentEnd,
			},
		},
	}

	if strings.TrimSpace(opts.Driver) != "" {
		relationships["driver"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Driver,
			},
		}
	}

	if strings.TrimSpace(opts.InboundProjectOfficeExplicit) != "" {
		relationships["inbound-project-office-explicit"] = map[string]any{
			"data": map[string]any{
				"type": "project-offices",
				"id":   opts.InboundProjectOfficeExplicit,
			},
		}
	}

	data := map[string]any{
		"type":          "project-transport-plan-drivers",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-drivers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan driver %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanDriversCreateOptions(cmd *cobra.Command) (doProjectTransportPlanDriversCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	driver, _ := cmd.Flags().GetString("driver")
	status, _ := cmd.Flags().GetString("status")
	confirmNote, _ := cmd.Flags().GetString("confirm-note")
	confirmAtMax, _ := cmd.Flags().GetString("confirm-at-max")
	skipAssignmentRulesValidation, _ := cmd.Flags().GetBool("skip-assignment-rules-validation")
	assignmentRuleOverrideReason, _ := cmd.Flags().GetString("assignment-rule-override-reason")
	inboundProjectOfficeExplicit, _ := cmd.Flags().GetString("inbound-project-office-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanDriversCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ProjectTransportPlan:          projectTransportPlan,
		SegmentStart:                  segmentStart,
		SegmentEnd:                    segmentEnd,
		Driver:                        driver,
		Status:                        status,
		ConfirmNote:                   confirmNote,
		ConfirmAtMax:                  confirmAtMax,
		SkipAssignmentRulesValidation: skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:  assignmentRuleOverrideReason,
		InboundProjectOfficeExplicit:  inboundProjectOfficeExplicit,
	}, nil
}
