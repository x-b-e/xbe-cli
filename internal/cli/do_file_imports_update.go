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

type doFileImportsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	ProcessedAt string
	Note        string
}

func newDoFileImportsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a file import",
		Long: `Update a file import.

Optional flags:
  --processed-at   Processed at timestamp (ISO 8601)
  --note           Import note`,
		Example: `  # Update a file import note
  xbe do file-imports update 123 --note "Processed successfully"

  # Update processed-at
  xbe do file-imports update 123 --processed-at 2024-01-02T03:04:05Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoFileImportsUpdate,
	}
	initDoFileImportsUpdateFlags(cmd)
	return cmd
}

func init() {
	doFileImportsCmd.AddCommand(newDoFileImportsUpdateCmd())
}

func initDoFileImportsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("processed-at", "", "Processed at timestamp (ISO 8601)")
	cmd.Flags().String("note", "", "Import note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFileImportsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoFileImportsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("processed-at") {
		attributes["processed-at"] = opts.ProcessedAt
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "file-imports",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/file-imports/"+opts.ID, jsonBody)
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

	row := buildFileImportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated file import %s\n", row.ID)
	return nil
}

func parseDoFileImportsUpdateOptions(cmd *cobra.Command, args []string) (doFileImportsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	processedAt, _ := cmd.Flags().GetString("processed-at")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFileImportsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		ProcessedAt: processedAt,
		Note:        note,
	}, nil
}
