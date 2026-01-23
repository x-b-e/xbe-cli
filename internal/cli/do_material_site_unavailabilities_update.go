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

type doMaterialSiteUnavailabilitiesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	StartAt     string
	EndAt       string
	Description string
}

func newDoMaterialSiteUnavailabilitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material site unavailability",
		Long: `Update a material site unavailability.

Optional flags:
  --start-at     Start timestamp (RFC3339)
  --end-at       End timestamp (RFC3339)
  --description  Description`,
		Example: `  # Update description
  xbe do material-site-unavailabilities update 123 --description "Updated notes"

  # Update end timestamp
  xbe do material-site-unavailabilities update 123 --end-at 2026-01-24T12:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSiteUnavailabilitiesUpdate,
	}
	initDoMaterialSiteUnavailabilitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteUnavailabilitiesCmd.AddCommand(newDoMaterialSiteUnavailabilitiesUpdateCmd())
}

func initDoMaterialSiteUnavailabilitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (RFC3339)")
	cmd.Flags().String("end-at", "", "End timestamp (RFC3339)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteUnavailabilitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteUnavailabilitiesUpdateOptions(cmd, args)
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
			"type":       "material-site-unavailabilities",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-site-unavailabilities/"+opts.ID, jsonBody)
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

	row := buildMaterialSiteUnavailabilityRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material site unavailability %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteUnavailabilitiesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSiteUnavailabilitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteUnavailabilitiesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		StartAt:     startAt,
		EndAt:       endAt,
		Description: description,
	}, nil
}
