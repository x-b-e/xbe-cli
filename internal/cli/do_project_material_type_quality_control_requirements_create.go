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

type doProjectMaterialTypeQualityControlRequirementsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ProjectMaterialType          string
	QualityControlClassification string
	Note                         string
}

func newDoProjectMaterialTypeQualityControlRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project material type quality control requirement",
		Long: `Create a project material type quality control requirement.

Required flags:
  --project-material-type          Project material type ID (required)
  --quality-control-classification Quality control classification ID (required)

Optional flags:
  --note  Optional note`,
		Example: `  # Create a requirement
  xbe do project-material-type-quality-control-requirements create \\
    --project-material-type 123 \\
    --quality-control-classification 456

  # Create with a note
  xbe do project-material-type-quality-control-requirements create \\
    --project-material-type 123 \\
    --quality-control-classification 456 \\
    --note \"Daily temperature check\"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectMaterialTypeQualityControlRequirementsCreate,
	}
	initDoProjectMaterialTypeQualityControlRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectMaterialTypeQualityControlRequirementsCmd.AddCommand(newDoProjectMaterialTypeQualityControlRequirementsCreateCmd())
}

func initDoProjectMaterialTypeQualityControlRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-material-type", "", "Project material type ID (required)")
	cmd.Flags().String("quality-control-classification", "", "Quality control classification ID (required)")
	cmd.Flags().String("note", "", "Optional note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-material-type")
	_ = cmd.MarkFlagRequired("quality-control-classification")
}

func runDoProjectMaterialTypeQualityControlRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectMaterialTypeQualityControlRequirementsCreateOptions(cmd)
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

	relationships := map[string]any{
		"project-material-type": map[string]any{
			"data": map[string]any{
				"type": "project-material-types",
				"id":   opts.ProjectMaterialType,
			},
		},
		"quality-control-classification": map[string]any{
			"data": map[string]any{
				"type": "quality-control-classifications",
				"id":   opts.QualityControlClassification,
			},
		},
	}

	data := map[string]any{
		"type":          "project-material-type-quality-control-requirements",
		"relationships": relationships,
	}

	if cmd.Flags().Changed("note") {
		data["attributes"] = map[string]any{
			"note": opts.Note,
		}
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-material-type-quality-control-requirements", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project material type quality control requirement %s\n", row.ID)
	return nil
}

func parseDoProjectMaterialTypeQualityControlRequirementsCreateOptions(cmd *cobra.Command) (doProjectMaterialTypeQualityControlRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectMaterialType, _ := cmd.Flags().GetString("project-material-type")
	qualityControlClassification, _ := cmd.Flags().GetString("quality-control-classification")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectMaterialTypeQualityControlRequirementsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ProjectMaterialType:          projectMaterialType,
		QualityControlClassification: qualityControlClassification,
		Note:                         note,
	}, nil
}

func projectMaterialTypeQualityControlRequirementRowFromSingle(resp jsonAPISingleResponse) projectMaterialTypeQualityControlRequirementRow {
	row := projectMaterialTypeQualityControlRequirementRow{
		ID:   resp.Data.ID,
		Note: stringAttr(resp.Data.Attributes, "note"),
	}

	if rel, ok := resp.Data.Relationships["project-material-type"]; ok && rel.Data != nil {
		row.ProjectMaterialTypeID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["quality-control-classification"]; ok && rel.Data != nil {
		row.QualityControlClassificationID = rel.Data.ID
	}

	return row
}
