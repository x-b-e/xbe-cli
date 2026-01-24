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

type doPredictionKnowledgeBasesCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	BrokerID string
}

func newDoPredictionKnowledgeBasesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction knowledge base",
		Long: `Create a prediction knowledge base.

Required flags:
  --broker   Broker ID`,
		Example: `  # Create a prediction knowledge base
  xbe do prediction-knowledge-bases create --broker 123

  # Output as JSON
  xbe do prediction-knowledge-bases create --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionKnowledgeBasesCreate,
	}
	initDoPredictionKnowledgeBasesCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionKnowledgeBasesCmd.AddCommand(newDoPredictionKnowledgeBasesCreateCmd())
}

func initDoPredictionKnowledgeBasesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
}

func runDoPredictionKnowledgeBasesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionKnowledgeBasesCreateOptions(cmd)
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
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-knowledge-bases",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-knowledge-bases", jsonBody)
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

	row := predictionKnowledgeBaseRowFromResource(resp.Data)
	if row.BrokerID == "" {
		row.BrokerID = opts.BrokerID
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.BrokerID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created prediction knowledge base %s for broker %s\n", row.ID, row.BrokerID)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction knowledge base %s\n", row.ID)
	return nil
}

func parseDoPredictionKnowledgeBasesCreateOptions(cmd *cobra.Command) (doPredictionKnowledgeBasesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionKnowledgeBasesCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		BrokerID: brokerID,
	}, nil
}
