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

type doMaterialTypesUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	Name                string
	ExplicitDisplayName string
	Description         string
	AggregateBed        string
	AggregateGradation  string
	AggregateECCE       string
	LbsPerCubicFoot     string
	StartSiteType       string
	IsArchived          string
	ParentMaterialType  string
	MaterialSupplier    string
}

func newDoMaterialTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material type",
		Long: `Update a material type.

All flags are optional. Only provided flags will update the material type.

Optional flags:
  --name                   Material type name
  --explicit-display-name  Display name override
  --description            Description
  --aggregate-bed          Aggregate bed
  --aggregate-gradation    Aggregate gradation
  --aggregate-ecce         Aggregate ECCE
  --lbs-per-cubic-foot     Pounds per cubic foot
  --start-site-type        Start site type
  --is-archived            Archive status (true/false)

Relationships:
  --parent-material-type   Parent material type ID
  --material-supplier      Material supplier ID`,
		Example: `  # Update material type name
  xbe do material-types update 123 --name "New Name"

  # Archive a material type
  xbe do material-types update 123 --is-archived true

  # Update display name
  xbe do material-types update 123 --explicit-display-name "Display Name"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTypesUpdate,
	}
	initDoMaterialTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypesCmd.AddCommand(newDoMaterialTypesUpdateCmd())
}

func initDoMaterialTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material type name")
	cmd.Flags().String("explicit-display-name", "", "Display name override")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("aggregate-bed", "", "Aggregate bed")
	cmd.Flags().String("aggregate-gradation", "", "Aggregate gradation")
	cmd.Flags().String("aggregate-ecce", "", "Aggregate ECCE")
	cmd.Flags().String("lbs-per-cubic-foot", "", "Pounds per cubic foot")
	cmd.Flags().String("start-site-type", "", "Start site type")
	cmd.Flags().String("is-archived", "", "Archive status (true/false)")
	cmd.Flags().String("parent-material-type", "", "Parent material type ID")
	cmd.Flags().String("material-supplier", "", "Material supplier ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTypesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("explicit-display-name") {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("aggregate-bed") {
		attributes["aggregate-bed"] = opts.AggregateBed
	}
	if cmd.Flags().Changed("aggregate-gradation") {
		attributes["aggregate-gradation"] = opts.AggregateGradation
	}
	if cmd.Flags().Changed("aggregate-ecce") {
		attributes["aggregate-ecce"] = opts.AggregateECCE
	}
	if cmd.Flags().Changed("lbs-per-cubic-foot") {
		attributes["lbs-per-cubic-foot"] = opts.LbsPerCubicFoot
	}
	if cmd.Flags().Changed("start-site-type") {
		attributes["start-site-type"] = opts.StartSiteType
	}
	if cmd.Flags().Changed("is-archived") {
		attributes["is-archived"] = opts.IsArchived == "true"
	}

	if cmd.Flags().Changed("parent-material-type") {
		if opts.ParentMaterialType == "" {
			relationships["parent-material-type"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["parent-material-type"] = map[string]any{
				"data": map[string]any{
					"type": "material-types",
					"id":   opts.ParentMaterialType,
				},
			}
		}
	}
	if cmd.Flags().Changed("material-supplier") {
		if opts.MaterialSupplier == "" {
			relationships["material-supplier"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["material-supplier"] = map[string]any{
				"data": map[string]any{
					"type": "material-suppliers",
					"id":   opts.MaterialSupplier,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "material-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-types/"+opts.ID, jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":   resp.Data.ID,
			"name": stringAttr(resp.Data.Attributes, "name"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material type %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "name"))
	return nil
}

func parseDoMaterialTypesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	explicitDisplayName, _ := cmd.Flags().GetString("explicit-display-name")
	description, _ := cmd.Flags().GetString("description")
	aggregateBed, _ := cmd.Flags().GetString("aggregate-bed")
	aggregateGradation, _ := cmd.Flags().GetString("aggregate-gradation")
	aggregateECCE, _ := cmd.Flags().GetString("aggregate-ecce")
	lbsPerCubicFoot, _ := cmd.Flags().GetString("lbs-per-cubic-foot")
	startSiteType, _ := cmd.Flags().GetString("start-site-type")
	isArchived, _ := cmd.Flags().GetString("is-archived")
	parentMaterialType, _ := cmd.Flags().GetString("parent-material-type")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypesUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		Name:                name,
		ExplicitDisplayName: explicitDisplayName,
		Description:         description,
		AggregateBed:        aggregateBed,
		AggregateGradation:  aggregateGradation,
		AggregateECCE:       aggregateECCE,
		LbsPerCubicFoot:     lbsPerCubicFoot,
		StartSiteType:       startSiteType,
		IsArchived:          isArchived,
		ParentMaterialType:  parentMaterialType,
		MaterialSupplier:    materialSupplier,
	}, nil
}
