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

type doTenderCancellationsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Tender  string
	Comment string
}

type tenderCancellationRow struct {
	ID       string `json:"id"`
	TenderID string `json:"tender_id,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

func newDoTenderCancellationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cancel a tender",
		Long: `Cancel a tender.

Required flags:
  --tender  Tender ID (required)

Optional flags:
  --comment  Cancellation comment`,
		Example: `  # Cancel a tender
  xbe do tender-cancellations create --tender 123 --comment "Cancelled"`,
		Args: cobra.NoArgs,
		RunE: runDoTenderCancellationsCreate,
	}
	initDoTenderCancellationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderCancellationsCmd.AddCommand(newDoTenderCancellationsCreateCmd())
}

func initDoTenderCancellationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender", "", "Tender ID (required)")
	cmd.Flags().String("comment", "", "Cancellation comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderCancellationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderCancellationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.Tender) == "" {
		err := fmt.Errorf("--tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": "tenders",
				"id":   opts.Tender,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-cancellations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tender-cancellations", jsonBody)
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

	row := tenderCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender cancellation %s\n", row.ID)
	return nil
}

func tenderCancellationRowFromSingle(resp jsonAPISingleResponse) tenderCancellationRow {
	attrs := resp.Data.Attributes
	row := tenderCancellationRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderID = rel.Data.ID
	}

	return row
}

func parseDoTenderCancellationsCreateOptions(cmd *cobra.Command) (doTenderCancellationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tender, _ := cmd.Flags().GetString("tender")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderCancellationsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Tender:  tender,
		Comment: comment,
	}, nil
}
