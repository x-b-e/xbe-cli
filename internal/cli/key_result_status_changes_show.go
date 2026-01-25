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

type keyResultStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type keyResultStatusChangeDetails struct {
	ID          string `json:"id"`
	KeyResultID string `json:"key_result_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newKeyResultStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show key result status change details",
		Long: `Show the full details of a key result status change.

Output Fields:
  ID          Status change identifier
  KEY RESULT  Key result ID
  STATUS      Key result status
  CHANGED AT  Status change timestamp
  CHANGED BY  User who changed the status
  COMMENT     Status change comment

Arguments:
  <id>  Status change ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a status change
  xbe view key-result-status-changes show 123

  # Output as JSON
  xbe view key-result-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runKeyResultStatusChangesShow,
	}
	initKeyResultStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	keyResultStatusChangesCmd.AddCommand(newKeyResultStatusChangesShowCmd())
}

func initKeyResultStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseKeyResultStatusChangesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("key result status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[key-result-status-changes]", "key-result,status,changed-at,comment,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/key-result-status-changes/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildKeyResultStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderKeyResultStatusChangeDetails(cmd, details)
}

func parseKeyResultStatusChangesShowOptions(cmd *cobra.Command) (keyResultStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildKeyResultStatusChangeDetails(resp jsonAPISingleResponse) keyResultStatusChangeDetails {
	attrs := resp.Data.Attributes
	details := keyResultStatusChangeDetails{
		ID:        resp.Data.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["key-result"]; ok && rel.Data != nil {
		details.KeyResultID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
	}

	return details
}

func renderKeyResultStatusChangeDetails(cmd *cobra.Command, details keyResultStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.KeyResultID != "" {
		fmt.Fprintf(out, "Key Result: %s\n", details.KeyResultID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
