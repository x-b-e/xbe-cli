package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type materialTypeUnavailabilitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTypeUnavailabilityDetails struct {
	ID             string `json:"id"`
	MaterialTypeID string `json:"material_type_id,omitempty"`
	MaterialType   string `json:"material_type,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	Description    string `json:"description,omitempty"`
}

func newMaterialTypeUnavailabilitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material type unavailability details",
		Long: `Show the full details of a material type unavailability.

Output Fields:
  ID             Material type unavailability identifier
  Material Type  Material type (name or ID)
  Start At       Start timestamp
  End At         End timestamp
  Description    Description

Arguments:
  <id>  Material type unavailability ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a material type unavailability
  xbe view material-type-unavailabilities show 123

  # JSON output
  xbe view material-type-unavailabilities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTypeUnavailabilitiesShow,
	}
	initMaterialTypeUnavailabilitiesShowFlags(cmd)
	return cmd
}

func init() {
	materialTypeUnavailabilitiesCmd.AddCommand(newMaterialTypeUnavailabilitiesShowCmd())
}

func initMaterialTypeUnavailabilitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeUnavailabilitiesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialTypeUnavailabilitiesShowOptions(cmd)
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
		return fmt.Errorf("material type unavailability id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-type-unavailabilities]", "start-at,end-at,description,material-type")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("include", "material-type")

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-unavailabilities/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildMaterialTypeUnavailabilityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTypeUnavailabilityDetails(cmd, details)
}

func parseMaterialTypeUnavailabilitiesShowOptions(cmd *cobra.Command) (materialTypeUnavailabilitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeUnavailabilitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTypeUnavailabilityDetails(resp jsonAPISingleResponse) materialTypeUnavailabilityDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := materialTypeUnavailabilityDetails{
		ID:          resp.Data.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = firstNonEmpty(
				stringAttr(materialType.Attributes, "display-name"),
				stringAttr(materialType.Attributes, "name"),
			)
		}
	}

	return details
}

func renderMaterialTypeUnavailabilityDetails(cmd *cobra.Command, details materialTypeUnavailabilityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialType != "" && details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type: %s (%s)\n", details.MaterialType, details.MaterialTypeID)
	} else if details.MaterialType != "" {
		fmt.Fprintf(out, "Material Type: %s\n", details.MaterialType)
	} else if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
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

	return nil
}
