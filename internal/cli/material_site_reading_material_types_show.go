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

type materialSiteReadingMaterialTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteReadingMaterialTypeDetails struct {
	ID               string `json:"id"`
	ExternalID       string `json:"external_id,omitempty"`
	MaterialSiteID   string `json:"material_site_id,omitempty"`
	MaterialSiteName string `json:"material_site_name,omitempty"`
	MaterialTypeID   string `json:"material_type_id,omitempty"`
	MaterialTypeName string `json:"material_type_name,omitempty"`
}

func newMaterialSiteReadingMaterialTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site reading material type details",
		Long: `Show the full details of a material site reading material type.

Includes the external identifier plus the associated material site and
material type.

Arguments:
  <id>  The material site reading material type ID (required).`,
		Example: `  # Show a material site reading material type
  xbe view material-site-reading-material-types show 123

  # Output as JSON
  xbe view material-site-reading-material-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteReadingMaterialTypesShow,
	}
	initMaterialSiteReadingMaterialTypesShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteReadingMaterialTypesCmd.AddCommand(newMaterialSiteReadingMaterialTypesShowCmd())
}

func initMaterialSiteReadingMaterialTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteReadingMaterialTypesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialSiteReadingMaterialTypesShowOptions(cmd)
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
		return fmt.Errorf("material site reading material type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-reading-material-types]", "external-id,material-site,material-type")
	query.Set("include", "material-site,material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name,display-name")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-reading-material-types/"+id, query)
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

	details := buildMaterialSiteReadingMaterialTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteReadingMaterialTypeDetails(cmd, details)
}

func parseMaterialSiteReadingMaterialTypesShowOptions(cmd *cobra.Command) (materialSiteReadingMaterialTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteReadingMaterialTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteReadingMaterialTypeDetails(resp jsonAPISingleResponse) materialSiteReadingMaterialTypeDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialSiteReadingMaterialTypeDetails{
		ID:         resp.Data.ID,
		ExternalID: stringAttr(resp.Data.Attributes, "external-id"),
	}

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = materialTypeLabel(materialType.Attributes)
		}
	}

	return details
}

func renderMaterialSiteReadingMaterialTypeDetails(cmd *cobra.Command, details materialSiteReadingMaterialTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalID != "" {
		fmt.Fprintf(out, "External ID: %s\n", details.ExternalID)
	}
	if details.MaterialSiteID != "" {
		label := details.MaterialSiteID
		if details.MaterialSiteName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSiteName, details.MaterialSiteID)
		}
		fmt.Fprintf(out, "Material Site: %s\n", label)
	}
	if details.MaterialTypeID != "" {
		label := details.MaterialTypeID
		if details.MaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialTypeName, details.MaterialTypeID)
		}
		fmt.Fprintf(out, "Material Type: %s\n", label)
	}

	return nil
}
