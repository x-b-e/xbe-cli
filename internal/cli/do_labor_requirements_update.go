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

type doLaborRequirementsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	JobProductionPlan         string
	LaborClassification       string
	Laborer                   string
	CraftClass                string
	ProjectCostClassification string
	StartAt                   string
	EndAt                     string
	MobilizationMethod        string
	Note                      string
	RequiresInboundMovement   bool
	RequiresOutboundMovement  bool
	IsValidatingOverlapping   bool
	ExplicitInboundLatitude   string
	ExplicitInboundLongitude  string
	ExplicitOutboundLatitude  string
	ExplicitOutboundLongitude string
}

func newDoLaborRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a labor requirement",
		Long: `Update a labor requirement.

Optional:
  --job-production-plan         Job production plan ID
  --labor-classification        Labor classification ID
  --laborer                     Laborer ID to assign
  --craft-class                 Craft class ID
  --project-cost-classification Project cost classification ID
  --start-at                    Start time (ISO 8601)
  --end-at                      End time (ISO 8601)
  --mobilization-method         Mobilization method (crew/heavy_equipment_transport/lowboy/itself/trailer)
  --note                        Note
  --is-validating-overlapping   Validate overlaps (true/false)
  --explicit-inbound-latitude   Explicit inbound latitude
  --explicit-inbound-longitude  Explicit inbound longitude
  --explicit-outbound-latitude  Explicit outbound latitude
  --explicit-outbound-longitude Explicit outbound longitude
  --requires-inbound-movement   Set inbound movement requirement (true/false)
  --requires-outbound-movement  Set outbound movement requirement (true/false)`,
		Example: `  # Update schedule
  xbe do labor-requirements update 123 --start-at \"2026-01-23T09:00:00Z\" --end-at \"2026-01-23T13:00:00Z\"

  # Assign a laborer
  xbe do labor-requirements update 123 --laborer 456

  # Update note
  xbe do labor-requirements update 123 --note \"Updated note\"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLaborRequirementsUpdate,
	}
	initDoLaborRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doLaborRequirementsCmd.AddCommand(newDoLaborRequirementsUpdateCmd())
}

func initDoLaborRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("labor-classification", "", "Labor classification ID")
	cmd.Flags().String("laborer", "", "Laborer ID")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601)")
	cmd.Flags().String("mobilization-method", "", "Mobilization method (crew/heavy_equipment_transport/lowboy/itself/trailer)")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().Bool("requires-inbound-movement", false, "Inbound movement requirement")
	cmd.Flags().Bool("requires-outbound-movement", false, "Outbound movement requirement")
	cmd.Flags().Bool("is-validating-overlapping", false, "Validate overlaps")
	cmd.Flags().String("explicit-inbound-latitude", "", "Explicit inbound latitude")
	cmd.Flags().String("explicit-inbound-longitude", "", "Explicit inbound longitude")
	cmd.Flags().String("explicit-outbound-latitude", "", "Explicit outbound latitude")
	cmd.Flags().String("explicit-outbound-longitude", "", "Explicit outbound longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborRequirementsUpdateOptions(cmd, args)
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

	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("requires-inbound-movement") {
		attributes["requires-inbound-movement"] = opts.RequiresInboundMovement
	}
	if cmd.Flags().Changed("requires-outbound-movement") {
		attributes["requires-outbound-movement"] = opts.RequiresOutboundMovement
	}
	if cmd.Flags().Changed("is-validating-overlapping") {
		attributes["is-validating-overlapping"] = opts.IsValidatingOverlapping
	}
	if opts.ExplicitInboundLatitude != "" {
		attributes["explicit-inbound-latitude"] = opts.ExplicitInboundLatitude
	}
	if opts.ExplicitInboundLongitude != "" {
		attributes["explicit-inbound-longitude"] = opts.ExplicitInboundLongitude
	}
	if opts.ExplicitOutboundLatitude != "" {
		attributes["explicit-outbound-latitude"] = opts.ExplicitOutboundLatitude
	}
	if opts.ExplicitOutboundLongitude != "" {
		attributes["explicit-outbound-longitude"] = opts.ExplicitOutboundLongitude
	}

	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if opts.LaborClassification != "" {
		relationships["resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": "labor-classifications",
				"id":   opts.LaborClassification,
			},
		}
	}
	if opts.Laborer != "" {
		relationships["resource"] = map[string]any{
			"data": map[string]any{
				"type": "laborers",
				"id":   opts.Laborer,
			},
		}
	}
	if opts.CraftClass != "" {
		relationships["craft-class"] = map[string]any{
			"data": map[string]any{
				"type": "craft-classes",
				"id":   opts.CraftClass,
			},
		}
	}
	if opts.ProjectCostClassification != "" {
		relationships["project-cost-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   opts.ProjectCostClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "labor-requirements",
			"id":         opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/labor-requirements/"+opts.ID, jsonBody)
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

	row := buildLaborRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated labor requirement %s\n", row.ID)
	return nil
}

func parseDoLaborRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doLaborRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	laborClassification, _ := cmd.Flags().GetString("labor-classification")
	laborer, _ := cmd.Flags().GetString("laborer")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	note, _ := cmd.Flags().GetString("note")
	requiresInboundMovement, _ := cmd.Flags().GetBool("requires-inbound-movement")
	requiresOutboundMovement, _ := cmd.Flags().GetBool("requires-outbound-movement")
	isValidatingOverlapping, _ := cmd.Flags().GetBool("is-validating-overlapping")
	explicitInboundLatitude, _ := cmd.Flags().GetString("explicit-inbound-latitude")
	explicitInboundLongitude, _ := cmd.Flags().GetString("explicit-inbound-longitude")
	explicitOutboundLatitude, _ := cmd.Flags().GetString("explicit-outbound-latitude")
	explicitOutboundLongitude, _ := cmd.Flags().GetString("explicit-outbound-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborRequirementsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
		JobProductionPlan:         jobProductionPlan,
		LaborClassification:       laborClassification,
		Laborer:                   laborer,
		CraftClass:                craftClass,
		ProjectCostClassification: projectCostClassification,
		StartAt:                   startAt,
		EndAt:                     endAt,
		MobilizationMethod:        mobilizationMethod,
		Note:                      note,
		RequiresInboundMovement:   requiresInboundMovement,
		RequiresOutboundMovement:  requiresOutboundMovement,
		IsValidatingOverlapping:   isValidatingOverlapping,
		ExplicitInboundLatitude:   explicitInboundLatitude,
		ExplicitInboundLongitude:  explicitInboundLongitude,
		ExplicitOutboundLatitude:  explicitOutboundLatitude,
		ExplicitOutboundLongitude: explicitOutboundLongitude,
	}, nil
}
