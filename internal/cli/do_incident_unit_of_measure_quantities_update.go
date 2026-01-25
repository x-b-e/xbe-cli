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

type doIncidentUnitOfMeasureQuantitiesUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Quantity      string
	UnitOfMeasure string
	IncidentType  string
	IncidentID    string
}

func newDoIncidentUnitOfMeasureQuantitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an incident unit of measure quantity",
		Long: `Update an incident unit of measure quantity.

Optional flags:
  --quantity         Update quantity value (empty to clear)
  --unit-of-measure  Update unit of measure ID (empty to clear)
  --incident-type    Update incident type (use with --incident-id)
  --incident-id      Update incident ID (use with --incident-type)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity
  xbe do incident-unit-of-measure-quantities update 123 --quantity 15

  # Update unit of measure
  xbe do incident-unit-of-measure-quantities update 123 --unit-of-measure 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentUnitOfMeasureQuantitiesUpdate,
	}
	initDoIncidentUnitOfMeasureQuantitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doIncidentUnitOfMeasureQuantitiesCmd.AddCommand(newDoIncidentUnitOfMeasureQuantitiesUpdateCmd())
}

func initDoIncidentUnitOfMeasureQuantitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Quantity value")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("incident-type", "", "Incident type")
	cmd.Flags().String("incident-id", "", "Incident ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentUnitOfMeasureQuantitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentUnitOfMeasureQuantitiesUpdateOptions(cmd, args)
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
		if strings.TrimSpace(opts.Quantity) == "" {
			attributes["quantity"] = nil
		} else {
			attributes["quantity"] = opts.Quantity
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

	incidentTypeChanged := cmd.Flags().Changed("incident-type")
	incidentIDChanged := cmd.Flags().Changed("incident-id")
	if incidentTypeChanged != incidentIDChanged {
		err := fmt.Errorf("--incident-type and --incident-id must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if incidentTypeChanged && incidentIDChanged {
		if strings.TrimSpace(opts.IncidentType) == "" && strings.TrimSpace(opts.IncidentID) == "" {
			relationships["incident"] = map[string]any{"data": nil}
		} else if strings.TrimSpace(opts.IncidentType) == "" || strings.TrimSpace(opts.IncidentID) == "" {
			err := fmt.Errorf("--incident-type and --incident-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			relationships["incident"] = map[string]any{
				"data": map[string]any{
					"type": opts.IncidentType,
					"id":   opts.IncidentID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "incident-unit-of-measure-quantities",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/incident-unit-of-measure-quantities/"+opts.ID, jsonBody)
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

	row := buildIncidentUnitOfMeasureQuantityRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated incident unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoIncidentUnitOfMeasureQuantitiesUpdateOptions(cmd *cobra.Command, args []string) (doIncidentUnitOfMeasureQuantitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentUnitOfMeasureQuantitiesUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Quantity:      quantity,
		UnitOfMeasure: unitOfMeasure,
		IncidentType:  incidentType,
		IncidentID:    incidentID,
	}, nil
}
