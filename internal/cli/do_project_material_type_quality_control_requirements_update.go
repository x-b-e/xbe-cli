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

type doProjectMaterialTypeQualityControlRequirementsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	QualityControlClassification string
	Note                         string
}

func newDoProjectMaterialTypeQualityControlRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project material type quality control requirement",
		Long: `Update a project material type quality control requirement.

All flags are optional. Only provided flags will update the requirement.

Optional flags:
  --note  Optional note (set empty to clear)

Relationships:
  --quality-control-classification  Quality control classification ID`,
		Example: `  # Update the note
  xbe do project-material-type-quality-control-requirements update 123 --note \"Updated note\"

  # Update the quality control classification
  xbe do project-material-type-quality-control-requirements update 123 --quality-control-classification 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectMaterialTypeQualityControlRequirementsUpdate,
	}
	initDoProjectMaterialTypeQualityControlRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectMaterialTypeQualityControlRequirementsCmd.AddCommand(newDoProjectMaterialTypeQualityControlRequirementsUpdateCmd())
}

func initDoProjectMaterialTypeQualityControlRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quality-control-classification", "", "Quality control classification ID")
	cmd.Flags().String("note", "", "Optional note (empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectMaterialTypeQualityControlRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectMaterialTypeQualityControlRequirementsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	if cmd.Flags().Changed("quality-control-classification") {
		if strings.TrimSpace(opts.QualityControlClassification) == "" {
			err := fmt.Errorf("--quality-control-classification cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["quality-control-classification"] = map[string]any{
			"data": map[string]any{
				"type": "quality-control-classifications",
				"id":   opts.QualityControlClassification,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-material-type-quality-control-requirements",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-material-type-quality-control-requirements/"+opts.ID, jsonBody)
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

	row := projectMaterialTypeQualityControlRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project material type quality control requirement %s\n", row.ID)
	return nil
}

func parseDoProjectMaterialTypeQualityControlRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doProjectMaterialTypeQualityControlRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	qualityControlClassification, _ := cmd.Flags().GetString("quality-control-classification")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectMaterialTypeQualityControlRequirementsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		QualityControlClassification: qualityControlClassification,
		Note:                         note,
	}, nil
}
