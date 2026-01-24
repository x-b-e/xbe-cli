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

type doProcessNonProcessedTimeCardTimeChangesCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	TimeCardTimeChangeIDs []string
	DeleteUnprocessed     bool
}

type processNonProcessedTimeCardTimeChangesRow struct {
	ID                    string   `json:"id"`
	TimeCardTimeChangeIDs []string `json:"time_card_time_change_ids,omitempty"`
	DeleteUnprocessed     bool     `json:"delete_unprocessed,omitempty"`
	InvoiceIDs            []string `json:"invoice_ids,omitempty"`
	Results               any      `json:"results,omitempty"`
	Messages              any      `json:"messages,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

func newDoProcessNonProcessedTimeCardTimeChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Process time card time changes",
		Long: `Process non-processed time card time changes.

Required flags:
  --time-card-time-change-ids  Time card time change IDs (required, comma-separated or repeated)

Optional flags:
  --delete-unprocessed  Delete unprocessed time card time changes after processing (default: true)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Process time card time changes
  xbe do process-non-processed-time-card-time-changes create --time-card-time-change-ids 123,456

  # Keep unprocessed changes
  xbe do process-non-processed-time-card-time-changes create --time-card-time-change-ids 123 --delete-unprocessed false

  # Output as JSON
  xbe do process-non-processed-time-card-time-changes create --time-card-time-change-ids 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProcessNonProcessedTimeCardTimeChangesCreate,
	}
	initDoProcessNonProcessedTimeCardTimeChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doProcessNonProcessedTimeCardTimeChangesCmd.AddCommand(newDoProcessNonProcessedTimeCardTimeChangesCreateCmd())
}

func initDoProcessNonProcessedTimeCardTimeChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("time-card-time-change-ids", nil, "Time card time change IDs (required, comma-separated or repeated)")
	cmd.Flags().Bool("delete-unprocessed", true, "Delete unprocessed time card time changes after processing")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProcessNonProcessedTimeCardTimeChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProcessNonProcessedTimeCardTimeChangesCreateOptions(cmd)
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

	timeCardTimeChangeIDs := make([]string, 0, len(opts.TimeCardTimeChangeIDs))
	for _, id := range opts.TimeCardTimeChangeIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			timeCardTimeChangeIDs = append(timeCardTimeChangeIDs, trimmed)
		}
	}
	if len(timeCardTimeChangeIDs) == 0 {
		err := fmt.Errorf("--time-card-time-change-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"time-card-time-change-ids": timeCardTimeChangeIDs,
		"delete-unprocessed":        opts.DeleteUnprocessed,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "process-non-processed-time-card-time-changes",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/sombreros/process-non-processed-time-card-time-changes", jsonBody)
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

	row := processNonProcessedTimeCardTimeChangesRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created process non-processed time card time changes %s\n", row.ID)
	return nil
}

func processNonProcessedTimeCardTimeChangesRowFromSingle(resp jsonAPISingleResponse) processNonProcessedTimeCardTimeChangesRow {
	attrs := resp.Data.Attributes
	row := processNonProcessedTimeCardTimeChangesRow{
		ID:                    resp.Data.ID,
		TimeCardTimeChangeIDs: stringSliceAttr(attrs, "time-card-time-change-ids"),
		DeleteUnprocessed:     boolAttr(attrs, "delete-unprocessed"),
		InvoiceIDs:            stringSliceAttr(attrs, "invoice-ids"),
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if attrs != nil {
		if value, ok := attrs["results"]; ok {
			row.Results = value
		}
		if value, ok := attrs["messages"]; ok {
			row.Messages = value
		}
	}

	return row
}

func parseDoProcessNonProcessedTimeCardTimeChangesCreateOptions(cmd *cobra.Command) (doProcessNonProcessedTimeCardTimeChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardTimeChangeIDs, _ := cmd.Flags().GetStringSlice("time-card-time-change-ids")
	deleteUnprocessed, _ := cmd.Flags().GetBool("delete-unprocessed")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProcessNonProcessedTimeCardTimeChangesCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		TimeCardTimeChangeIDs: timeCardTimeChangeIDs,
		DeleteUnprocessed:     deleteUnprocessed,
	}, nil
}
