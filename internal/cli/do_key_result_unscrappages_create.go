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

type doKeyResultUnscrappagesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	KeyResultID string
	Comment     string
}

type keyResultUnscrappageRow struct {
	ID          string `json:"id"`
	KeyResultID string `json:"key_result_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newDoKeyResultUnscrappagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unscrap a key result",
		Long: `Unscrap a key result.

This action restores a key result from scrapped status to its most recent
non-scrapped status (or unknown when none exists).

Required flags:
  --key-result   Key result ID

Optional flags:
  --comment      Comment for the unscrappage`,
		Example: `  # Unscrap a key result
  xbe do key-result-unscrappages create --key-result 123 --comment "Restoring key result"

  # JSON output
  xbe do key-result-unscrappages create --key-result 123 --comment "Restoring key result" --json`,
		RunE: runDoKeyResultUnscrappagesCreate,
	}
	initDoKeyResultUnscrappagesCreateFlags(cmd)
	return cmd
}

func init() {
	doKeyResultUnscrappagesCmd.AddCommand(newDoKeyResultUnscrappagesCreateCmd())
}

func initDoKeyResultUnscrappagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("key-result", "", "Key result ID (required)")
	cmd.Flags().String("comment", "", "Comment for the unscrappage")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("key-result")
}

func runDoKeyResultUnscrappagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoKeyResultUnscrappagesCreateOptions(cmd)
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
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"key-result": map[string]any{
			"data": map[string]any{
				"type": "key-results",
				"id":   opts.KeyResultID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "key-result-unscrappages",
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

	body, _, err := client.Post(cmd.Context(), "/v1/key-result-unscrappages", jsonBody)
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

	row := buildKeyResultUnscrappageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created key result unscrappage %s\n", row.ID)
	return nil
}

func parseDoKeyResultUnscrappagesCreateOptions(cmd *cobra.Command) (doKeyResultUnscrappagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	keyResultID, _ := cmd.Flags().GetString("key-result")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doKeyResultUnscrappagesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		KeyResultID: keyResultID,
		Comment:     comment,
	}, nil
}

func buildKeyResultUnscrappageRowFromSingle(resp jsonAPISingleResponse) keyResultUnscrappageRow {
	resource := resp.Data
	row := keyResultUnscrappageRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["key-result"]; ok && rel.Data != nil {
		row.KeyResultID = rel.Data.ID
	}
	return row
}
