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

type doPredictionSubjectGapPortionsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	PredictionSubjectGap string
	Name                 string
	Amount               string
	Status               string
	Description          string
	CreatedBy            string
}

func newDoPredictionSubjectGapPortionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction subject gap portion",
		Long: `Create a prediction subject gap portion.

Required flags:
  --prediction-subject-gap  Prediction subject gap ID
  --name                    Portion name
  --amount                  Portion amount
  --status                  Portion status (draft/approved)

Optional flags:
  --description             Portion description
  --created-by              Creator user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction subject gap portion
  xbe do prediction-subject-gap-portions create \
    --prediction-subject-gap 123 \
    --name "Labor" \
    --amount 42 \
    --status draft

  # Create with description
  xbe do prediction-subject-gap-portions create \
    --prediction-subject-gap 123 \
    --name "Equipment" \
    --amount 15.5 \
    --status approved \
    --description "Equipment availability impact"

  # Output as JSON
  xbe do prediction-subject-gap-portions create --prediction-subject-gap 123 --name "Labor" --amount 42 --status draft --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectGapPortionsCreate,
	}
	initDoPredictionSubjectGapPortionsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapPortionsCmd.AddCommand(newDoPredictionSubjectGapPortionsCreateCmd())
}

func initDoPredictionSubjectGapPortionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject-gap", "", "Prediction subject gap ID (required)")
	cmd.Flags().String("name", "", "Portion name (required)")
	cmd.Flags().String("amount", "", "Portion amount (required)")
	cmd.Flags().String("status", "", "Portion status (draft/approved) (required)")
	cmd.Flags().String("description", "", "Portion description")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("prediction-subject-gap")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("amount")
	_ = cmd.MarkFlagRequired("status")
}

func runDoPredictionSubjectGapPortionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectGapPortionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.PredictionSubjectGap) == "" {
		err := fmt.Errorf("--prediction-subject-gap is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Name) == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Amount) == "" {
		err := fmt.Errorf("--amount is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Status) == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":   opts.Name,
		"amount": opts.Amount,
		"status": opts.Status,
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	relationships := map[string]any{
		"prediction-subject-gap": map[string]any{
			"data": map[string]any{
				"type": "prediction-subject-gaps",
				"id":   opts.PredictionSubjectGap,
			},
		},
	}

	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-subject-gap-portions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subject-gap-portions", jsonBody)
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

	row := predictionSubjectGapPortionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction subject gap portion %s\n", row.ID)
	return nil
}

func parseDoPredictionSubjectGapPortionsCreateOptions(cmd *cobra.Command) (doPredictionSubjectGapPortionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubjectGap, _ := cmd.Flags().GetString("prediction-subject-gap")
	name, _ := cmd.Flags().GetString("name")
	amount, _ := cmd.Flags().GetString("amount")
	status, _ := cmd.Flags().GetString("status")
	description, _ := cmd.Flags().GetString("description")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapPortionsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		PredictionSubjectGap: predictionSubjectGap,
		Name:                 name,
		Amount:               amount,
		Status:               status,
		Description:          description,
		CreatedBy:            createdBy,
	}, nil
}
