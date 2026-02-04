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

type projectMaterialTypeQualityControlRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectMaterialTypeQualityControlRequirementDetails struct {
	ID                               string `json:"id"`
	ProjectMaterialTypeID            string `json:"project_material_type_id,omitempty"`
	ProjectMaterialTypeName          string `json:"project_material_type_name,omitempty"`
	QualityControlClassificationID   string `json:"quality_control_classification_id,omitempty"`
	QualityControlClassificationName string `json:"quality_control_classification_name,omitempty"`
	Note                             string `json:"note,omitempty"`
}

func newProjectMaterialTypeQualityControlRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project material type quality control requirement details",
		Long: `Show the full details of a project material type quality control requirement.

Includes the associated project material type and quality control classification.

Arguments:
  <id>  The requirement ID (required).`,
		Example: `  # Show a requirement
  xbe view project-material-type-quality-control-requirements show 123

  # Output as JSON
  xbe view project-material-type-quality-control-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectMaterialTypeQualityControlRequirementsShow,
	}
	initProjectMaterialTypeQualityControlRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	projectMaterialTypeQualityControlRequirementsCmd.AddCommand(newProjectMaterialTypeQualityControlRequirementsShowCmd())
}

func initProjectMaterialTypeQualityControlRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectMaterialTypeQualityControlRequirementsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectMaterialTypeQualityControlRequirementsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project material type quality control requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-material-type-quality-control-requirements]", "project-material-type,quality-control-classification,note")
	query.Set("include", "project-material-type,quality-control-classification")
	query.Set("fields[project-material-types]", "display-name,explicit-display-name")
	query.Set("fields[quality-control-classifications]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/project-material-type-quality-control-requirements/"+id, query)
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

	details := buildProjectMaterialTypeQualityControlRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectMaterialTypeQualityControlRequirementDetails(cmd, details)
}

func parseProjectMaterialTypeQualityControlRequirementsShowOptions(cmd *cobra.Command) (projectMaterialTypeQualityControlRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectMaterialTypeQualityControlRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectMaterialTypeQualityControlRequirementDetails(resp jsonAPISingleResponse) projectMaterialTypeQualityControlRequirementDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectMaterialTypeQualityControlRequirementDetails{
		ID:   resp.Data.ID,
		Note: stringAttr(resp.Data.Attributes, "note"),
	}

	if rel, ok := resp.Data.Relationships["project-material-type"]; ok && rel.Data != nil {
		details.ProjectMaterialTypeID = rel.Data.ID
		if projectMaterialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectMaterialTypeName = projectMaterialTypeLabel(projectMaterialType.Attributes)
		}
	}

	if rel, ok := resp.Data.Relationships["quality-control-classification"]; ok && rel.Data != nil {
		details.QualityControlClassificationID = rel.Data.ID
		if classification, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.QualityControlClassificationName = stringAttr(classification.Attributes, "name")
		}
	}

	return details
}

func renderProjectMaterialTypeQualityControlRequirementDetails(cmd *cobra.Command, details projectMaterialTypeQualityControlRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectMaterialTypeID != "" {
		label := details.ProjectMaterialTypeID
		if details.ProjectMaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.ProjectMaterialTypeName, details.ProjectMaterialTypeID)
		}
		fmt.Fprintf(out, "Project Material Type: %s\n", label)
	}
	if details.QualityControlClassificationID != "" {
		label := details.QualityControlClassificationID
		if details.QualityControlClassificationName != "" {
			label = fmt.Sprintf("%s (%s)", details.QualityControlClassificationName, details.QualityControlClassificationID)
		}
		fmt.Fprintf(out, "Quality Control Classification: %s\n", label)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}

	return nil
}
