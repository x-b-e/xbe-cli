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

type materialSiteMeasuresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteMeasureDetails struct {
	ID                     string   `json:"id"`
	Slug                   string   `json:"slug"`
	Name                   string   `json:"name"`
	ValidReadingValueMin   string   `json:"valid_reading_value_min,omitempty"`
	ValidReadingValueMax   string   `json:"valid_reading_value_max,omitempty"`
	MaterialSiteReadingIDs []string `json:"material_site_reading_ids,omitempty"`
}

func newMaterialSiteMeasuresShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site measure details",
		Long: `Show the full details of a material site measure.

Material site measures define the measurement types used for
material site readings and the valid value ranges.

Output Fields:
  ID        Material site measure identifier
  Slug      URL-friendly identifier
  Name      Measure name
  Min       Minimum valid reading value
  Max       Maximum valid reading value
  Readings  Material site reading IDs

Arguments:
  <id>    The material site measure ID (required). You can find IDs using the list command.`,
		Example: `  # Show a material site measure
  xbe view material-site-measures show 123

  # Get JSON output
  xbe view material-site-measures show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteMeasuresShow,
	}
	initMaterialSiteMeasuresShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteMeasuresCmd.AddCommand(newMaterialSiteMeasuresShowCmd())
}

func initMaterialSiteMeasuresShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteMeasuresShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSiteMeasuresShowOptions(cmd)
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
		return fmt.Errorf("material site measure id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-measures]", "slug,name,valid-reading-value-min,valid-reading-value-max,material-site-readings")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-measures/"+id, query)
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

	details := buildMaterialSiteMeasureDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteMeasureDetails(cmd, details)
}

func parseMaterialSiteMeasuresShowOptions(cmd *cobra.Command) (materialSiteMeasuresShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteMeasuresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteMeasureDetails(resp jsonAPISingleResponse) materialSiteMeasureDetails {
	attrs := resp.Data.Attributes
	details := materialSiteMeasureDetails{
		ID:                   resp.Data.ID,
		Slug:                 stringAttr(attrs, "slug"),
		Name:                 stringAttr(attrs, "name"),
		ValidReadingValueMin: stringAttr(attrs, "valid-reading-value-min"),
		ValidReadingValueMax: stringAttr(attrs, "valid-reading-value-max"),
	}

	if rel, ok := resp.Data.Relationships["material-site-readings"]; ok {
		details.MaterialSiteReadingIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderMaterialSiteMeasureDetails(cmd *cobra.Command, details materialSiteMeasureDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Slug != "" {
		fmt.Fprintf(out, "Slug: %s\n", details.Slug)
	}
	if details.ValidReadingValueMin != "" {
		fmt.Fprintf(out, "Valid Reading Min: %s\n", details.ValidReadingValueMin)
	}
	if details.ValidReadingValueMax != "" {
		fmt.Fprintf(out, "Valid Reading Max: %s\n", details.ValidReadingValueMax)
	}
	if len(details.MaterialSiteReadingIDs) > 0 {
		fmt.Fprintf(out, "Material Site Reading IDs: %s\n", strings.Join(details.MaterialSiteReadingIDs, ", "))
	}

	return nil
}
