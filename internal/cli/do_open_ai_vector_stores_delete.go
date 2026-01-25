package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doOpenAiVectorStoresDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoOpenAiVectorStoresDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an OpenAI vector store",
		Long: `Delete an OpenAI vector store.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The vector store ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a vector store
  xbe do open-ai-vector-stores delete 123 --confirm

  # Get JSON output
  xbe do open-ai-vector-stores delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOpenAiVectorStoresDelete,
	}
	initDoOpenAiVectorStoresDeleteFlags(cmd)
	return cmd
}

func init() {
	doOpenAiVectorStoresCmd.AddCommand(newDoOpenAiVectorStoresDeleteCmd())
}

func initDoOpenAiVectorStoresDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenAiVectorStoresDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOpenAiVectorStoresDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("open ai vector store id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/open-ai-vector-stores/"+id)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      id,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted open ai vector store %s\n", id)
	return nil
}

func parseDoOpenAiVectorStoresDeleteOptions(cmd *cobra.Command) (doOpenAiVectorStoresDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenAiVectorStoresDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
