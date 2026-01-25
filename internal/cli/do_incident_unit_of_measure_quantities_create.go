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

type doIncidentUnitOfMeasureQuantitiesCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	IncidentType  string
	IncidentID    string
	UnitOfMeasure string
	Quantity      string
}

func newDoIncidentUnitOfMeasureQuantitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an incident unit of measure quantity",
		Long: `Create an incident unit of measure quantity.

Required flags:
  --incident-type    Incident type (use incidents for standard incident records)
  --incident-id      Incident ID
  --unit-of-measure  Unit of measure ID
  --quantity         Quantity value

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an incident unit of measure quantity
  xbe do incident-unit-of-measure-quantities create \
    --incident-type incidents --incident-id 123 \
    --unit-of-measure 456 --quantity 12.5

  # JSON output
  xbe do incident-unit-of-measure-quantities create \
    --incident-type incidents --incident-id 123 \
    --unit-of-measure 456 --quantity 12.5 --json`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentUnitOfMeasureQuantitiesCreate,
	}
	initDoIncidentUnitOfMeasureQuantitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentUnitOfMeasureQuantitiesCmd.AddCommand(newDoIncidentUnitOfMeasureQuantitiesCreateCmd())
}

func initDoIncidentUnitOfMeasureQuantitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident-type", "", "Incident type (required)")
	cmd.Flags().String("incident-id", "", "Incident ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("quantity", "", "Quantity value (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentUnitOfMeasureQuantitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentUnitOfMeasureQuantitiesCreateOptions(cmd)
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

	if opts.IncidentType == "" {
		err := fmt.Errorf("--incident-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IncidentID == "" {
		err := fmt.Errorf("--incident-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.UnitOfMeasure == "" {
		err := fmt.Errorf("--unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Quantity == "" {
		err := fmt.Errorf("--quantity is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"quantity": opts.Quantity,
	}

	relationships := map[string]any{
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
		"incident": map[string]any{
			"data": map[string]any{
				"type": opts.IncidentType,
				"id":   opts.IncidentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-unit-of-measure-quantities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/incident-unit-of-measure-quantities", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoIncidentUnitOfMeasureQuantitiesCreateOptions(cmd *cobra.Command) (doIncidentUnitOfMeasureQuantitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentUnitOfMeasureQuantitiesCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		IncidentType:  incidentType,
		IncidentID:    incidentID,
		UnitOfMeasure: unitOfMeasure,
		Quantity:      quantity,
	}, nil
}
