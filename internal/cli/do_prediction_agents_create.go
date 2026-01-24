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

type doPredictionAgentsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	PredictionSubject  string
	CreatedBy          string
	CustomInstructions string
}

func newDoPredictionAgentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction agent",
		Long: `Create a prediction agent for a prediction subject.

Required flags:
  --prediction-subject  Prediction subject ID (required)

Optional flags:
  --created-by          Creator user ID
  --custom-instructions Custom instructions to guide the agent

Notes:
  - Prediction subjects must have at least two submitted predictions.
  - Only one active prediction agent is allowed per prediction subject.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction agent
  xbe do prediction-agents create --prediction-subject 123

  # Create with custom instructions
  xbe do prediction-agents create --prediction-subject 123 \
    --custom-instructions "Weight recent performance more heavily"

  # Output as JSON
  xbe do prediction-agents create --prediction-subject 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionAgentsCreate,
	}
	initDoPredictionAgentsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionAgentsCmd.AddCommand(newDoPredictionAgentsCreateCmd())
}

func initDoPredictionAgentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("custom-instructions", "", "Custom instructions to guide the agent")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionAgentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionAgentsCreateOptions(cmd)
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

	if opts.PredictionSubject == "" {
		err := fmt.Errorf("--prediction-subject is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("custom-instructions") {
		attributes["custom-instructions"] = opts.CustomInstructions
	}

	relationships := map[string]any{
		"prediction-subject": map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubject,
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
			"type":          "prediction-agents",
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

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-agents", jsonBody)
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

	row := predictionAgentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction agent %s\n", row.ID)
	return nil
}

func parseDoPredictionAgentsCreateOptions(cmd *cobra.Command) (doPredictionAgentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	createdBy, _ := cmd.Flags().GetString("created-by")
	customInstructions, _ := cmd.Flags().GetString("custom-instructions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionAgentsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		PredictionSubject:  predictionSubject,
		CreatedBy:          createdBy,
		CustomInstructions: customInstructions,
	}, nil
}
