package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type keyResultScrappagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	ID      string
}

type keyResultScrappageDetails struct {
	ID          string `json:"id"`
	KeyResultID string `json:"key_result_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newKeyResultScrappagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show key result scrappage details",
		Long: `Show full details of a key result scrappage.

Output Fields:
  ID          Scrappage identifier
  KEY RESULT  Key result ID
  COMMENT     Comment

Arguments:
  <id>    Key result scrappage ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a key result scrappage
  xbe view key-result-scrappages show 123

  # JSON output
  xbe view key-result-scrappages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runKeyResultScrappagesShow,
	}
	initKeyResultScrappagesShowFlags(cmd)
	return cmd
}

func init() {
	keyResultScrappagesCmd.AddCommand(newKeyResultScrappagesShowCmd())
}

func initKeyResultScrappagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultScrappagesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseKeyResultScrappagesShowOptions(cmd)
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
		return fmt.Errorf("key result scrappage id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[key-result-scrappages]", "key-result,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/key-result-scrappages/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderKeyResultScrappagesShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildKeyResultScrappageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderKeyResultScrappageDetails(cmd, details)
}

func renderKeyResultScrappagesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), keyResultScrappageDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Key result scrappages are write-only; show is not available.")
	return nil
}

func parseKeyResultScrappagesShowOptions(cmd *cobra.Command) (keyResultScrappagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultScrappagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildKeyResultScrappageDetails(resp jsonAPISingleResponse) keyResultScrappageDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := keyResultScrappageDetails{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["key-result"]; ok && rel.Data != nil {
		details.KeyResultID = rel.Data.ID
	}

	return details
}

func renderKeyResultScrappageDetails(cmd *cobra.Command, details keyResultScrappageDetails) error {
	fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", details.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "Key Result: %s\n", details.KeyResultID)
	if details.Comment != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Comment: %s\n", details.Comment)
	}
	return nil
}
