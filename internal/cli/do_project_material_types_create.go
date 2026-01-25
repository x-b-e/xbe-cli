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

type doProjectMaterialTypesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Project             string
	MaterialType        string
	Quantity            string
	ExplicitDisplayName string
	PickupAtMin         string
	PickupAtMax         string
	DeliverAtMin        string
	DeliverAtMax        string
	UnitOfMeasure       string
	MaterialSite        string
	JobSite             string
	PickupLocation      string
	DeliveryLocation    string
}

func newDoProjectMaterialTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project material type",
		Long: `Create a project material type.

Required flags:
  --project            Project ID (required)
  --material-type      Material type ID (required)

Optional flags:
  --quantity           Quantity
  --explicit-display-name  Explicit display name override
  --pickup-at-min      Earliest pickup time (ISO 8601)
  --pickup-at-max      Latest pickup time (ISO 8601)
  --deliver-at-min     Earliest delivery time (ISO 8601)
  --deliver-at-max     Latest delivery time (ISO 8601)
  --unit-of-measure    Unit of measure ID
  --material-site      Material site ID
  --job-site           Job site ID
  --pickup-location    Pickup location ID (project transport location)
  --delivery-location  Delivery location ID (project transport location)`,
		Example: `  # Create a project material type
  xbe do project-material-types create \
    --project 123 \
    --material-type 456 \
    --quantity 12.5 \
    --unit-of-measure 789

  # Create with pickup/delivery windows
  xbe do project-material-types create \
    --project 123 \
    --material-type 456 \
    --pickup-at-min 2026-01-01T08:00:00Z \
    --deliver-at-max 2026-01-01T18:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoProjectMaterialTypesCreate,
	}
	initDoProjectMaterialTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectMaterialTypesCmd.AddCommand(newDoProjectMaterialTypesCreateCmd())
}

func initDoProjectMaterialTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("quantity", "", "Quantity")
	cmd.Flags().String("explicit-display-name", "", "Explicit display name override")
	cmd.Flags().String("pickup-at-min", "", "Earliest pickup time (ISO 8601)")
	cmd.Flags().String("pickup-at-max", "", "Latest pickup time (ISO 8601)")
	cmd.Flags().String("deliver-at-min", "", "Earliest delivery time (ISO 8601)")
	cmd.Flags().String("deliver-at-max", "", "Latest delivery time (ISO 8601)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("job-site", "", "Job site ID")
	cmd.Flags().String("pickup-location", "", "Pickup location ID (project transport location)")
	cmd.Flags().String("delivery-location", "", "Delivery location ID (project transport location)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectMaterialTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectMaterialTypesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.MaterialType) == "" {
		err := fmt.Errorf("--material-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Quantity != "" {
		attributes["quantity"] = opts.Quantity
	}
	if opts.ExplicitDisplayName != "" {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if opts.PickupAtMin != "" {
		attributes["pickup-at-min"] = opts.PickupAtMin
	}
	if opts.PickupAtMax != "" {
		attributes["pickup-at-max"] = opts.PickupAtMax
	}
	if opts.DeliverAtMin != "" {
		attributes["deliver-at-min"] = opts.DeliverAtMin
	}
	if opts.DeliverAtMax != "" {
		attributes["deliver-at-max"] = opts.DeliverAtMax
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
	}

	if strings.TrimSpace(opts.UnitOfMeasure) != "" {
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		}
	}
	if strings.TrimSpace(opts.MaterialSite) != "" {
		relationships["material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		}
	}
	if strings.TrimSpace(opts.JobSite) != "" {
		relationships["job-site"] = map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   opts.JobSite,
			},
		}
	}
	if strings.TrimSpace(opts.PickupLocation) != "" {
		relationships["pickup-location"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.PickupLocation,
			},
		}
	}
	if strings.TrimSpace(opts.DeliveryLocation) != "" {
		relationships["delivery-location"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.DeliveryLocation,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-material-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-material-types", jsonBody)
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

	row := buildProjectMaterialTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project material type %s\n", row.ID)
	return nil
}

func parseDoProjectMaterialTypesCreateOptions(cmd *cobra.Command) (doProjectMaterialTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	materialType, _ := cmd.Flags().GetString("material-type")
	quantity, _ := cmd.Flags().GetString("quantity")
	explicitDisplayName, _ := cmd.Flags().GetString("explicit-display-name")
	pickupAtMin, _ := cmd.Flags().GetString("pickup-at-min")
	pickupAtMax, _ := cmd.Flags().GetString("pickup-at-max")
	deliverAtMin, _ := cmd.Flags().GetString("deliver-at-min")
	deliverAtMax, _ := cmd.Flags().GetString("deliver-at-max")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	materialSite, _ := cmd.Flags().GetString("material-site")
	jobSite, _ := cmd.Flags().GetString("job-site")
	pickupLocation, _ := cmd.Flags().GetString("pickup-location")
	deliveryLocation, _ := cmd.Flags().GetString("delivery-location")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectMaterialTypesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Project:             project,
		MaterialType:        materialType,
		Quantity:            quantity,
		ExplicitDisplayName: explicitDisplayName,
		PickupAtMin:         pickupAtMin,
		PickupAtMax:         pickupAtMax,
		DeliverAtMin:        deliverAtMin,
		DeliverAtMax:        deliverAtMax,
		UnitOfMeasure:       unitOfMeasure,
		MaterialSite:        materialSite,
		JobSite:             jobSite,
		PickupLocation:      pickupLocation,
		DeliveryLocation:    deliveryLocation,
	}, nil
}

func buildProjectMaterialTypeRowFromSingle(resp jsonAPISingleResponse) projectMaterialTypeRow {
	attrs := resp.Data.Attributes

	row := projectMaterialTypeRow{
		ID:                  resp.Data.ID,
		Quantity:            stringAttr(attrs, "quantity"),
		DisplayName:         stringAttr(attrs, "display-name"),
		ExplicitDisplayName: stringAttr(attrs, "explicit-display-name"),
		PickupAtMin:         formatDateTime(stringAttr(attrs, "pickup-at-min")),
		PickupAtMax:         formatDateTime(stringAttr(attrs, "pickup-at-max")),
		DeliverAtMin:        formatDateTime(stringAttr(attrs, "deliver-at-min")),
		DeliverAtMax:        formatDateTime(stringAttr(attrs, "deliver-at-max")),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-site"]; ok && rel.Data != nil {
		row.JobSiteID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["pickup-location"]; ok && rel.Data != nil {
		row.PickupLocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["delivery-location"]; ok && rel.Data != nil {
		row.DeliveryLocationID = rel.Data.ID
	}

	return row
}
