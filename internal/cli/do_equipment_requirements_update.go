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

type doEquipmentRequirementsUpdateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	ID                                       string
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

func newDoEquipmentRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment requirement",
		Long: `Update an equipment requirement.

Provide at least one attribute or relationship to update.`,
		Example: `  # Update note and mobilization method
  xbe do equipment-requirements update 123 --note "Updated note" --mobilization-method trailer

  # Update schedule
  xbe do equipment-requirements update 123 --start-at 2025-01-02T08:00:00Z --end-at 2025-01-02T16:00:00Z

  # Update resource assignment
  xbe do equipment-requirements update 123 --resource-type equipment --resource-id 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentRequirementsUpdate,
	}
	initDoEquipmentRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentRequirementsCmd.AddCommand(newDoEquipmentRequirementsUpdateCmd())
}

func initDoEquipmentRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("resource-classification-type", "", "Resource classification type (JSON:API type, e.g., equipment-classifications)")
	cmd.Flags().String("resource-classification-id", "", "Resource classification ID (requires --resource-classification-type)")
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
	cmd.Flags().String("crew-requirement-credential-classifications", "", "Comma-separated credential classification link IDs (empty clears)")
	cmd.Flags().String("explicit-inbound-latitude", "", "Explicit inbound latitude")
	cmd.Flags().String("explicit-inbound-longitude", "", "Explicit inbound longitude")
	cmd.Flags().String("explicit-outbound-latitude", "", "Explicit outbound latitude")
	cmd.Flags().String("explicit-outbound-longitude", "", "Explicit outbound longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentRequirementsUpdateOptions(cmd, args)
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

	if opts.ResourceClassificationID != "" && opts.ResourceClassificationType == "" {
		return fmt.Errorf("--resource-classification-type is required when --resource-classification-id is set")
	}
	if opts.ResourceID != "" && opts.ResourceType == "" {
		return fmt.Errorf("--resource-type is required when --resource-id is set")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	startAtChanged := cmd.Flags().Changed("start-at")
	endAtChanged := cmd.Flags().Changed("end-at")
	if startAtChanged || endAtChanged {
		if startAtChanged != endAtChanged {
			return fmt.Errorf("--start-at and --end-at must be set together")
		}
		if opts.StartAt == "" && opts.EndAt == "" {
			attributes["start-at"] = nil
			attributes["end-at"] = nil
		} else if opts.StartAt == "" || opts.EndAt == "" {
			return fmt.Errorf("--start-at and --end-at must be set together")
		} else {
			attributes["start-at"] = opts.StartAt
			attributes["end-at"] = opts.EndAt
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("mobilization-method") {
		if opts.MobilizationMethod == "" {
			attributes["mobilization-method"] = nil
		} else {
			attributes["mobilization-method"] = opts.MobilizationMethod
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-inbound-latitude") {
		if opts.ExplicitInboundLatitude == "" {
			attributes["explicit-inbound-latitude"] = nil
		} else {
			attributes["explicit-inbound-latitude"] = opts.ExplicitInboundLatitude
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-inbound-longitude") {
		if opts.ExplicitInboundLongitude == "" {
			attributes["explicit-inbound-longitude"] = nil
		} else {
			attributes["explicit-inbound-longitude"] = opts.ExplicitInboundLongitude
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-outbound-latitude") {
		if opts.ExplicitOutboundLatitude == "" {
			attributes["explicit-outbound-latitude"] = nil
		} else {
			attributes["explicit-outbound-latitude"] = opts.ExplicitOutboundLatitude
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-outbound-longitude") {
		if opts.ExplicitOutboundLongitude == "" {
			attributes["explicit-outbound-longitude"] = nil
		} else {
			attributes["explicit-outbound-longitude"] = opts.ExplicitOutboundLongitude
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("requires-inbound-movement") {
		value, err := parseCrewRequirementBool(opts.RequiresInboundMovement, "requires-inbound-movement")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requires-inbound-movement"] = value
		hasChanges = true
	}
	if cmd.Flags().Changed("requires-outbound-movement") {
		value, err := parseCrewRequirementBool(opts.RequiresOutboundMovement, "requires-outbound-movement")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requires-outbound-movement"] = value
		hasChanges = true
	}
	if cmd.Flags().Changed("is-validating-overlapping") {
		value, err := parseCrewRequirementBool(opts.IsValidatingOverlapping, "is-validating-overlapping")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["is-validating-overlapping"] = value
		hasChanges = true
	}

	if cmd.Flags().Changed("job-production-plan") {
		if opts.JobProductionPlan == "" {
			relationships["job-production-plan"] = map[string]any{"data": nil}
		} else {
			relationships["job-production-plan"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plans",
					"id":   opts.JobProductionPlan,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("resource-classification-type") || cmd.Flags().Changed("resource-classification-id") {
		if opts.ResourceClassificationType == "" || opts.ResourceClassificationID == "" {
			relationships["resource-classification"] = map[string]any{"data": nil}
		} else {
			relationships["resource-classification"] = map[string]any{
				"data": map[string]any{
					"type": opts.ResourceClassificationType,
					"id":   opts.ResourceClassificationID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("resource-type") || cmd.Flags().Changed("resource-id") {
		if opts.ResourceType == "" || opts.ResourceID == "" {
			relationships["resource"] = map[string]any{"data": nil}
		} else {
			relationships["resource"] = map[string]any{
				"data": map[string]any{
					"type": opts.ResourceType,
					"id":   opts.ResourceID,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("origin-material-site") {
		if opts.OriginMaterialSite == "" {
			relationships["origin-material-site"] = map[string]any{"data": nil}
		} else {
			relationships["origin-material-site"] = map[string]any{
				"data": map[string]any{
					"type": "material-sites",
					"id":   opts.OriginMaterialSite,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("labor-requirement") {
		if opts.LaborRequirement == "" {
			relationships["labor-requirement"] = map[string]any{"data": nil}
		} else {
			relationships["labor-requirement"] = map[string]any{
				"data": map[string]any{
					"type": "labor-requirements",
					"id":   opts.LaborRequirement,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("craft-class") {
		if opts.CraftClass == "" {
			relationships["craft-class"] = map[string]any{"data": nil}
		} else {
			relationships["craft-class"] = map[string]any{
				"data": map[string]any{
					"type": "craft-classes",
					"id":   opts.CraftClass,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("project-cost-classification") {
		if opts.ProjectCostClassification == "" {
			relationships["project-cost-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-cost-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-cost-classifications",
					"id":   opts.ProjectCostClassification,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("crew-requirement-credential-classifications") {
		if opts.CrewRequirementCredentialClassifications == "" {
			relationships["crew-requirement-credential-classifications"] = map[string]any{"data": []any{}}
		} else {
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
			relationships["crew-requirement-credential-classifications"] = map[string]any{"data": data}
		}
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("at least one attribute or relationship must be specified")
	}

	data := map[string]any{
		"type":       "equipment-requirements",
		"id":         opts.ID,
		"attributes": attributes,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-requirements/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment requirement %s\n", row.ID)
	return nil
}

func parseDoEquipmentRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentRequirementsUpdateOptions, error) {
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

	return doEquipmentRequirementsUpdateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		ID:                                       args[0],
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
