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

type doMaintenanceRequirementPartsUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
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

func newDoMaintenanceRequirementPartsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a maintenance requirement part",
		Long: `Update a maintenance requirement part.

Optional flags:
  --name          Part name
  --part-number   Part number
  --description   Part description
  --notes         Additional notes
  --is-template   Whether the part is a template (true/false)
  --make          Part make
  --model         Part model
  --year          Part year

Relationships:
  --broker                   Broker ID
  --equipment-classification Equipment classification ID`,
		Example: `  # Update a part
  xbe do maintenance-requirement-parts update 123 --name "Updated Part"

  # Update with broker and equipment classification
  xbe do maintenance-requirement-parts update 123 --broker 456 --equipment-classification 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementPartsUpdate,
	}
	initDoMaintenanceRequirementPartsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementPartsCmd.AddCommand(newDoMaintenanceRequirementPartsUpdateCmd())
}

func initDoMaintenanceRequirementPartsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Part name")
	cmd.Flags().String("part-number", "", "Part number")
	cmd.Flags().String("description", "", "Part description")
	cmd.Flags().String("notes", "", "Additional notes")
	cmd.Flags().String("is-template", "", "Whether the part is a template (true/false)")
	cmd.Flags().String("make", "", "Part make")
	cmd.Flags().String("model", "", "Part model")
	cmd.Flags().String("year", "", "Part year")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementPartsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementPartsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("part-number") {
		attributes["part-number"] = opts.PartNumber
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("is-template") {
		attributes["is-template"] = opts.IsTemplate == "true"
	}
	if cmd.Flags().Changed("make") {
		attributes["make"] = opts.Make
	}
	if cmd.Flags().Changed("model") {
		attributes["model"] = opts.Model
	}
	if cmd.Flags().Changed("year") {
		attributes["year"] = opts.Year
	}

	var relationships map[string]any
	if cmd.Flags().Changed("broker") {
		if relationships == nil {
			relationships = map[string]any{}
		}
		if opts.Broker == "" {
			relationships["broker"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["broker"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
	}
	if cmd.Flags().Changed("equipment-classification") {
		if relationships == nil {
			relationships = map[string]any{}
		}
		if opts.EquipmentClassification == "" {
			relationships["equipment-classification"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["equipment-classification"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-classifications",
					"id":   opts.EquipmentClassification,
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
		"type": "maintenance-requirement-parts",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/maintenance-requirement-parts/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated maintenance requirement part %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementPartsUpdateOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementPartsUpdateOptions, error) {
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

	return doMaintenanceRequirementPartsUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
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
