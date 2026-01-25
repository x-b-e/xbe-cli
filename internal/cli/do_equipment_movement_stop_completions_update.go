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

type doEquipmentMovementStopCompletionsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	CompletedAt string
	Latitude    string
	Longitude   string
	Note        string
}

func newDoEquipmentMovementStopCompletionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement stop completion",
		Long: `Update an equipment movement stop completion.

Arguments:
  <id>    The stop completion ID (required).

Optional flags:
  --completed-at  Completion timestamp (ISO 8601)
  --latitude      Completion latitude
  --longitude     Completion longitude
  --note          Completion note`,
		Example: `  # Update completion timestamp
  xbe do equipment-movement-stop-completions update 123 --completed-at 2026-01-23T13:00:00Z

  # Update coordinates and note
  xbe do equipment-movement-stop-completions update 123 \
    --latitude 34.06 \
    --longitude -118.26 \
    --note "Updated completion note"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementStopCompletionsUpdate,
	}
	initDoEquipmentMovementStopCompletionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopCompletionsCmd.AddCommand(newDoEquipmentMovementStopCompletionsUpdateCmd())
}

func initDoEquipmentMovementStopCompletionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("completed-at", "", "Completion timestamp (ISO 8601)")
	cmd.Flags().String("latitude", "", "Completion latitude")
	cmd.Flags().String("longitude", "", "Completion longitude")
	cmd.Flags().String("note", "", "Completion note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementStopCompletionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementStopCompletionsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("equipment movement stop completion id is required")
	}

	completedAtChanged := cmd.Flags().Changed("completed-at")
	latitudeChanged := cmd.Flags().Changed("latitude")
	longitudeChanged := cmd.Flags().Changed("longitude")
	noteChanged := cmd.Flags().Changed("note")

	if !completedAtChanged && !latitudeChanged && !longitudeChanged && !noteChanged {
		err := fmt.Errorf("no fields provided to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if completedAtChanged && strings.TrimSpace(opts.CompletedAt) == "" {
		err := fmt.Errorf("--completed-at cannot be empty")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if latitudeChanged != longitudeChanged {
		err := fmt.Errorf("--latitude and --longitude must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if completedAtChanged {
		attributes["completed-at"] = opts.CompletedAt
	}
	if latitudeChanged && longitudeChanged {
		attributes["latitude"] = opts.Latitude
		attributes["longitude"] = opts.Longitude
	}
	if noteChanged {
		attributes["note"] = opts.Note
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-movement-stop-completions",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-stop-completions/"+id, jsonBody)
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

	row := buildEquipmentMovementStopCompletionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement stop completion %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementStopCompletionsUpdateOptions(cmd *cobra.Command) (doEquipmentMovementStopCompletionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	completedAt, _ := cmd.Flags().GetString("completed-at")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementStopCompletionsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		CompletedAt: completedAt,
		Latitude:    latitude,
		Longitude:   longitude,
		Note:        note,
	}, nil
}
