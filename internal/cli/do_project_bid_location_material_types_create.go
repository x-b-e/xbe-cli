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

type doProjectBidLocationMaterialTypesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ProjectBidLocation string
	MaterialType       string
	UnitOfMeasure      string
	Quantity           string
	Notes              string
}

func newDoProjectBidLocationMaterialTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project bid location material type",
		Long: `Create a project bid location material type.

Required flags:
  --project-bid-location  Project bid location ID (required)
  --material-type         Material type ID (required)

Optional flags:
  --unit-of-measure       Unit of measure ID
  --quantity              Planned quantity
  --notes                 Notes`,
		Example: `  # Create a project bid location material type
  xbe do project-bid-location-material-types create \
    --project-bid-location 123 \
    --material-type 456 \
    --quantity 12.5

  # Create with notes and unit of measure
  xbe do project-bid-location-material-types create \
    --project-bid-location 123 \
    --material-type 456 \
    --unit-of-measure 789 \
    --quantity 12.5 \
    --notes "Initial estimate"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectBidLocationMaterialTypesCreate,
	}
	initDoProjectBidLocationMaterialTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectBidLocationMaterialTypesCmd.AddCommand(newDoProjectBidLocationMaterialTypesCreateCmd())
}

func initDoProjectBidLocationMaterialTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-bid-location", "", "Project bid location ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectBidLocationMaterialTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectBidLocationMaterialTypesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectBidLocation) == "" {
		err := fmt.Errorf("--project-bid-location is required")
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
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}

	relationships := map[string]any{
		"project-bid-location": map[string]any{
			"data": map[string]any{
				"type": "project-bid-locations",
				"id":   opts.ProjectBidLocation,
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-bid-location-material-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-bid-location-material-types", jsonBody)
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

	row := buildProjectBidLocationMaterialTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project bid location material type %s\n", row.ID)
	return nil
}

func parseDoProjectBidLocationMaterialTypesCreateOptions(cmd *cobra.Command) (doProjectBidLocationMaterialTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectBidLocation, _ := cmd.Flags().GetString("project-bid-location")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectBidLocationMaterialTypesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ProjectBidLocation: projectBidLocation,
		MaterialType:       materialType,
		UnitOfMeasure:      unitOfMeasure,
		Quantity:           quantity,
		Notes:              notes,
	}, nil
}

func buildProjectBidLocationMaterialTypeRowFromSingle(resp jsonAPISingleResponse) projectBidLocationMaterialTypeRow {
	attrs := resp.Data.Attributes

	row := projectBidLocationMaterialTypeRow{
		ID:       resp.Data.ID,
		Quantity: stringAttr(attrs, "quantity"),
		Notes:    stringAttr(attrs, "notes"),
	}

	if rel, ok := resp.Data.Relationships["project-bid-location"]; ok && rel.Data != nil {
		row.ProjectBidLocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}

	return row
}
