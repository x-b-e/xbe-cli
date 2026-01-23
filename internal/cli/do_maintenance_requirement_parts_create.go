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

type doMaintenanceRequirementPartsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Name                    string
	PartNumber              string
	Description             string
	Notes                   string
	IsTemplate              string
	Make                    string
	Model                   string
	Year                    string
	Broker                  string
	EquipmentClassification string
}

func newDoMaintenanceRequirementPartsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement part",
		Long: `Create a maintenance requirement part.

Required flags:
  --name        Part name (required)

Optional flags:
  --part-number   Part number
  --description   Part description
  --notes         Additional notes
  --is-template   Whether the part is a template (true/false)
  --make          Part make
  --model         Part model
  --year          Part year

Relationships:
  --broker                   Broker ID (required for templates)
  --equipment-classification Equipment classification ID`,
		Example: `  # Create a template part
  xbe do maintenance-requirement-parts create --name "Oil Filter" --is-template true --broker 123

  # Create with full details
  xbe do maintenance-requirement-parts create \
    --name "Air Filter" \
    --part-number "AF-123" \
    --make "ACME" \
    --model "AF-1" \
    --year 2024 \
    --equipment-classification 456 \
    --description "Replacement air filter" \
    --notes "Store in dry area"`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementPartsCreate,
	}
	initDoMaintenanceRequirementPartsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementPartsCmd.AddCommand(newDoMaintenanceRequirementPartsCreateCmd())
}

func initDoMaintenanceRequirementPartsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Part name (required)")
	cmd.Flags().String("part-number", "", "Part number")
	cmd.Flags().String("description", "", "Part description")
	cmd.Flags().String("notes", "", "Additional notes")
	cmd.Flags().String("is-template", "", "Whether the part is a template (true/false)")
	cmd.Flags().String("make", "", "Part make")
	cmd.Flags().String("model", "", "Part model")
	cmd.Flags().String("year", "", "Part year")
	cmd.Flags().String("broker", "", "Broker ID (required for templates)")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
}

func runDoMaintenanceRequirementPartsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementPartsCreateOptions(cmd)
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

	if opts.PartNumber != "" {
		attributes["part-number"] = opts.PartNumber
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.IsTemplate != "" {
		attributes["is-template"] = opts.IsTemplate == "true"
	}
	if opts.Make != "" {
		attributes["make"] = opts.Make
	}
	if opts.Model != "" {
		attributes["model"] = opts.Model
	}
	if opts.Year != "" {
		attributes["year"] = opts.Year
	}

	relationships := map[string]any{}
	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if opts.EquipmentClassification != "" {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassification,
			},
		}
	}

	data := map[string]any{
		"type":       "maintenance-requirement-parts",
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

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-parts", jsonBody)
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

	row := buildMaintenanceRequirementPartRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement part %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementPartsCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementPartsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	partNumber, _ := cmd.Flags().GetString("part-number")
	description, _ := cmd.Flags().GetString("description")
	notes, _ := cmd.Flags().GetString("notes")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	make, _ := cmd.Flags().GetString("make")
	model, _ := cmd.Flags().GetString("model")
	year, _ := cmd.Flags().GetString("year")
	broker, _ := cmd.Flags().GetString("broker")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementPartsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Name:                    name,
		PartNumber:              partNumber,
		Description:             description,
		Notes:                   notes,
		IsTemplate:              isTemplate,
		Make:                    make,
		Model:                   model,
		Year:                    year,
		Broker:                  broker,
		EquipmentClassification: equipmentClassification,
	}, nil
}

func buildMaintenanceRequirementPartRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementPartRow {
	attrs := resp.Data.Attributes

	row := maintenanceRequirementPartRow{
		ID:         resp.Data.ID,
		PartNumber: stringAttr(attrs, "part-number"),
		Name:       stringAttr(attrs, "name"),
		Make:       stringAttr(attrs, "make"),
		Model:      stringAttr(attrs, "model"),
		Year:       stringAttr(attrs, "year"),
		IsTemplate: boolAttr(attrs, "is-template"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		row.EquipmentClassificationID = rel.Data.ID
	}

	return row
}
