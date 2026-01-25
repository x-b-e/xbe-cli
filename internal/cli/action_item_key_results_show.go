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

type actionItemKeyResultsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type actionItemKeyResultDetails struct {
	ID              string `json:"id"`
	ActionItemID    string `json:"action_item_id,omitempty"`
	ActionItemTitle string `json:"action_item_title,omitempty"`
	KeyResultID     string `json:"key_result_id,omitempty"`
	KeyResultTitle  string `json:"key_result_title,omitempty"`
}

func newActionItemKeyResultsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item key result details",
		Long: `Show the full details of an action item key result link.

Output Fields:
  ID           Link identifier
  Action Item  Action item
  Key Result   Key result

Arguments:
  <id>    Action item key result ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an action item key result link
  xbe view action-item-key-results show 123

  # JSON output
  xbe view action-item-key-results show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemKeyResultsShow,
	}
	initActionItemKeyResultsShowFlags(cmd)
	return cmd
}

func init() {
	actionItemKeyResultsCmd.AddCommand(newActionItemKeyResultsShowCmd())
}

func initActionItemKeyResultsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemKeyResultsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseActionItemKeyResultsShowOptions(cmd)
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
		return fmt.Errorf("action item key result id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-key-results]", "action-item,key-result")
	query.Set("fields[action-items]", "title")
	query.Set("fields[key-results]", "title")
	query.Set("include", "action-item,key-result")

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-key-results/"+id, query)
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

	details := buildActionItemKeyResultDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemKeyResultDetails(cmd, details)
}

func parseActionItemKeyResultsShowOptions(cmd *cobra.Command) (actionItemKeyResultsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemKeyResultsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildActionItemKeyResultDetails(resp jsonAPISingleResponse) actionItemKeyResultDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := actionItemKeyResultDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["action-item"]; ok && rel.Data != nil {
		details.ActionItemID = rel.Data.ID
		if actionItem, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ActionItemTitle = strings.TrimSpace(stringAttr(actionItem.Attributes, "title"))
		}
	}

	if rel, ok := resp.Data.Relationships["key-result"]; ok && rel.Data != nil {
		details.KeyResultID = rel.Data.ID
		if keyResult, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.KeyResultTitle = strings.TrimSpace(stringAttr(keyResult.Attributes, "title"))
		}
	}

	return details
}

func renderActionItemKeyResultDetails(cmd *cobra.Command, details actionItemKeyResultDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Action Item", details.ActionItemTitle, details.ActionItemID)
	writeLabelWithID(out, "Key Result", details.KeyResultTitle, details.KeyResultID)

	return nil
}
