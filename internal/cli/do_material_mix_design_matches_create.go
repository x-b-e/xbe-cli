package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialMixDesignMatchesCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	MaterialType  string
	AsOf          string
	MaterialSites string
}

type materialMixDesignMatchDetail struct {
	ID         string `json:"id"`
	Precedence int    `json:"precedence,omitempty"`
}

type materialMixDesignMatchRow struct {
	ID                 string                         `json:"id"`
	AsOf               string                         `json:"as_of,omitempty"`
	MaterialTypeID     string                         `json:"material_type_id,omitempty"`
	MaterialSiteIDs    []string                       `json:"material_site_ids,omitempty"`
	MaterialMixDesigns []materialMixDesignMatchDetail `json:"material_mix_designs,omitempty"`
}

func newDoMaterialMixDesignMatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Match material mix designs",
		Long: `Match material mix designs.

Required:
  --material-type  Material type ID
  --as-of          Match as of timestamp (RFC3339)

Optional:
  --material-sites  Material site IDs (comma-separated)`,
		Example: `  # Match material mix designs for a material type
  xbe do material-mix-design-matches create --material-type 123 --as-of "2026-01-23T00:00:00Z"

  # Match with material sites
  xbe do material-mix-design-matches create \
    --material-type 123 \
    --as-of "2026-01-23T00:00:00Z" \
    --material-sites 456,789

  # JSON output
  xbe do material-mix-design-matches create --material-type 123 --as-of "2026-01-23T00:00:00Z" --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialMixDesignMatchesCreate,
	}
	initDoMaterialMixDesignMatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialMixDesignMatchesCmd.AddCommand(newDoMaterialMixDesignMatchesCreateCmd())
}

func initDoMaterialMixDesignMatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("as-of", "", "Match as of timestamp (RFC3339)")
	cmd.Flags().String("material-sites", "", "Material site IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialMixDesignMatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialMixDesignMatchesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialType) == "" {
		err := fmt.Errorf("--material-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.AsOf) == "" {
		err := fmt.Errorf("--as-of is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"as-of": opts.AsOf,
	}

	relationships := map[string]any{
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
	}

	materialSiteIDs := splitCommaList(opts.MaterialSites)
	if len(materialSiteIDs) > 0 {
		data := make([]map[string]any, len(materialSiteIDs))
		for i, id := range materialSiteIDs {
			data[i] = map[string]any{
				"type": "material-sites",
				"id":   id,
			}
		}
		relationships["material-sites"] = map[string]any{"data": data}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-mix-design-matches",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-mix-design-matches", jsonBody)
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

	row := buildMaterialMixDesignMatchRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderMaterialMixDesignMatch(cmd, row)
}

func parseDoMaterialMixDesignMatchesCreateOptions(cmd *cobra.Command) (doMaterialMixDesignMatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	asOf, _ := cmd.Flags().GetString("as-of")
	materialSites, _ := cmd.Flags().GetString("material-sites")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialMixDesignMatchesCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		MaterialType:  materialType,
		AsOf:          asOf,
		MaterialSites: materialSites,
	}, nil
}

func buildMaterialMixDesignMatchRow(resp jsonAPISingleResponse) materialMixDesignMatchRow {
	resource := resp.Data
	attrs := resource.Attributes

	row := materialMixDesignMatchRow{
		ID:                 resource.ID,
		AsOf:               formatDateTime(stringAttr(attrs, "as-of")),
		MaterialMixDesigns: parseMaterialMixDesignMatchDetails(attrs),
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-sites"]; ok && rel.raw != nil {
		row.MaterialSiteIDs = relationshipIDStrings(rel)
	}

	if len(row.MaterialMixDesigns) == 0 {
		if rel, ok := resource.Relationships["material-mix-designs"]; ok && rel.raw != nil {
			ids := relationshipIDStrings(rel)
			for _, id := range ids {
				row.MaterialMixDesigns = append(row.MaterialMixDesigns, materialMixDesignMatchDetail{ID: id})
			}
		}
	}

	return row
}

func parseMaterialMixDesignMatchDetails(attrs map[string]any) []materialMixDesignMatchDetail {
	if attrs == nil {
		return nil
	}
	value, ok := attrs["material-mix-designs-details"]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []map[string]any:
		details := make([]materialMixDesignMatchDetail, 0, len(typed))
		for _, item := range typed {
			id := stringAttr(item, "id")
			if id == "" {
				continue
			}
			details = append(details, materialMixDesignMatchDetail{
				ID:         id,
				Precedence: intAttr(item, "precedence"),
			})
		}
		return details
	case []any:
		details := make([]materialMixDesignMatchDetail, 0, len(typed))
		for _, item := range typed {
			mapped, ok := item.(map[string]any)
			if !ok || mapped == nil {
				continue
			}
			id := stringAttr(mapped, "id")
			if id == "" {
				continue
			}
			details = append(details, materialMixDesignMatchDetail{
				ID:         id,
				Precedence: intAttr(mapped, "precedence"),
			})
		}
		return details
	default:
		return nil
	}
}

func renderMaterialMixDesignMatch(cmd *cobra.Command, row materialMixDesignMatchRow) error {
	out := cmd.OutOrStdout()

	if row.ID != "" {
		fmt.Fprintf(out, "Created material mix design match %s\n", row.ID)
	} else {
		fmt.Fprintln(out, "Created material mix design match")
	}

	if row.AsOf != "" {
		fmt.Fprintf(out, "As Of: %s\n", row.AsOf)
	}
	if row.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type: %s\n", row.MaterialTypeID)
	}
	if len(row.MaterialSiteIDs) > 0 {
		fmt.Fprintf(out, "Material Sites: %s\n", strings.Join(row.MaterialSiteIDs, ", "))
	}

	if len(row.MaterialMixDesigns) == 0 {
		fmt.Fprintln(out, "No material mix designs matched.")
		return nil
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Matched Material Mix Designs:")
	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPRECEDENCE")
	for _, detail := range row.MaterialMixDesigns {
		fmt.Fprintf(writer, "%s\t%d\n", detail.ID, detail.Precedence)
	}

	return writer.Flush()
}
