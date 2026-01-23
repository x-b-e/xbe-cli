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

type doMaterialSiteUnavailabilitiesCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	MaterialSiteID string
	StartAt        string
	EndAt          string
	Description    string
}

func newDoMaterialSiteUnavailabilitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site unavailability",
		Long: `Create a material site unavailability.

Required flags:
  --material-site  Material site ID (required)

Optional flags:
  --start-at      Start timestamp (RFC3339)
  --end-at        End timestamp (RFC3339)
  --description   Description

Notes:
  You must supply at least one of --start-at or --end-at.`,
		Example: `  # Create a material site unavailability
  xbe do material-site-unavailabilities create --material-site 123 \
    --start-at 2026-01-24T08:00:00Z --end-at 2026-01-24T12:00:00Z

  # Create an open-ended unavailability
  xbe do material-site-unavailabilities create --material-site 123 \
    --start-at 2026-01-24T08:00:00Z --description "Planned maintenance"

  # Output as JSON
  xbe do material-site-unavailabilities create --material-site 123 --start-at 2026-01-24T08:00:00Z --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSiteUnavailabilitiesCreate,
	}
	initDoMaterialSiteUnavailabilitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteUnavailabilitiesCmd.AddCommand(newDoMaterialSiteUnavailabilitiesCreateCmd())
}

func initDoMaterialSiteUnavailabilitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (RFC3339)")
	cmd.Flags().String("end-at", "", "End timestamp (RFC3339)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-site")
}

func runDoMaterialSiteUnavailabilitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteUnavailabilitiesCreateOptions(cmd)
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

	if opts.MaterialSiteID == "" {
		err := fmt.Errorf("--material-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.StartAt == "" && opts.EndAt == "" {
		err := fmt.Errorf("either --start-at or --end-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-site-unavailabilities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-unavailabilities", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site unavailability %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteUnavailabilitiesCreateOptions(cmd *cobra.Command) (doMaterialSiteUnavailabilitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteUnavailabilitiesCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		MaterialSiteID: materialSiteID,
		StartAt:        startAt,
		EndAt:          endAt,
		Description:    description,
	}, nil
}
