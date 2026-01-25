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

type doEquipmentMovementTripDispatchesUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	Trucker                string
	Driver                 string
	Trailer                string
	TellClerkSynchronously bool
}

func newDoEquipmentMovementTripDispatchesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement trip dispatch",
		Long: `Update an equipment movement trip dispatch.

Optional flags:
  --trucker                   Trucker ID (use empty string to clear)
  --driver                    Driver (user) ID (use empty string to clear)
  --trailer                   Trailer ID (use empty string to clear)
  --tell-clerk-synchronously  Process fulfillment synchronously

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Assign a driver
  xbe do equipment-movement-trip-dispatches update 123 --driver 456

  # Clear the trailer
  xbe do equipment-movement-trip-dispatches update 123 --trailer ""

  # Process synchronously
  xbe do equipment-movement-trip-dispatches update 123 --tell-clerk-synchronously`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementTripDispatchesUpdate,
	}
	initDoEquipmentMovementTripDispatchesUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripDispatchesCmd.AddCommand(newDoEquipmentMovementTripDispatchesUpdateCmd())
}

func initDoEquipmentMovementTripDispatchesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (leave blank to clear)")
	cmd.Flags().String("driver", "", "Driver (user) ID (leave blank to clear)")
	cmd.Flags().String("trailer", "", "Trailer ID (leave blank to clear)")
	cmd.Flags().Bool("tell-clerk-synchronously", false, "Process fulfillment synchronously")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripDispatchesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementTripDispatchesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("tell-clerk-synchronously") {
		attributes["tell-clerk-synchronously"] = opts.TellClerkSynchronously
	}

	if cmd.Flags().Changed("trucker") {
		if strings.TrimSpace(opts.Trucker) == "" {
			relationships["trucker"] = map[string]any{"data": nil}
		} else {
			relationships["trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
	}
	if cmd.Flags().Changed("driver") {
		if strings.TrimSpace(opts.Driver) == "" {
			relationships["driver"] = map[string]any{"data": nil}
		} else {
			relationships["driver"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.Driver,
				},
			}
		}
	}
	if cmd.Flags().Changed("trailer") {
		if strings.TrimSpace(opts.Trailer) == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.Trailer,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment-movement-trip-dispatches",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-trip-dispatches/"+opts.ID, jsonBody)
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

	row := buildEquipmentMovementTripDispatchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement trip dispatch %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementTripDispatchesUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementTripDispatchesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	driver, _ := cmd.Flags().GetString("driver")
	trailer, _ := cmd.Flags().GetString("trailer")
	tellClerkSynchronously, _ := cmd.Flags().GetBool("tell-clerk-synchronously")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripDispatchesUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		Trucker:                trucker,
		Driver:                 driver,
		Trailer:                trailer,
		TellClerkSynchronously: tellClerkSynchronously,
	}, nil
}
