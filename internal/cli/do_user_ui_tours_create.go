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

type doUserUiToursCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	CompletedAt string
	SkippedAt   string
	UserID      string
	UiTourID    string
}

func newDoUserUiToursCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user UI tour",
		Long: `Create a user UI tour completion or skip record.

Required flags:
  --user           User ID (required)
  --ui-tour        UI tour ID (required)
  --completed-at   Completion timestamp (ISO 8601, required when not skipped)
  --skipped-at     Skipped timestamp (ISO 8601, required when not completed)

Note: Exactly one of --completed-at or --skipped-at must be set.`,
		Example: `  # Mark a UI tour completed
  xbe do user-ui-tours create --user 123 --ui-tour 456 --completed-at 2025-01-10T12:00:00Z

  # Mark a UI tour skipped
  xbe do user-ui-tours create --user 123 --ui-tour 456 --skipped-at 2025-01-10T12:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoUserUiToursCreate,
	}
	initDoUserUiToursCreateFlags(cmd)
	return cmd
}

func init() {
	doUserUiToursCmd.AddCommand(newDoUserUiToursCreateCmd())
}

func initDoUserUiToursCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("ui-tour", "", "UI tour ID (required)")
	cmd.Flags().String("completed-at", "", "Completion timestamp (ISO 8601)")
	cmd.Flags().String("skipped-at", "", "Skipped timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserUiToursCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserUiToursCreateOptions(cmd)
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

	if strings.TrimSpace(opts.UserID) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.UiTourID) == "" {
		err := fmt.Errorf("--ui-tour is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	completedAt := strings.TrimSpace(opts.CompletedAt)
	skippedAt := strings.TrimSpace(opts.SkippedAt)
	if completedAt == "" && skippedAt == "" {
		err := fmt.Errorf("one of --completed-at or --skipped-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if completedAt != "" && skippedAt != "" {
		err := fmt.Errorf("only one of --completed-at or --skipped-at may be set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if completedAt != "" {
		attributes["completed-at"] = completedAt
	}
	if skippedAt != "" {
		attributes["skipped-at"] = skippedAt
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"ui-tour": map[string]any{
			"data": map[string]any{
				"type": "ui-tours",
				"id":   opts.UiTourID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-ui-tours",
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

	body, _, err := client.Post(cmd.Context(), "/v1/user-ui-tours", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created user UI tour %s\n", row.ID)
	return nil
}

func parseDoUserUiToursCreateOptions(cmd *cobra.Command) (doUserUiToursCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	uiTourID, _ := cmd.Flags().GetString("ui-tour")
	completedAt, _ := cmd.Flags().GetString("completed-at")
	skippedAt, _ := cmd.Flags().GetString("skipped-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserUiToursCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		CompletedAt: completedAt,
		SkippedAt:   skippedAt,
		UserID:      userID,
		UiTourID:    uiTourID,
	}, nil
}
