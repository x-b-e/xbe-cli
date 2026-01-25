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

type doPredictionAgentsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	CustomInstructions string
}

func newDoPredictionAgentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction agent",
		Long: `Update an existing prediction agent.

Provide the prediction agent ID as an argument, then use flags to specify
which fields to update.

Updatable fields:
  --custom-instructions  Custom instructions to guide the agent

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update custom instructions
  xbe do prediction-agents update 123 --custom-instructions "Emphasize recent performance"

  # Output as JSON
  xbe do prediction-agents update 123 --custom-instructions "Focus on volatility" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionAgentsUpdate,
	}
	initDoPredictionAgentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionAgentsCmd.AddCommand(newDoPredictionAgentsUpdateCmd())
}

func initDoPredictionAgentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("custom-instructions", "", "Custom instructions to guide the agent")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionAgentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionAgentsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("custom-instructions") {
		attributes["custom-instructions"] = opts.CustomInstructions
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --custom-instructions")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-agents",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-agents/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prediction agent %s\n", row.ID)
	return nil
}

func parseDoPredictionAgentsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionAgentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customInstructions, _ := cmd.Flags().GetString("custom-instructions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionAgentsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		CustomInstructions: customInstructions,
	}, nil
}
