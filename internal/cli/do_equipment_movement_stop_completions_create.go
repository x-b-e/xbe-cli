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

type doEquipmentMovementStopCompletionsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	StopID      string
	CompletedAt string
	Latitude    string
	Longitude   string
	Note        string
}

func newDoEquipmentMovementStopCompletionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement stop completion",
		Long: `Create an equipment movement stop completion.

Required flags:
  --stop          Stop ID (required)
  --completed-at  Completion timestamp (ISO 8601)

Optional flags:
  --latitude      Completion latitude
  --longitude     Completion longitude
  --note          Completion note`,
		Example: `  # Create a stop completion
  xbe do equipment-movement-stop-completions create \
    --stop 123 \
    --completed-at 2026-01-23T12:34:56Z

  # Create with coordinates and note
  xbe do equipment-movement-stop-completions create \
    --stop 123 \
    --completed-at 2026-01-23T12:34:56Z \
    --latitude 34.05 \
    --longitude -118.25 \
    --note "Arrived at destination"`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementStopCompletionsCreate,
	}
	initDoEquipmentMovementStopCompletionsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopCompletionsCmd.AddCommand(newDoEquipmentMovementStopCompletionsCreateCmd())
}

func initDoEquipmentMovementStopCompletionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("stop", "", "Stop ID (required)")
	cmd.Flags().String("completed-at", "", "Completion timestamp (ISO 8601)")
	cmd.Flags().String("latitude", "", "Completion latitude")
	cmd.Flags().String("longitude", "", "Completion longitude")
	cmd.Flags().String("note", "", "Completion note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementStopCompletionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementStopCompletionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.StopID) == "" {
		err := fmt.Errorf("--stop is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.CompletedAt) == "" {
		err := fmt.Errorf("--completed-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if (strings.TrimSpace(opts.Latitude) == "") != (strings.TrimSpace(opts.Longitude) == "") {
		err := fmt.Errorf("--latitude and --longitude must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"completed-at": opts.CompletedAt,
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if opts.Latitude != "" && opts.Longitude != "" {
		attributes["latitude"] = opts.Latitude
		attributes["longitude"] = opts.Longitude
	}

	relationships := map[string]any{
		"stop": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-stops",
				"id":   opts.StopID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-stop-completions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-stop-completions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement stop completion %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementStopCompletionsCreateOptions(cmd *cobra.Command) (doEquipmentMovementStopCompletionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	stopID, _ := cmd.Flags().GetString("stop")
	completedAt, _ := cmd.Flags().GetString("completed-at")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementStopCompletionsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		StopID:      stopID,
		CompletedAt: completedAt,
		Latitude:    latitude,
		Longitude:   longitude,
		Note:        note,
	}, nil
}

func buildEquipmentMovementStopCompletionRowFromSingle(resp jsonAPISingleResponse) equipmentMovementStopCompletionRow {
	attrs := resp.Data.Attributes

	row := equipmentMovementStopCompletionRow{
		ID:          resp.Data.ID,
		CompletedAt: formatDateTime(stringAttr(attrs, "completed-at")),
		Latitude:    stringAttr(attrs, "latitude"),
		Longitude:   stringAttr(attrs, "longitude"),
		Note:        stringAttr(attrs, "note"),
	}

	if rel, ok := resp.Data.Relationships["stop"]; ok && rel.Data != nil {
		row.StopID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
