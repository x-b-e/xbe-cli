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

type doDeereEquipmentsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	EquipmentID string
}

func newDoDeereEquipmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update Deere equipment",
		Long: `Update Deere equipment metadata or assignment.

Optional flags:
  --equipment  Assigned equipment ID (use empty string to clear)`,
		Example: `  # Update equipment assignment
  xbe do deere-equipments update 123 --equipment 456

  # Clear equipment assignment
  xbe do deere-equipments update 123 --equipment ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeereEquipmentsUpdate,
	}
	initDoDeereEquipmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeereEquipmentsCmd.AddCommand(newDoDeereEquipmentsUpdateCmd())
}

func initDoDeereEquipmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Assigned equipment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeereEquipmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeereEquipmentsUpdateOptions(cmd, args)
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
		err := fmt.Errorf("no fields to update; specify --equipment")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "deere-equipments",
		"id":   opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/deere-equipments/"+opts.ID, jsonBody)
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

	row := deereEquipmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated Deere equipment %s\n", row.ID)
	return nil
}

func parseDoDeereEquipmentsUpdateOptions(cmd *cobra.Command, args []string) (doDeereEquipmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentID, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeereEquipmentsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		EquipmentID: equipmentID,
	}, nil
}
