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

type doEquipmentRequirementsCreateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	JobProductionPlan                        string
	ResourceClassificationType               string
	ResourceClassificationID                 string
	ResourceType                             string
	ResourceID                               string
	OriginMaterialSite                       string
	LaborRequirement                         string
	CraftClass                               string
	ProjectCostClassification                string
	CrewRequirementCredentialClassifications string
	StartAt                                  string
	EndAt                                    string
	MobilizationMethod                       string
	Note                                     string
	RequiresInboundMovement                  string
	RequiresOutboundMovement                 string
	IsValidatingOverlapping                  string
	ExplicitInboundLatitude                  string
	ExplicitInboundLongitude                 string
	ExplicitOutboundLatitude                 string
	ExplicitOutboundLongitude                string
}

func newDoEquipmentRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment requirement",
		Long: `Create an equipment requirement.

Required flags:
  --job-production-plan           Job production plan ID (required)
  --resource-classification-type  Resource classification type (JSON:API type, e.g., equipment-classifications) (required)
  --resource-classification-id    Resource classification ID (required)

Optional flags:
  --resource-type                 Resource type (JSON:API type, e.g., equipment)
  --resource-id                   Resource ID (requires --resource-type)
  --start-at                      Start time (ISO 8601; requires --end-at)
  --end-at                        End time (ISO 8601; requires --start-at)
  --mobilization-method           Mobilization method (crew, heavy_equipment_transport, lowboy, itself, trailer)
  --note                          Requirement note
  --requires-inbound-movement     Requires inbound movement (true/false)
  --requires-outbound-movement    Requires outbound movement (true/false)
  --is-validating-overlapping     Validate overlapping assignments (true/false)
  --origin-material-site          Origin material site ID
  --labor-requirement             Labor requirement ID
  --craft-class                   Craft class ID
  --project-cost-classification   Project cost classification ID
  --crew-requirement-credential-classifications Comma-separated credential classification link IDs
  --explicit-inbound-latitude     Explicit inbound latitude
  --explicit-inbound-longitude    Explicit inbound longitude
  --explicit-outbound-latitude    Explicit outbound latitude
  --explicit-outbound-longitude   Explicit outbound longitude`,
		Example: `  # Create an equipment requirement
  xbe do equipment-requirements create \\
    --job-production-plan 123 \\
    --resource-classification-type equipment-classifications \\
    --resource-classification-id 456 \\
    --resource-type equipment \\
    --resource-id 789 \\
    --start-at 2025-01-01T08:00:00Z \\
    --end-at 2025-01-01T16:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentRequirementsCreate,
	}
	initDoEquipmentRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentRequirementsCmd.AddCommand(newDoEquipmentRequirementsCreateCmd())
}

func initDoEquipmentRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("resource-classification-type", "", "Resource classification type (JSON:API type, e.g., equipment-classifications) (required)")
	cmd.Flags().String("resource-classification-id", "", "Resource classification ID (required)")
	cmd.Flags().String("resource-type", "", "Resource type (JSON:API type, e.g., equipment)")
	cmd.Flags().String("resource-id", "", "Resource ID (requires --resource-type)")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601; requires --end-at)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601; requires --start-at)")
	cmd.Flags().String("mobilization-method", "", "Mobilization method (crew, heavy_equipment_transport, lowboy, itself, trailer)")
	cmd.Flags().String("note", "", "Requirement note")
	cmd.Flags().String("requires-inbound-movement", "", "Requires inbound movement (true/false)")
	cmd.Flags().String("requires-outbound-movement", "", "Requires outbound movement (true/false)")
	cmd.Flags().String("is-validating-overlapping", "", "Validate overlapping assignments (true/false)")
	cmd.Flags().String("origin-material-site", "", "Origin material site ID")
	cmd.Flags().String("labor-requirement", "", "Labor requirement ID")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID")
	cmd.Flags().String("crew-requirement-credential-classifications", "", "Comma-separated credential classification link IDs")
	cmd.Flags().String("explicit-inbound-latitude", "", "Explicit inbound latitude")
	cmd.Flags().String("explicit-inbound-longitude", "", "Explicit inbound longitude")
	cmd.Flags().String("explicit-outbound-latitude", "", "Explicit outbound latitude")
	cmd.Flags().String("explicit-outbound-longitude", "", "Explicit outbound longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentRequirementsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if opts.JobProductionPlan == "" {
		return fmt.Errorf("--job-production-plan is required")
	}
	if opts.ResourceClassificationType == "" {
		return fmt.Errorf("--resource-classification-type is required")
	}
	if opts.ResourceClassificationID == "" {
		return fmt.Errorf("--resource-classification-id is required")
	}
	if opts.ResourceType != "" && opts.ResourceID == "" {
		return fmt.Errorf("--resource-id is required when --resource-type is set")
	}
	if opts.ResourceID != "" && opts.ResourceType == "" {
		return fmt.Errorf("--resource-type is required when --resource-id is set")
	}

	if (opts.StartAt != "" && opts.EndAt == "") || (opts.StartAt == "" && opts.EndAt != "") {
		return fmt.Errorf("--start-at and --end-at must be set together")
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

	if cmd.Flags().Changed("requires-inbound-movement") {
		value, err := parseCrewRequirementBool(opts.RequiresInboundMovement, "requires-inbound-movement")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requires-inbound-movement"] = value
	}
	if cmd.Flags().Changed("requires-outbound-movement") {
		value, err := parseCrewRequirementBool(opts.RequiresOutboundMovement, "requires-outbound-movement")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requires-outbound-movement"] = value
	}
	if cmd.Flags().Changed("is-validating-overlapping") {
		value, err := parseCrewRequirementBool(opts.IsValidatingOverlapping, "is-validating-overlapping")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["is-validating-overlapping"] = value
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
				"type": opts.ResourceClassificationType,
				"id":   opts.ResourceClassificationID,
			},
		},
	}

	if opts.ResourceType != "" && opts.ResourceID != "" {
		relationships["resource"] = map[string]any{
			"data": map[string]any{
				"type": opts.ResourceType,
				"id":   opts.ResourceID,
			},
		}
	}
	if opts.OriginMaterialSite != "" {
		relationships["origin-material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.OriginMaterialSite,
			},
		}
	}
	if opts.LaborRequirement != "" {
		relationships["labor-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "labor-requirements",
				"id":   opts.LaborRequirement,
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
	if opts.CrewRequirementCredentialClassifications != "" {
		ids := strings.Split(opts.CrewRequirementCredentialClassifications, ",")
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			data = append(data, map[string]any{
				"type": "crew-requirement-credential-classifications",
				"id":   id,
			})
		}
		if len(data) > 0 {
			relationships["crew-requirement-credential-classifications"] = map[string]any{"data": data}
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-requirements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-requirements", jsonBody)
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

	row := buildEquipmentRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment requirement %s\n", row.ID)
	return nil
}

func parseDoEquipmentRequirementsCreateOptions(cmd *cobra.Command) (doEquipmentRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	originMaterialSite, _ := cmd.Flags().GetString("origin-material-site")
	laborRequirement, _ := cmd.Flags().GetString("labor-requirement")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	crewRequirementCredentialClassifications, _ := cmd.Flags().GetString("crew-requirement-credential-classifications")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	note, _ := cmd.Flags().GetString("note")
	requiresInboundMovement, _ := cmd.Flags().GetString("requires-inbound-movement")
	requiresOutboundMovement, _ := cmd.Flags().GetString("requires-outbound-movement")
	isValidatingOverlapping, _ := cmd.Flags().GetString("is-validating-overlapping")
	explicitInboundLatitude, _ := cmd.Flags().GetString("explicit-inbound-latitude")
	explicitInboundLongitude, _ := cmd.Flags().GetString("explicit-inbound-longitude")
	explicitOutboundLatitude, _ := cmd.Flags().GetString("explicit-outbound-latitude")
	explicitOutboundLongitude, _ := cmd.Flags().GetString("explicit-outbound-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentRequirementsCreateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		JobProductionPlan:                        jobProductionPlan,
		ResourceClassificationType:               resourceClassificationType,
		ResourceClassificationID:                 resourceClassificationID,
		ResourceType:                             resourceType,
		ResourceID:                               resourceID,
		OriginMaterialSite:                       originMaterialSite,
		LaborRequirement:                         laborRequirement,
		CraftClass:                               craftClass,
		ProjectCostClassification:                projectCostClassification,
		CrewRequirementCredentialClassifications: crewRequirementCredentialClassifications,
		StartAt:                                  startAt,
		EndAt:                                    endAt,
		MobilizationMethod:                       mobilizationMethod,
		Note:                                     note,
		RequiresInboundMovement:                  requiresInboundMovement,
		RequiresOutboundMovement:                 requiresOutboundMovement,
		IsValidatingOverlapping:                  isValidatingOverlapping,
		ExplicitInboundLatitude:                  explicitInboundLatitude,
		ExplicitInboundLongitude:                 explicitInboundLongitude,
		ExplicitOutboundLatitude:                 explicitOutboundLatitude,
		ExplicitOutboundLongitude:                explicitOutboundLongitude,
	}, nil
}
