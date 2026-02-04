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

type materialTypeConversionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTypeConversionDetails struct {
	ID                        string `json:"id"`
	MaterialSupplierID        string `json:"material_supplier_id,omitempty"`
	MaterialSupplier          string `json:"material_supplier,omitempty"`
	MaterialSiteID            string `json:"material_site_id,omitempty"`
	MaterialSite              string `json:"material_site,omitempty"`
	MaterialTypeID            string `json:"material_type_id,omitempty"`
	MaterialType              string `json:"material_type,omitempty"`
	ForeignMaterialSupplierID string `json:"foreign_material_supplier_id,omitempty"`
	ForeignMaterialSupplier   string `json:"foreign_material_supplier,omitempty"`
	ForeignMaterialSiteID     string `json:"foreign_material_site_id,omitempty"`
	ForeignMaterialSite       string `json:"foreign_material_site,omitempty"`
	ForeignMaterialTypeID     string `json:"foreign_material_type_id,omitempty"`
	ForeignMaterialType       string `json:"foreign_material_type,omitempty"`
	CreatedAt                 string `json:"created_at,omitempty"`
	UpdatedAt                 string `json:"updated_at,omitempty"`
}

func newMaterialTypeConversionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material type conversion details",
		Long: `Show the full details of a material type conversion.

Output Fields:
  ID
  Material Supplier
  Material Site
  Material Type
  Foreign Material Supplier
  Foreign Material Site
  Foreign Material Type
  Created At
  Updated At

Arguments:
  <id>    The material type conversion ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show conversion details
  xbe view material-type-conversions show 123

  # Get JSON output
  xbe view material-type-conversions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTypeConversionsShow,
	}
	initMaterialTypeConversionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTypeConversionsCmd.AddCommand(newMaterialTypeConversionsShowCmd())
}

func initMaterialTypeConversionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeConversionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialTypeConversionsShowOptions(cmd)
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
		return fmt.Errorf("material type conversion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-type-conversions]", "created-at,updated-at,material-supplier,material-site,material-type,foreign-material-supplier,foreign-material-site,foreign-material-type")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("include", "material-supplier,material-site,material-type,foreign-material-supplier,foreign-material-site,foreign-material-type")

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-conversions/"+id, query)
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

	details := buildMaterialTypeConversionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTypeConversionDetails(cmd, details)
}

func parseMaterialTypeConversionsShowOptions(cmd *cobra.Command) (materialTypeConversionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeConversionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTypeConversionDetails(resp jsonAPISingleResponse) materialTypeConversionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	details := materialTypeConversionDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.MaterialSupplierID = relationshipIDFromMap(resource.Relationships, "material-supplier")
	details.MaterialSupplier = resolveMaterialSupplierName(details.MaterialSupplierID, included)
	details.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
	details.MaterialSite = resolveMaterialSiteName(details.MaterialSiteID, included)
	details.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
	details.MaterialType = resolveMaterialTypeName(details.MaterialTypeID, included)
	details.ForeignMaterialSupplierID = relationshipIDFromMap(resource.Relationships, "foreign-material-supplier")
	details.ForeignMaterialSupplier = resolveMaterialSupplierName(details.ForeignMaterialSupplierID, included)
	details.ForeignMaterialSiteID = relationshipIDFromMap(resource.Relationships, "foreign-material-site")
	details.ForeignMaterialSite = resolveMaterialSiteName(details.ForeignMaterialSiteID, included)
	details.ForeignMaterialTypeID = relationshipIDFromMap(resource.Relationships, "foreign-material-type")
	details.ForeignMaterialType = resolveMaterialTypeName(details.ForeignMaterialTypeID, included)

	return details
}

func renderMaterialTypeConversionDetails(cmd *cobra.Command, details materialTypeConversionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	if details.MaterialSupplierID != "" {
		label := details.MaterialSupplierID
		if details.MaterialSupplier != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSupplier, details.MaterialSupplierID)
		}
		fmt.Fprintf(out, "Material Supplier: %s\n", label)
	}
	if details.MaterialSiteID != "" {
		label := details.MaterialSiteID
		if details.MaterialSite != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSite, details.MaterialSiteID)
		}
		fmt.Fprintf(out, "Material Site: %s\n", label)
	}
	if details.MaterialTypeID != "" {
		label := details.MaterialTypeID
		if details.MaterialType != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialType, details.MaterialTypeID)
		}
		fmt.Fprintf(out, "Material Type: %s\n", label)
	}
	if details.ForeignMaterialSupplierID != "" {
		label := details.ForeignMaterialSupplierID
		if details.ForeignMaterialSupplier != "" {
			label = fmt.Sprintf("%s (%s)", details.ForeignMaterialSupplier, details.ForeignMaterialSupplierID)
		}
		fmt.Fprintf(out, "Foreign Material Supplier: %s\n", label)
	}
	if details.ForeignMaterialSiteID != "" {
		label := details.ForeignMaterialSiteID
		if details.ForeignMaterialSite != "" {
			label = fmt.Sprintf("%s (%s)", details.ForeignMaterialSite, details.ForeignMaterialSiteID)
		}
		fmt.Fprintf(out, "Foreign Material Site: %s\n", label)
	}
	if details.ForeignMaterialTypeID != "" {
		label := details.ForeignMaterialTypeID
		if details.ForeignMaterialType != "" {
			label = fmt.Sprintf("%s (%s)", details.ForeignMaterialType, details.ForeignMaterialTypeID)
		}
		fmt.Fprintf(out, "Foreign Material Type: %s\n", label)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
