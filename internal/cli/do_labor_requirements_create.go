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

type doLaborRequirementsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
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

func newDoLaborRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a labor requirement",
		Long: `Create a labor requirement.

Required:
  --job-production-plan   Job production plan ID
  --labor-classification  Labor classification ID

Optional:
  --laborer                      Laborer ID to assign
  --craft-class                  Craft class ID
  --project-cost-classification  Project cost classification ID
  --start-at                     Start time (ISO 8601)
  --end-at                       End time (ISO 8601)
  --mobilization-method          Mobilization method (crew/heavy_equipment_transport/lowboy/itself/trailer)
  --note                         Note
  --is-validating-overlapping    Validate overlaps (true/false)
  --explicit-inbound-latitude    Explicit inbound latitude
  --explicit-inbound-longitude   Explicit inbound longitude
  --explicit-outbound-latitude   Explicit outbound latitude
  --explicit-outbound-longitude  Explicit outbound longitude
  --requires-inbound-movement    Set inbound movement requirement (true/false)
  --requires-outbound-movement   Set outbound movement requirement (true/false)`,
		Example: `  # Create a labor requirement
  xbe do labor-requirements create --job-production-plan 123 --labor-classification 456

  # Create with schedule and assignment
  xbe do labor-requirements create --job-production-plan 123 --labor-classification 456 \
    --start-at \"2026-01-23T08:00:00Z\" --end-at \"2026-01-23T12:00:00Z\" --laborer 789

  # Create with coordinates
  xbe do labor-requirements create --job-production-plan 123 --labor-classification 456 \
    --explicit-inbound-latitude 41.881 --explicit-inbound-longitude -87.623`,
		Args: cobra.NoArgs,
		RunE: runDoLaborRequirementsCreate,
	}
	initDoLaborRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doLaborRequirementsCmd.AddCommand(newDoLaborRequirementsCreateCmd())
}

func initDoLaborRequirementsCreateFlags(cmd *cobra.Command) {
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

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("labor-classification")
}

func runDoLaborRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLaborRequirementsCreateOptions(cmd)
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

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"resource-classification": map[string]any{
			"data": map[string]any{
				"type": "labor-classifications",
				"id":   opts.LaborClassification,
			},
		},
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
			"type":          "labor-requirements",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/labor-requirements", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created labor requirement %s\n", row.ID)
	return nil
}

func parseDoLaborRequirementsCreateOptions(cmd *cobra.Command) (doLaborRequirementsCreateOptions, error) {
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

	return doLaborRequirementsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
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

func buildLaborRequirementRowFromSingle(resp jsonAPISingleResponse) laborRequirementRow {
	return buildLaborRequirementRow(resp.Data)
}
