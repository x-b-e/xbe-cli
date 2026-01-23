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

type doFeaturesUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	NameGeneric           string
	NameBranded           string
	Description           string
	ReleasedOn            string
	PDCAStage             string
	DifferentiationDegree string
	Scale                 string
}

func newDoFeaturesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a feature",
		Long: `Update an existing feature.

Optional flags:
  --name-generic           Generic name
  --name-branded           Branded name
  --description            Feature description
  --released-on            Release date (YYYY-MM-DD)
  --pdca-stage             PDCA stage (plan/do/check/act)
  --differentiation-degree Differentiation degree
  --scale                  Feature scale`,
		Example: `  # Update feature name
  xbe do features update 123 --name-generic "Updated Feature"

  # Update PDCA stage
  xbe do features update 123 --pdca-stage act`,
		Args: cobra.ExactArgs(1),
		RunE: runDoFeaturesUpdate,
	}
	initDoFeaturesUpdateFlags(cmd)
	return cmd
}

func init() {
	doFeaturesCmd.AddCommand(newDoFeaturesUpdateCmd())
}

func initDoFeaturesUpdateFlags(cmd *cobra.Command) {
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

func runDoFeaturesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoFeaturesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name-generic") {
		attributes["name-generic"] = opts.NameGeneric
	}
	if cmd.Flags().Changed("name-branded") {
		attributes["name-branded"] = opts.NameBranded
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("released-on") {
		attributes["released-on"] = opts.ReleasedOn
	}
	if cmd.Flags().Changed("pdca-stage") {
		attributes["pdca-stage"] = opts.PDCAStage
	}
	if cmd.Flags().Changed("differentiation-degree") {
		attributes["differentiation-degree"] = opts.DifferentiationDegree
	}
	if cmd.Flags().Changed("scale") {
		attributes["scale"] = opts.Scale
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "features",
		"id":         opts.ID,
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

	path := fmt.Sprintf("/v1/features/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated feature %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoFeaturesUpdateOptions(cmd *cobra.Command, args []string) (doFeaturesUpdateOptions, error) {
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

	return doFeaturesUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		NameGeneric:           nameGeneric,
		NameBranded:           nameBranded,
		Description:           description,
		ReleasedOn:            releasedOn,
		PDCAStage:             pdcaStage,
		DifferentiationDegree: differentiationDegree,
		Scale:                 scale,
	}, nil
}
