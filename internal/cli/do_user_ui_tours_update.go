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

type doUserUiToursUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	CompletedAt string
	SkippedAt   string
}

func newDoUserUiToursUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user UI tour",
		Long: `Update a user UI tour completion or skip record.

Optional flags:
  --completed-at  Completion timestamp (ISO 8601, empty to clear)
  --skipped-at    Skipped timestamp (ISO 8601, empty to clear)

Note: Exactly one of --completed-at or --skipped-at must be set after update.`,
		Example: `  # Update completion timestamp
  xbe do user-ui-tours update 123 --completed-at 2025-01-10T12:00:00Z

  # Switch from completed to skipped
  xbe do user-ui-tours update 123 --completed-at "" --skipped-at 2025-01-10T12:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUserUiToursUpdate,
	}
	initDoUserUiToursUpdateFlags(cmd)
	return cmd
}

func init() {
	doUserUiToursCmd.AddCommand(newDoUserUiToursUpdateCmd())
}

func initDoUserUiToursUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("completed-at", "", "Completion timestamp (ISO 8601)")
	cmd.Flags().String("skipped-at", "", "Skipped timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserUiToursUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserUiToursUpdateOptions(cmd, args)
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

	completedChanged := cmd.Flags().Changed("completed-at")
	skippedChanged := cmd.Flags().Changed("skipped-at")
	if !completedChanged && !skippedChanged {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	completedValue := strings.TrimSpace(opts.CompletedAt)
	skippedValue := strings.TrimSpace(opts.SkippedAt)
	if completedChanged && skippedChanged {
		if completedValue == "" && skippedValue == "" {
			err := fmt.Errorf("one of --completed-at or --skipped-at must be set")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if completedValue != "" && skippedValue != "" {
			err := fmt.Errorf("only one of --completed-at or --skipped-at may be set")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if completedChanged && completedValue == "" && !skippedChanged {
		err := fmt.Errorf("--completed-at cannot be cleared without setting --skipped-at")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if skippedChanged && skippedValue == "" && !completedChanged {
		err := fmt.Errorf("--skipped-at cannot be cleared without setting --completed-at")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if completedChanged {
		attributes["completed-at"] = opts.CompletedAt
	}
	if skippedChanged {
		attributes["skipped-at"] = opts.SkippedAt
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "user-ui-tours",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/user-ui-tours/"+opts.ID, jsonBody)
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

	row := userUiTourRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user UI tour %s\n", row.ID)
	return nil
}

func parseDoUserUiToursUpdateOptions(cmd *cobra.Command, args []string) (doUserUiToursUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	completedAt, _ := cmd.Flags().GetString("completed-at")
	skippedAt, _ := cmd.Flags().GetString("skipped-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserUiToursUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		CompletedAt: completedAt,
		SkippedAt:   skippedAt,
	}, nil
}
