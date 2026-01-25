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

type doVerizonRevealVehiclesUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	Trailer   string
	Tractor   string
	Equipment string
}

func newDoVerizonRevealVehiclesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update Verizon Reveal vehicle assignments",
		Long: `Update Verizon Reveal vehicle assignments.

Common flags:
  --trailer     Update the trailer assignment
  --tractor     Update the tractor assignment
  --equipment   Update the equipment assignment`,
		Example: `  # Update trailer assignment
  xbe do verizon-reveal-vehicles update 123 --trailer 456

  # Update tractor assignment
  xbe do verizon-reveal-vehicles update 123 --tractor 789

  # Update equipment assignment
  xbe do verizon-reveal-vehicles update 123 --equipment 321`,
		Args: cobra.ExactArgs(1),
		RunE: runDoVerizonRevealVehiclesUpdate,
	}
	initDoVerizonRevealVehiclesUpdateFlags(cmd)
	return cmd
}

func init() {
	doVerizonRevealVehiclesCmd.AddCommand(newDoVerizonRevealVehiclesUpdateCmd())
}

func initDoVerizonRevealVehiclesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoVerizonRevealVehiclesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoVerizonRevealVehiclesUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("trailer") {
		if strings.TrimSpace(opts.Trailer) == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]string{
					"type": "trailers",
					"id":   opts.Trailer,
				},
			}
		}
	}

	if cmd.Flags().Changed("tractor") {
		if strings.TrimSpace(opts.Tractor) == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]string{
					"type": "tractors",
					"id":   opts.Tractor,
				},
			}
		}
	}

	if cmd.Flags().Changed("equipment") {
		if strings.TrimSpace(opts.Equipment) == "" {
			relationships["equipment"] = map[string]any{"data": nil}
		} else {
			relationships["equipment"] = map[string]any{
				"data": map[string]string{
					"type": "equipment",
					"id":   opts.Equipment,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "verizon-reveal-vehicles",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/verizon-reveal-vehicles/"+opts.ID, jsonBody)
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
		row := verizonRevealVehicleRow{
			ID:                    resp.Data.ID,
			VehicleID:             stringAttr(resp.Data.Attributes, "vehicle-id"),
			VehicleNumber:         stringAttr(resp.Data.Attributes, "vehicle-number"),
			IntegrationIdentifier: stringAttr(resp.Data.Attributes, "integration-identifier"),
			TrailerSetAt:          formatDateTime(stringAttr(resp.Data.Attributes, "trailer-set-at")),
			TractorSetAt:          formatDateTime(stringAttr(resp.Data.Attributes, "tractor-set-at")),
			EquipmentSetAt:        formatDateTime(stringAttr(resp.Data.Attributes, "equipment-set-at")),
			BrokerID:              relationshipIDFromMap(resp.Data.Relationships, "broker"),
			TruckerID:             relationshipIDFromMap(resp.Data.Relationships, "trucker"),
			TrailerID:             relationshipIDFromMap(resp.Data.Relationships, "trailer"),
			TractorID:             relationshipIDFromMap(resp.Data.Relationships, "tractor"),
			EquipmentID:           relationshipIDFromMap(resp.Data.Relationships, "equipment"),
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated Verizon Reveal vehicle %s\n", resp.Data.ID)
	return nil
}

func parseDoVerizonRevealVehiclesUpdateOptions(cmd *cobra.Command, args []string) (doVerizonRevealVehiclesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	equipment, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doVerizonRevealVehiclesUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		Trailer:   trailer,
		Tractor:   tractor,
		Equipment: equipment,
	}, nil
}
