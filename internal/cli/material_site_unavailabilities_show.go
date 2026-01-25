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

type materialSiteUnavailabilitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteUnavailabilityDetails struct {
	ID             string `json:"id"`
	MaterialSiteID string `json:"material_site_id,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	Description    string `json:"description,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

func newMaterialSiteUnavailabilitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site unavailability details",
		Long: `Show the full details of a material site unavailability.

Output Fields:
  ID             Unavailability identifier
  Material Site  Material site ID
  Start At       Start timestamp
  End At         End timestamp
  Description    Description
  Created At     Created timestamp
  Updated At     Updated timestamp

Arguments:
  <id>    The material site unavailability ID (required). You can find IDs using the list command.`,
		Example: `  # Show a material site unavailability
  xbe view material-site-unavailabilities show 123

  # Get JSON output
  xbe view material-site-unavailabilities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteUnavailabilitiesShow,
	}
	initMaterialSiteUnavailabilitiesShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteUnavailabilitiesCmd.AddCommand(newMaterialSiteUnavailabilitiesShowCmd())
}

func initMaterialSiteUnavailabilitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteUnavailabilitiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSiteUnavailabilitiesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material site unavailability id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-unavailabilities/"+id, nil)
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

	details := buildMaterialSiteUnavailabilityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteUnavailabilityDetails(cmd, details)
}

func parseMaterialSiteUnavailabilitiesShowOptions(cmd *cobra.Command) (materialSiteUnavailabilitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteUnavailabilitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteUnavailabilityDetails(resp jsonAPISingleResponse) materialSiteUnavailabilityDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := materialSiteUnavailabilityDetails{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Description: stringAttr(attrs, "description"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
	}

	return details
}

func renderMaterialSiteUnavailabilityDetails(cmd *cobra.Command, details materialSiteUnavailabilityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSiteID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
