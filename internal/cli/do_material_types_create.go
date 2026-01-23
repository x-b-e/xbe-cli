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

type doMaterialTypesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Name                string
	ExplicitDisplayName string
	Description         string
	AggregateBed        string
	AggregateGradation  string
	AggregateECCE       string
	LbsPerCubicFoot     string
	StartSiteType       string
	ParentMaterialType  string
	MaterialSupplier    string
}

func newDoMaterialTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new material type",
		Long: `Create a new material type.

Required flags:
  --name                   Material type name

Optional flags:
  --explicit-display-name  Display name override
  --description            Description
  --aggregate-bed          Aggregate bed
  --aggregate-gradation    Aggregate gradation
  --aggregate-ecce         Aggregate ECCE
  --lbs-per-cubic-foot     Pounds per cubic foot
  --start-site-type        Start site type

Relationships:
  --parent-material-type   Parent material type ID
  --material-supplier      Material supplier ID`,
		Example: `  # Create a material type
  xbe do material-types create --name "Asphalt Mix"

  # Create with supplier
  xbe do material-types create --name "Concrete" --material-supplier 123`,
		RunE: runDoMaterialTypesCreate,
	}
	initDoMaterialTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypesCmd.AddCommand(newDoMaterialTypesCreateCmd())
}

func initDoMaterialTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material type name (required)")
	cmd.Flags().String("explicit-display-name", "", "Display name override")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("aggregate-bed", "", "Aggregate bed")
	cmd.Flags().String("aggregate-gradation", "", "Aggregate gradation")
	cmd.Flags().String("aggregate-ecce", "", "Aggregate ECCE")
	cmd.Flags().String("lbs-per-cubic-foot", "", "Pounds per cubic foot")
	cmd.Flags().String("start-site-type", "", "Start site type")
	cmd.Flags().String("parent-material-type", "", "Parent material type ID")
	cmd.Flags().String("material-supplier", "", "Material supplier ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
}

func runDoMaterialTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTypesCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.ExplicitDisplayName != "" {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.AggregateBed != "" {
		attributes["aggregate-bed"] = opts.AggregateBed
	}
	if opts.AggregateGradation != "" {
		attributes["aggregate-gradation"] = opts.AggregateGradation
	}
	if opts.AggregateECCE != "" {
		attributes["aggregate-ecce"] = opts.AggregateECCE
	}
	if opts.LbsPerCubicFoot != "" {
		attributes["lbs-per-cubic-foot"] = opts.LbsPerCubicFoot
	}
	if opts.StartSiteType != "" {
		attributes["start-site-type"] = opts.StartSiteType
	}

	relationships := map[string]any{}

	if opts.ParentMaterialType != "" {
		relationships["parent-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.ParentMaterialType,
			},
		}
	}
	if opts.MaterialSupplier != "" {
		relationships["material-supplier"] = map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.MaterialSupplier,
			},
		}
	}

	data := map[string]any{
		"type":       "material-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material type %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "name"))
	return nil
}

func parseDoMaterialTypesCreateOptions(cmd *cobra.Command) (doMaterialTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	explicitDisplayName, _ := cmd.Flags().GetString("explicit-display-name")
	description, _ := cmd.Flags().GetString("description")
	aggregateBed, _ := cmd.Flags().GetString("aggregate-bed")
	aggregateGradation, _ := cmd.Flags().GetString("aggregate-gradation")
	aggregateECCE, _ := cmd.Flags().GetString("aggregate-ecce")
	lbsPerCubicFoot, _ := cmd.Flags().GetString("lbs-per-cubic-foot")
	startSiteType, _ := cmd.Flags().GetString("start-site-type")
	parentMaterialType, _ := cmd.Flags().GetString("parent-material-type")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Name:                name,
		ExplicitDisplayName: explicitDisplayName,
		Description:         description,
		AggregateBed:        aggregateBed,
		AggregateGradation:  aggregateGradation,
		AggregateECCE:       aggregateECCE,
		LbsPerCubicFoot:     lbsPerCubicFoot,
		StartSiteType:       startSiteType,
		ParentMaterialType:  parentMaterialType,
		MaterialSupplier:    materialSupplier,
	}, nil
}
