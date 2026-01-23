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

type doJobProductionPlanMaterialTypeQualityControlRequirementsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	JobProductionPlanMaterialType string
	QualityControlClassification  string
	Note                          string
}

func newDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan material type quality control requirement",
		Long: `Create a job production plan material type quality control requirement.

Required:
  --job-production-plan-material-type  Job production plan material type ID
  --quality-control-classification     Quality control classification ID

Optional:
  --note                               Requirement note`,
		Example: `  # Create a requirement
  xbe do job-production-plan-material-type-quality-control-requirements create \
    --job-production-plan-material-type 123 \
    --quality-control-classification 456

  # Create with a note
  xbe do job-production-plan-material-type-quality-control-requirements create \
    --job-production-plan-material-type 123 \
    --quality-control-classification 456 \
    --note "Temperature check"`,
		RunE: runDoJobProductionPlanMaterialTypeQualityControlRequirementsCreate,
	}
	initDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTypeQualityControlRequirementsCmd.AddCommand(newDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateCmd())
}

func initDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-material-type", "", "Job production plan material type ID")
	cmd.Flags().String("quality-control-classification", "", "Quality control classification ID")
	cmd.Flags().String("note", "", "Requirement note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan-material-type")
	_ = cmd.MarkFlagRequired("quality-control-classification")
}

func runDoJobProductionPlanMaterialTypeQualityControlRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateOptions(cmd)
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
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"job-production-plan-material-type": map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-material-types",
				"id":   opts.JobProductionPlanMaterialType,
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
		"type":          "job-production-plan-material-type-quality-control-requirements",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-material-type-quality-control-requirements", jsonBody)
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

	if opts.JSON {
		row := jobProductionPlanMaterialTypeQualityControlRequirementRow{
			ID: resp.Data.ID,
		}
		if rel, ok := resp.Data.Relationships["job-production-plan-material-type"]; ok && rel.Data != nil {
			row.JobProductionPlanMaterialType = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["quality-control-classification"]; ok && rel.Data != nil {
			row.QualityControlClassificationID = rel.Data.ID
		}
		row.Note = stringAttr(resp.Data.Attributes, "note")
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan material type quality control requirement %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanMaterialTypeQualityControlRequirementsCreateOptions(cmd *cobra.Command) (doJobProductionPlanMaterialTypeQualityControlRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanMaterialType, _ := cmd.Flags().GetString("job-production-plan-material-type")
	qualityControlClassification, _ := cmd.Flags().GetString("quality-control-classification")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialTypeQualityControlRequirementsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		JobProductionPlanMaterialType: jobProductionPlanMaterialType,
		QualityControlClassification:  qualityControlClassification,
		Note:                          note,
	}, nil
}
