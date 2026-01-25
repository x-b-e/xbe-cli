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

type doMaterialTypeUnavailabilitiesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	StartAt     string
	EndAt       string
	Description string
}

func newDoMaterialTypeUnavailabilitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material type unavailability",
		Long: `Update a material type unavailability.

Optional:
  --start-at      Start timestamp (ISO 8601)
  --end-at        End timestamp (ISO 8601)
  --description   Description`,
		Example: `  # Update description
  xbe do material-type-unavailabilities update 123 --description "Updated description"

  # Update time window
  xbe do material-type-unavailabilities update 123 --start-at 2025-01-02T00:00:00Z --end-at 2025-01-03T00:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTypeUnavailabilitiesUpdate,
	}
	initDoMaterialTypeUnavailabilitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeUnavailabilitiesCmd.AddCommand(newDoMaterialTypeUnavailabilitiesUpdateCmd())
}

func initDoMaterialTypeUnavailabilitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTypeUnavailabilitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTypeUnavailabilitiesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "material-type-unavailabilities",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-type-unavailabilities/"+opts.ID, jsonBody)
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
		row := materialTypeUnavailabilityRow{
			ID:          resp.Data.ID,
			StartAt:     formatDateTime(stringAttr(resp.Data.Attributes, "start-at")),
			EndAt:       formatDateTime(stringAttr(resp.Data.Attributes, "end-at")),
			Description: stringAttr(resp.Data.Attributes, "description"),
		}
		if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material type unavailability %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialTypeUnavailabilitiesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTypeUnavailabilitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeUnavailabilitiesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		StartAt:     startAt,
		EndAt:       endAt,
		Description: description,
	}, nil
}
