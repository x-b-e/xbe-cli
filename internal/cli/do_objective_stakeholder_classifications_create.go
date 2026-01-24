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

type doObjectiveStakeholderClassificationsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	Objective                 string
	StakeholderClassification string
	InterestDegree            float64
}

func newDoObjectiveStakeholderClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an objective stakeholder classification",
		Long: `Create an objective stakeholder classification.

Required flags:
  --objective                 Objective ID (required; must be a template)
  --stakeholder-classification Stakeholder classification ID (required)
  --interest-degree           Interest degree between 0 and 1 (required)

Note: Only admin users can create objective stakeholder classifications.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an objective stakeholder classification
  xbe do objective-stakeholder-classifications create \
    --objective 123 \
    --stakeholder-classification 456 \
    --interest-degree 0.75

  # Get JSON output
  xbe do objective-stakeholder-classifications create \
    --objective 123 \
    --stakeholder-classification 456 \
    --interest-degree 0.6 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoObjectiveStakeholderClassificationsCreate,
	}
	initDoObjectiveStakeholderClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doObjectiveStakeholderClassificationsCmd.AddCommand(newDoObjectiveStakeholderClassificationsCreateCmd())
}

func initDoObjectiveStakeholderClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("objective", "", "Objective ID (required)")
	cmd.Flags().String("stakeholder-classification", "", "Stakeholder classification ID (required)")
	cmd.Flags().Float64("interest-degree", 0, "Interest degree between 0 and 1 (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectiveStakeholderClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoObjectiveStakeholderClassificationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.Objective) == "" {
		err := fmt.Errorf("--objective is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.StakeholderClassification) == "" {
		err := fmt.Errorf("--stakeholder-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("interest-degree") {
		err := fmt.Errorf("--interest-degree is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.InterestDegree <= 0 || opts.InterestDegree > 1 {
		err := fmt.Errorf("--interest-degree must be between 0 and 1")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"interest-degree": opts.InterestDegree,
	}

	relationships := map[string]any{
		"objective": map[string]any{
			"data": map[string]any{
				"type": "objectives",
				"id":   opts.Objective,
			},
		},
		"stakeholder-classification": map[string]any{
			"data": map[string]any{
				"type": "stakeholder-classifications",
				"id":   opts.StakeholderClassification,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "objective-stakeholder-classifications",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/objective-stakeholder-classifications", jsonBody)
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

	row := buildObjectiveStakeholderClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created objective stakeholder classification %s\n", row.ID)
	return nil
}

func parseDoObjectiveStakeholderClassificationsCreateOptions(cmd *cobra.Command) (doObjectiveStakeholderClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	objective, _ := cmd.Flags().GetString("objective")
	stakeholderClassification, _ := cmd.Flags().GetString("stakeholder-classification")
	interestDegree, _ := cmd.Flags().GetFloat64("interest-degree")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectiveStakeholderClassificationsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		Objective:                 objective,
		StakeholderClassification: stakeholderClassification,
		InterestDegree:            interestDegree,
	}, nil
}

func buildObjectiveStakeholderClassificationRowFromSingle(resp jsonAPISingleResponse) objectiveStakeholderClassificationRow {
	resource := resp.Data
	attrs := resource.Attributes

	return objectiveStakeholderClassificationRow{
		ID:                          resource.ID,
		ObjectiveID:                 relationshipIDFromMap(resource.Relationships, "objective"),
		StakeholderClassificationID: relationshipIDFromMap(resource.Relationships, "stakeholder-classification"),
		InterestDegree:              floatAttrPointer(attrs, "interest-degree"),
	}
}
