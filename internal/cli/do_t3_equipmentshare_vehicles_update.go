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

type doT3EquipmentshareVehiclesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Trailer string
	Tractor string
}

func newDoT3EquipmentshareVehiclesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update T3 EquipmentShare vehicle assignments",
		Long: `Update T3 EquipmentShare vehicle assignments.

Common flags:
  --trailer    Update the trailer assignment
  --tractor    Update the tractor assignment`,
		Example: `  # Update trailer assignment
  xbe do t3-equipmentshare-vehicles update 123 --trailer 456

  # Update tractor assignment
  xbe do t3-equipmentshare-vehicles update 123 --tractor 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoT3EquipmentshareVehiclesUpdate,
	}
	initDoT3EquipmentshareVehiclesUpdateFlags(cmd)
	return cmd
}

func init() {
	doT3EquipmentshareVehiclesCmd.AddCommand(newDoT3EquipmentshareVehiclesUpdateCmd())
}

func initDoT3EquipmentshareVehiclesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoT3EquipmentshareVehiclesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoT3EquipmentshareVehiclesUpdateOptions(cmd, args)
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

	if len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "t3-equipmentshare-vehicles",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/t3-equipmentshare-vehicles/"+opts.ID, jsonBody)
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
		row := t3EquipmentshareVehicleRow{
			ID:                    resp.Data.ID,
			VehicleID:             stringAttr(resp.Data.Attributes, "vehicle-id"),
			VehicleNumber:         stringAttr(resp.Data.Attributes, "vehicle-number"),
			IntegrationIdentifier: stringAttr(resp.Data.Attributes, "integration-identifier"),
			TrailerSetAt:          formatDateTime(stringAttr(resp.Data.Attributes, "trailer-set-at")),
			TractorSetAt:          formatDateTime(stringAttr(resp.Data.Attributes, "tractor-set-at")),
			BrokerID:              relationshipIDFromMap(resp.Data.Relationships, "broker"),
			TruckerID:             relationshipIDFromMap(resp.Data.Relationships, "trucker"),
			TrailerID:             relationshipIDFromMap(resp.Data.Relationships, "trailer"),
			TractorID:             relationshipIDFromMap(resp.Data.Relationships, "tractor"),
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated T3 EquipmentShare vehicle %s\n", resp.Data.ID)
	return nil
}

func parseDoT3EquipmentshareVehiclesUpdateOptions(cmd *cobra.Command, args []string) (doT3EquipmentshareVehiclesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doT3EquipmentshareVehiclesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Trailer: trailer,
		Tractor: tractor,
	}, nil
}
