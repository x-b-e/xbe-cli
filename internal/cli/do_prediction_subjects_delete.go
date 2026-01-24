package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPredictionSubjectsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoPredictionSubjectsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prediction subject",
		Long: `Delete a prediction subject.

This permanently deletes the prediction subject.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The prediction subject ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a prediction subject
  xbe do prediction-subjects delete 123 --confirm

  # Get JSON output of deleted record
  xbe do prediction-subjects delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectsDelete,
	}
	initDoPredictionSubjectsDeleteFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectsCmd.AddCommand(newDoPredictionSubjectsDeleteCmd())
}

func initDoPredictionSubjectsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectsDeleteOptions(cmd)
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
		return fmt.Errorf("prediction subject id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subjects]", "name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/prediction-subjects/"+id, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var getResp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &getResp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	name := stringAttr(getResp.Data.Attributes, "name")
	row := buildPredictionSubjectRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/prediction-subjects/"+id)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted prediction subject %s (%s)\n", id, name)
	return nil
}

func parseDoPredictionSubjectsDeleteOptions(cmd *cobra.Command) (doPredictionSubjectsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
