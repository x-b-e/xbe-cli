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

type doTenderReturnsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TenderType string
	TenderID   string
	Comment    string
}

type tenderReturnRowCreate struct {
	ID         string `json:"id"`
	TenderID   string `json:"tender_id,omitempty"`
	TenderType string `json:"tender_type,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newDoTenderReturnsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Return a tender",
		Long: `Return a tender.

Tender returns transition accepted tenders to returned status.

Required flags:
  --tender-type   Tender resource type (required, e.g., broker-tenders)
  --tender-id     Tender ID (required)

Optional flags:
  --comment       Comment for the tender return`,
		Example: `  # Return a tender
  xbe do tender-returns create --tender-type broker-tenders --tender-id 123 --comment "Returned"

  # JSON output
  xbe do tender-returns create --tender-type broker-tenders --tender-id 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderReturnsCreate,
	}
	initDoTenderReturnsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderReturnsCmd.AddCommand(newDoTenderReturnsCreateCmd())
}

func initDoTenderReturnsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-type", "", "Tender resource type (required, e.g., broker-tenders)")
	cmd.Flags().String("tender-id", "", "Tender ID (required)")
	cmd.Flags().String("comment", "", "Comment for the tender return")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderReturnsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderReturnsCreateOptions(cmd)
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

	if opts.TenderType == "" {
		err := fmt.Errorf("--tender-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.TenderID == "" {
		err := fmt.Errorf("--tender-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": opts.TenderType,
				"id":   opts.TenderID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-returns",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tender-returns", jsonBody)
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

	row := buildTenderReturnRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender return %s\n", row.ID)
	return nil
}

func parseDoTenderReturnsCreateOptions(cmd *cobra.Command) (doTenderReturnsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderType, _ := cmd.Flags().GetString("tender-type")
	tenderID, _ := cmd.Flags().GetString("tender-id")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderReturnsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TenderType: tenderType,
		TenderID:   tenderID,
		Comment:    comment,
	}, nil
}

func buildTenderReturnRowFromSingle(resp jsonAPISingleResponse) tenderReturnRowCreate {
	resource := resp.Data
	row := tenderReturnRowCreate{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderID = rel.Data.ID
		row.TenderType = rel.Data.Type
	}
	return row
}
