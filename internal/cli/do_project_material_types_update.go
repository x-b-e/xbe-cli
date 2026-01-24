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

type doProjectMaterialTypesUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
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

func newDoProjectMaterialTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project material type",
		Long: `Update a project material type.

Note: project and material type cannot be changed after creation.

Optional flags:
  --quantity               Quantity
  --explicit-display-name  Explicit display name override (use empty to clear)
  --pickup-at-min          Earliest pickup time (ISO 8601, use empty to clear)
  --pickup-at-max          Latest pickup time (ISO 8601, use empty to clear)
  --deliver-at-min         Earliest delivery time (ISO 8601, use empty to clear)
  --deliver-at-max         Latest delivery time (ISO 8601, use empty to clear)
  --unit-of-measure        Unit of measure ID (use empty to clear)
  --material-site          Material site ID (use empty to clear)
  --job-site               Job site ID (use empty to clear)
  --pickup-location        Pickup location ID (use empty to clear)
  --delivery-location      Delivery location ID (use empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity and display name
  xbe do project-material-types update 123 --quantity 15 --explicit-display-name "Washed Rock"

  # Update pickup/delivery windows
  xbe do project-material-types update 123 --pickup-at-min 2026-01-01T08:00:00Z --deliver-at-max 2026-01-01T18:00:00Z

  # Clear unit of measure
  xbe do project-material-types update 123 --unit-of-measure ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectMaterialTypesUpdate,
	}
	initDoProjectMaterialTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectMaterialTypesCmd.AddCommand(newDoProjectMaterialTypesUpdateCmd())
}

func initDoProjectMaterialTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Quantity")
	cmd.Flags().String("explicit-display-name", "", "Explicit display name override (use empty to clear)")
	cmd.Flags().String("pickup-at-min", "", "Earliest pickup time (ISO 8601, use empty to clear)")
	cmd.Flags().String("pickup-at-max", "", "Latest pickup time (ISO 8601, use empty to clear)")
	cmd.Flags().String("deliver-at-min", "", "Earliest delivery time (ISO 8601, use empty to clear)")
	cmd.Flags().String("deliver-at-max", "", "Latest delivery time (ISO 8601, use empty to clear)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (use empty to clear)")
	cmd.Flags().String("material-site", "", "Material site ID (use empty to clear)")
	cmd.Flags().String("job-site", "", "Job site ID (use empty to clear)")
	cmd.Flags().String("pickup-location", "", "Pickup location ID (use empty to clear)")
	cmd.Flags().String("delivery-location", "", "Delivery location ID (use empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectMaterialTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectMaterialTypesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("explicit-display-name") {
		attributes["explicit-display-name"] = opts.ExplicitDisplayName
	}
	if cmd.Flags().Changed("pickup-at-min") {
		attributes["pickup-at-min"] = opts.PickupAtMin
	}
	if cmd.Flags().Changed("pickup-at-max") {
		attributes["pickup-at-max"] = opts.PickupAtMax
	}
	if cmd.Flags().Changed("deliver-at-min") {
		attributes["deliver-at-min"] = opts.DeliverAtMin
	}
	if cmd.Flags().Changed("deliver-at-max") {
		attributes["deliver-at-max"] = opts.DeliverAtMax
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

	if cmd.Flags().Changed("material-site") {
		if strings.TrimSpace(opts.MaterialSite) == "" {
			relationships["material-site"] = map[string]any{"data": nil}
		} else {
			relationships["material-site"] = map[string]any{
				"data": map[string]any{
					"type": "material-sites",
					"id":   opts.MaterialSite,
				},
			}
		}
	}

	if cmd.Flags().Changed("job-site") {
		if strings.TrimSpace(opts.JobSite) == "" {
			relationships["job-site"] = map[string]any{"data": nil}
		} else {
			relationships["job-site"] = map[string]any{
				"data": map[string]any{
					"type": "job-sites",
					"id":   opts.JobSite,
				},
			}
		}
	}

	if cmd.Flags().Changed("pickup-location") {
		if strings.TrimSpace(opts.PickupLocation) == "" {
			relationships["pickup-location"] = map[string]any{"data": nil}
		} else {
			relationships["pickup-location"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-locations",
					"id":   opts.PickupLocation,
				},
			}
		}
	}

	if cmd.Flags().Changed("delivery-location") {
		if strings.TrimSpace(opts.DeliveryLocation) == "" {
			relationships["delivery-location"] = map[string]any{"data": nil}
		} else {
			relationships["delivery-location"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-locations",
					"id":   opts.DeliveryLocation,
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
		"type": "project-material-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-material-types/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project material type %s\n", row.ID)
	return nil
}

func parseDoProjectMaterialTypesUpdateOptions(cmd *cobra.Command, args []string) (doProjectMaterialTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
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

	return doProjectMaterialTypesUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
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
