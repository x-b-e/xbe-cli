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

type tenderReturnsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderReturnDetails struct {
	ID         string `json:"id"`
	TenderID   string `json:"tender_id,omitempty"`
	TenderType string `json:"tender_type,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newTenderReturnsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender return details",
		Long: `Show the full details of a tender return.

Output Fields:
  ID          Tender return identifier
  Tender ID   Tender identifier
  Tender Type Tender resource type
  Comment     Return comment

Arguments:
  <id>    The tender return ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tender return
  xbe view tender-returns show 123

  # Get JSON output
  xbe view tender-returns show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderReturnsShow,
	}
	initTenderReturnsShowFlags(cmd)
	return cmd
}

func init() {
	tenderReturnsCmd.AddCommand(newTenderReturnsShowCmd())
}

func initTenderReturnsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderReturnsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTenderReturnsShowOptions(cmd)
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
		return fmt.Errorf("tender return id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-returns/"+id, nil)
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

	details := buildTenderReturnDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderReturnDetails(cmd, details)
}

func parseTenderReturnsShowOptions(cmd *cobra.Command) (tenderReturnsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderReturnsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderReturnDetails(resp jsonAPISingleResponse) tenderReturnDetails {
	resource := resp.Data
	details := tenderReturnDetails{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		details.TenderID = rel.Data.ID
		details.TenderType = rel.Data.Type
	}

	return details
}

func renderTenderReturnDetails(cmd *cobra.Command, details tenderReturnDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderID != "" {
		fmt.Fprintf(out, "Tender ID: %s\n", details.TenderID)
	}
	if details.TenderType != "" {
		fmt.Fprintf(out, "Tender Type: %s\n", details.TenderType)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
