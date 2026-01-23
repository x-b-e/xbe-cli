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

type doJobProductionPlanMaterialTypeQualityControlRequirementsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	QualityControlClassification string
	Note                         string
}

func newDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan material type quality control requirement",
		Long: `Update a job production plan material type quality control requirement.

Optional:
  --quality-control-classification  Quality control classification ID
  --note                            Requirement note`,
		Example: `  # Update the note
  xbe do job-production-plan-material-type-quality-control-requirements update 123 --note "Updated note"

  # Change the quality control classification
  xbe do job-production-plan-material-type-quality-control-requirements update 123 --quality-control-classification 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdate,
	}
	initDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTypeQualityControlRequirementsCmd.AddCommand(newDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateCmd())
}

func initDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quality-control-classification", "", "Quality control classification ID")
	cmd.Flags().String("note", "", "Requirement note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateOptions(cmd, args)
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

	data := map[string]any{
		"type": "job-production-plan-material-type-quality-control-requirements",
		"id":   opts.ID,
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("quality-control-classification") {
		relationships["quality-control-classification"] = map[string]any{
			"data": map[string]any{
				"type": "quality-control-classifications",
				"id":   opts.QualityControlClassification,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-material-type-quality-control-requirements/"+opts.ID, jsonBody)
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
		row := jobProductionPlanMaterialTypeQualityControlRequirementRow{ID: resp.Data.ID}
		if rel, ok := resp.Data.Relationships["job-production-plan-material-type"]; ok && rel.Data != nil {
			row.JobProductionPlanMaterialType = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["quality-control-classification"]; ok && rel.Data != nil {
			row.QualityControlClassificationID = rel.Data.ID
		}
		row.Note = stringAttr(resp.Data.Attributes, "note")
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan material type quality control requirement %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanMaterialTypeQualityControlRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanMaterialTypeQualityControlRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	qualityControlClassification, _ := cmd.Flags().GetString("quality-control-classification")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialTypeQualityControlRequirementsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		QualityControlClassification: qualityControlClassification,
		Note:                         note,
	}, nil
}
