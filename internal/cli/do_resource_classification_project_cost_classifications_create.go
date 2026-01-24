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

type doResourceClassificationProjectCostClassificationsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ResourceClassificationType string
	ResourceClassificationID   string
	ProjectCostClassification  string
	Broker                     string
}

func newDoResourceClassificationProjectCostClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource classification project cost classification",
		Long: `Create a resource classification project cost classification.

Required flags:
  --resource-classification-type  Resource classification type (LaborClassification, EquipmentClassification)
  --resource-classification-id    Resource classification ID
  --project-cost-classification   Project cost classification ID
  --broker                        Broker ID

Notes:
  - Resource classification must be LaborClassification or EquipmentClassification.
  - The project cost classification must belong to the same broker.`,
		Example: `  # Link a labor classification to a project cost classification
  xbe do resource-classification-project-cost-classifications create \
    --resource-classification-type LaborClassification \
    --resource-classification-id 456 \
    --project-cost-classification 789 \
    --broker 123

  # Output as JSON
  xbe do resource-classification-project-cost-classifications create \
    --resource-classification-type LaborClassification \
    --resource-classification-id 456 \
    --project-cost-classification 789 \
    --broker 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoResourceClassificationProjectCostClassificationsCreate,
	}
	initDoResourceClassificationProjectCostClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doResourceClassificationProjectCostClassificationsCmd.AddCommand(newDoResourceClassificationProjectCostClassificationsCreateCmd())
}

func initDoResourceClassificationProjectCostClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("resource-classification-type", "", "Resource classification type (LaborClassification, EquipmentClassification) (required)")
	cmd.Flags().String("resource-classification-id", "", "Resource classification ID (required)")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoResourceClassificationProjectCostClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoResourceClassificationProjectCostClassificationsCreateOptions(cmd)
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

	resourceClassificationType := strings.TrimSpace(opts.ResourceClassificationType)
	if resourceClassificationType == "" {
		err := fmt.Errorf("--resource-classification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	resourceClassificationID := strings.TrimSpace(opts.ResourceClassificationID)
	if resourceClassificationID == "" {
		err := fmt.Errorf("--resource-classification-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	projectCostClassification := strings.TrimSpace(opts.ProjectCostClassification)
	if projectCostClassification == "" {
		err := fmt.Errorf("--project-cost-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	broker := strings.TrimSpace(opts.Broker)
	if broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	resourceClassificationTypeMapped, err := parseCrewRateResourceClassificationType(resourceClassificationType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"resource-classification": map[string]any{
			"data": map[string]any{
				"type": resourceClassificationTypeMapped,
				"id":   resourceClassificationID,
			},
		},
		"project-cost-classification": map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   projectCostClassification,
			},
		},
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "resource-classification-project-cost-classifications",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/resource-classification-project-cost-classifications", jsonBody)
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

	details := buildResourceClassificationProjectCostClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created resource classification project cost classification %s\n", details.ID)
	return nil
}

func parseDoResourceClassificationProjectCostClassificationsCreateOptions(cmd *cobra.Command) (doResourceClassificationProjectCostClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doResourceClassificationProjectCostClassificationsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ResourceClassificationType: resourceClassificationType,
		ResourceClassificationID:   resourceClassificationID,
		ProjectCostClassification:  projectCostClassification,
		Broker:                     broker,
	}, nil
}
