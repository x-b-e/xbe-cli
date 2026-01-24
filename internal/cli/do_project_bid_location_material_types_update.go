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

type doProjectBidLocationMaterialTypesUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	MaterialType  string
	UnitOfMeasure string
	Quantity      string
	Notes         string
}

func newDoProjectBidLocationMaterialTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project bid location material type",
		Long: `Update a project bid location material type.

Note: project bid location cannot be changed after creation.

Optional flags:
  --material-type    Material type ID
  --unit-of-measure  Unit of measure ID (use empty to clear)
  --quantity         Planned quantity
  --notes            Notes

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity and notes
  xbe do project-bid-location-material-types update 123 --quantity 15 --notes "Updated"

  # Update material type
  xbe do project-bid-location-material-types update 123 --material-type 456

  # Clear unit of measure
  xbe do project-bid-location-material-types update 123 --unit-of-measure ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectBidLocationMaterialTypesUpdate,
	}
	initDoProjectBidLocationMaterialTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectBidLocationMaterialTypesCmd.AddCommand(newDoProjectBidLocationMaterialTypesUpdateCmd())
}

func initDoProjectBidLocationMaterialTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (use empty to clear)")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectBidLocationMaterialTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectBidLocationMaterialTypesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}

	if cmd.Flags().Changed("material-type") {
		if strings.TrimSpace(opts.MaterialType) == "" {
			err := fmt.Errorf("material-type id is required when updating material-type")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}

	if cmd.Flags().Changed("unit-of-measure") {
		if strings.TrimSpace(opts.UnitOfMeasure) == "" {
			relationships["unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasure,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-bid-location-material-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-bid-location-material-types/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project bid location material type %s\n", row.ID)
	return nil
}

func parseDoProjectBidLocationMaterialTypesUpdateOptions(cmd *cobra.Command, args []string) (doProjectBidLocationMaterialTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectBidLocationMaterialTypesUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		MaterialType:  materialType,
		UnitOfMeasure: unitOfMeasure,
		Quantity:      quantity,
		Notes:         notes,
	}, nil
}
