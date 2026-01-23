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

type doFeaturesCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NameGeneric           string
	NameBranded           string
	Description           string
	ReleasedOn            string
	PDCAStage             string
	DifferentiationDegree string
	Scale                 string
}

func newDoFeaturesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new feature",
		Long: `Create a new feature.

Optional flags:
  --name-generic           Generic name
  --name-branded           Branded name
  --description            Feature description
  --released-on            Release date (YYYY-MM-DD)
  --pdca-stage             PDCA stage (plan/do/check/act)
  --differentiation-degree Differentiation degree
  --scale                  Feature scale`,
		Example: `  # Create a feature with names
  xbe do features create --name-generic "New Feature" --name-branded "XBE Feature"

  # Create a feature with all details
  xbe do features create --name-generic "Dashboard" --released-on 2024-01-15 --pdca-stage act`,
		RunE: runDoFeaturesCreate,
	}
	initDoFeaturesCreateFlags(cmd)
	return cmd
}

func init() {
	doFeaturesCmd.AddCommand(newDoFeaturesCreateCmd())
}

func initDoFeaturesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name-generic", "", "Generic name")
	cmd.Flags().String("name-branded", "", "Branded name")
	cmd.Flags().String("description", "", "Feature description")
	cmd.Flags().String("released-on", "", "Release date (YYYY-MM-DD)")
	cmd.Flags().String("pdca-stage", "", "PDCA stage (plan/do/check/act)")
	cmd.Flags().String("differentiation-degree", "", "Differentiation degree")
	cmd.Flags().String("scale", "", "Feature scale")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFeaturesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoFeaturesCreateOptions(cmd)
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

	if opts.NameGeneric != "" {
		attributes["name-generic"] = opts.NameGeneric
	}
	if opts.NameBranded != "" {
		attributes["name-branded"] = opts.NameBranded
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.ReleasedOn != "" {
		attributes["released-on"] = opts.ReleasedOn
	}
	if opts.PDCAStage != "" {
		attributes["pdca-stage"] = opts.PDCAStage
	}
	if opts.DifferentiationDegree != "" {
		attributes["differentiation-degree"] = opts.DifferentiationDegree
	}
	if opts.Scale != "" {
		attributes["scale"] = opts.Scale
	}

	data := map[string]any{
		"type":       "features",
		"attributes": attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/features", jsonBody)
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

	row := featureRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created feature %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoFeaturesCreateOptions(cmd *cobra.Command) (doFeaturesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	nameGeneric, _ := cmd.Flags().GetString("name-generic")
	nameBranded, _ := cmd.Flags().GetString("name-branded")
	description, _ := cmd.Flags().GetString("description")
	releasedOn, _ := cmd.Flags().GetString("released-on")
	pdcaStage, _ := cmd.Flags().GetString("pdca-stage")
	differentiationDegree, _ := cmd.Flags().GetString("differentiation-degree")
	scale, _ := cmd.Flags().GetString("scale")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFeaturesCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NameGeneric:           nameGeneric,
		NameBranded:           nameBranded,
		Description:           description,
		ReleasedOn:            releasedOn,
		PDCAStage:             pdcaStage,
		DifferentiationDegree: differentiationDegree,
		Scale:                 scale,
	}, nil
}

func featureRowFromSingle(resp jsonAPISingleResponse) featureRow {
	nameGeneric := strings.TrimSpace(stringAttr(resp.Data.Attributes, "name-generic"))
	nameBranded := strings.TrimSpace(stringAttr(resp.Data.Attributes, "name-branded"))
	name := firstNonEmpty(nameBranded, nameGeneric)

	return featureRow{
		ID:                    resp.Data.ID,
		Name:                  name,
		NameGeneric:           nameGeneric,
		NameBranded:           nameBranded,
		Released:              formatDate(stringAttr(resp.Data.Attributes, "released-on")),
		PDCAStage:             stringAttr(resp.Data.Attributes, "pdca-stage"),
		Scale:                 stringAttr(resp.Data.Attributes, "scale"),
		DifferentiationDegree: stringAttr(resp.Data.Attributes, "differentiation-degree"),
	}
}
