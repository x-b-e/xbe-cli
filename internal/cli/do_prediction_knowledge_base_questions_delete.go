package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPredictionKnowledgeBaseQuestionsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoPredictionKnowledgeBaseQuestionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prediction knowledge base question",
		Long: `Delete a prediction knowledge base question.

Provide the question ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a knowledge base question
  xbe do prediction-knowledge-base-questions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionKnowledgeBaseQuestionsDelete,
	}
	initDoPredictionKnowledgeBaseQuestionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doPredictionKnowledgeBaseQuestionsCmd.AddCommand(newDoPredictionKnowledgeBaseQuestionsDeleteCmd())
}

func initDoPredictionKnowledgeBaseQuestionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionKnowledgeBaseQuestionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionKnowledgeBaseQuestionsDeleteOptions(cmd, args)
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

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a prediction knowledge base question")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/prediction-knowledge-base-questions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted prediction knowledge base question %s\n", opts.ID)
	return nil
}

func parseDoPredictionKnowledgeBaseQuestionsDeleteOptions(cmd *cobra.Command, args []string) (doPredictionKnowledgeBaseQuestionsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionKnowledgeBaseQuestionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
