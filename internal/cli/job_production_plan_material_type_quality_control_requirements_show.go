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

type jobProductionPlanMaterialTypeQualityControlRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanMaterialTypeQualityControlRequirementDetails struct {
	ID                             string `json:"id"`
	JobProductionPlanMaterialType  string `json:"job_production_plan_material_type_id,omitempty"`
	QualityControlClassificationID string `json:"quality_control_classification_id,omitempty"`
	Note                           string `json:"note,omitempty"`
}

func newJobProductionPlanMaterialTypeQualityControlRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan material type quality control requirement details",
		Long: `Show the full details of a job production plan material type quality control requirement.

Output Fields:
  ID               Requirement identifier
  Material Type    Job production plan material type ID
  QC Class         Quality control classification ID
  Note             Requirement note

Arguments:
  <id>    The requirement ID (required). You can find IDs using the list command.`,
		Example: `  # Show a requirement
  xbe view job-production-plan-material-type-quality-control-requirements show 123

  # Get JSON output
  xbe view job-production-plan-material-type-quality-control-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanMaterialTypeQualityControlRequirementsShow,
	}
	initJobProductionPlanMaterialTypeQualityControlRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialTypeQualityControlRequirementsCmd.AddCommand(newJobProductionPlanMaterialTypeQualityControlRequirementsShowCmd())
}

func initJobProductionPlanMaterialTypeQualityControlRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialTypeQualityControlRequirementsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanMaterialTypeQualityControlRequirementsShowOptions(cmd)
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
		return fmt.Errorf("job production plan material type quality control requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "job-production-plan-material-type,quality-control-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-type-quality-control-requirements/"+id, query)
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

	details := buildJobProductionPlanMaterialTypeQualityControlRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanMaterialTypeQualityControlRequirementDetails(cmd, details)
}

func parseJobProductionPlanMaterialTypeQualityControlRequirementsShowOptions(cmd *cobra.Command) (jobProductionPlanMaterialTypeQualityControlRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialTypeQualityControlRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanMaterialTypeQualityControlRequirementDetails(resp jsonAPISingleResponse) jobProductionPlanMaterialTypeQualityControlRequirementDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanMaterialTypeQualityControlRequirementDetails{
		ID:   resource.ID,
		Note: stringAttr(attrs, "note"),
	}

	if rel, ok := resource.Relationships["job-production-plan-material-type"]; ok && rel.Data != nil {
		details.JobProductionPlanMaterialType = rel.Data.ID
	}
	if rel, ok := resource.Relationships["quality-control-classification"]; ok && rel.Data != nil {
		details.QualityControlClassificationID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanMaterialTypeQualityControlRequirementDetails(cmd *cobra.Command, details jobProductionPlanMaterialTypeQualityControlRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanMaterialType != "" {
		fmt.Fprintf(out, "Material Type: %s\n", details.JobProductionPlanMaterialType)
	}
	if details.QualityControlClassificationID != "" {
		fmt.Fprintf(out, "QC Class: %s\n", details.QualityControlClassificationID)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}

	return nil
}
