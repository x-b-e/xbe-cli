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

type doTennaVehiclesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	TractorID   string
	TrailerID   string
	EquipmentID string
}

func newDoTennaVehiclesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update Tenna vehicle assignments",
		Long: `Update Tenna vehicle trailer, tractor, or equipment assignments.

Optional flags:
  --tractor   Tractor ID (use empty string to clear)
  --trailer   Trailer ID (use empty string to clear)
  --equipment Equipment ID (use empty string to clear)`,
		Example: `  # Update tractor assignment
  xbe do tenna-vehicles update 123 --tractor 456

  # Update trailer assignment
  xbe do tenna-vehicles update 123 --trailer 789

  # Update equipment assignment
  xbe do tenna-vehicles update 123 --equipment 321

  # Clear equipment assignment
  xbe do tenna-vehicles update 123 --equipment ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTennaVehiclesUpdate,
	}
	initDoTennaVehiclesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTennaVehiclesCmd.AddCommand(newDoTennaVehiclesUpdateCmd())
}

func initDoTennaVehiclesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTennaVehiclesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTennaVehiclesUpdateOptions(cmd, args)
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
	var hasChanges bool

	if cmd.Flags().Changed("tractor") {
		if strings.TrimSpace(opts.TractorID) == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]any{
					"type": "tractors",
					"id":   opts.TractorID,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("trailer") {
		if strings.TrimSpace(opts.TrailerID) == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.TrailerID,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("equipment") {
		if strings.TrimSpace(opts.EquipmentID) == "" {
			relationships["equipment"] = map[string]any{"data": nil}
		} else {
			relationships["equipment"] = map[string]any{
				"data": map[string]any{
					"type": "equipment",
					"id":   opts.EquipmentID,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify --tractor, --trailer, or --equipment")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "tenna-vehicles",
		"id":   opts.ID,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tenna-vehicles/"+opts.ID, jsonBody)
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

	row := tennaVehicleRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated Tenna vehicle %s\n", row.ID)
	return nil
}

func parseDoTennaVehiclesUpdateOptions(cmd *cobra.Command, args []string) (doTennaVehiclesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tractorID, _ := cmd.Flags().GetString("tractor")
	trailerID, _ := cmd.Flags().GetString("trailer")
	equipmentID, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTennaVehiclesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		TractorID:   tractorID,
		TrailerID:   trailerID,
		EquipmentID: equipmentID,
	}, nil
}
