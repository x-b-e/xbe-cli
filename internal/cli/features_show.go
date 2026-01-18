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

type featuresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type featureDetails struct {
	ID                    string `json:"id"`
	NameGeneric           string `json:"name_generic"`
	NameBranded           string `json:"name_branded"`
	Description           string `json:"description"`
	Released              string `json:"released"`
	PDCAStage             string `json:"pdca_stage"`
	Scale                 string `json:"scale"`
	DifferentiationDegree string `json:"differentiation_degree"`
}

func newFeaturesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show feature details",
		Long: `Show the full details of a specific feature.

Retrieves and displays comprehensive information about a feature including
its full description, release date, and categorization.

Output Fields (table format):
  ID                     Unique feature identifier
  Name (Generic)         Generic feature name
  Name (Branded)         Branded feature name
  Description            Full feature description
  Released               Release date
  PDCA Stage             Plan/Do/Check/Act stage
  Scale                  Feature scale
  Differentiation Degree Differentiation level

Arguments:
  <id>          The feature ID (required). You can find IDs using the list command.`,
		Example: `  # View a feature by ID
  xbe view features show 123

  # Get feature as JSON
  xbe view features show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runFeaturesShow,
	}
	initFeaturesShowFlags(cmd)
	return cmd
}

func init() {
	featuresCmd.AddCommand(newFeaturesShowCmd())
}

func initFeaturesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFeaturesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseFeaturesShowOptions(cmd)
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
		return fmt.Errorf("feature id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[features]", "name-generic,name-branded,description,released-on,pdca-stage,scale,differentiation-degree")

	body, _, err := client.Get(cmd.Context(), "/v1/features/"+id, query)
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

	details := buildFeatureDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderFeatureDetails(cmd, details)
}

func parseFeaturesShowOptions(cmd *cobra.Command) (featuresShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return featuresShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return featuresShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return featuresShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return featuresShowOptions{}, err
	}

	return featuresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildFeatureDetails(resp jsonAPISingleResponse) featureDetails {
	attrs := resp.Data.Attributes

	return featureDetails{
		ID:                    resp.Data.ID,
		NameGeneric:           strings.TrimSpace(stringAttr(attrs, "name-generic")),
		NameBranded:           strings.TrimSpace(stringAttr(attrs, "name-branded")),
		Description:           strings.TrimSpace(stringAttr(attrs, "description")),
		Released:              formatDate(stringAttr(attrs, "released-on")),
		PDCAStage:             stringAttr(attrs, "pdca-stage"),
		Scale:                 stringAttr(attrs, "scale"),
		DifferentiationDegree: stringAttr(attrs, "differentiation-degree"),
	}
}

func renderFeatureDetails(cmd *cobra.Command, details featureDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.NameGeneric != "" {
		fmt.Fprintf(out, "Name (Generic): %s\n", details.NameGeneric)
	}
	if details.NameBranded != "" {
		fmt.Fprintf(out, "Name (Branded): %s\n", details.NameBranded)
	}
	if details.Released != "" {
		fmt.Fprintf(out, "Released: %s\n", details.Released)
	}
	if details.PDCAStage != "" {
		fmt.Fprintf(out, "PDCA Stage: %s\n", details.PDCAStage)
	}
	if details.Scale != "" {
		fmt.Fprintf(out, "Scale: %s\n", details.Scale)
	}
	if details.DifferentiationDegree != "" {
		fmt.Fprintf(out, "Differentiation Degree: %s\n", details.DifferentiationDegree)
	}
	if details.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Description)
	}

	return nil
}
