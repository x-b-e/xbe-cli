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

type objectiveStakeholderClassificationQuotesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type objectiveStakeholderClassificationQuoteDetails struct {
	ID                                   string `json:"id"`
	Content                              string `json:"content,omitempty"`
	IsGenerated                          bool   `json:"is_generated"`
	ObjectiveStakeholderClassificationID string `json:"objective_stakeholder_classification_id,omitempty"`
	CreatedAt                            string `json:"created_at,omitempty"`
	UpdatedAt                            string `json:"updated_at,omitempty"`
}

func newObjectiveStakeholderClassificationQuotesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show objective stakeholder classification quote details",
		Long: `Show full details of an objective stakeholder classification quote.

Output Fields:
  ID              Quote identifier
  Content         Quote content
  Generated       Whether the quote was generated
  Classification  Objective stakeholder classification ID
  Created At      Creation timestamp
  Updated At      Last update timestamp

Arguments:
  <id>    Objective stakeholder classification quote ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a quote
  xbe view objective-stakeholder-classification-quotes show 123

  # JSON output
  xbe view objective-stakeholder-classification-quotes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runObjectiveStakeholderClassificationQuotesShow,
	}
	initObjectiveStakeholderClassificationQuotesShowFlags(cmd)
	return cmd
}

func init() {
	objectiveStakeholderClassificationQuotesCmd.AddCommand(newObjectiveStakeholderClassificationQuotesShowCmd())
}

func initObjectiveStakeholderClassificationQuotesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveStakeholderClassificationQuotesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseObjectiveStakeholderClassificationQuotesShowOptions(cmd)
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
		return fmt.Errorf("objective stakeholder classification quote id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[objective-stakeholder-classification-quotes]", "content,is-generated,created-at,updated-at,objective-stakeholder-classification")

	body, status, err := client.Get(cmd.Context(), "/v1/objective-stakeholder-classification-quotes/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderObjectiveStakeholderClassificationQuotesShowUnavailable(cmd, opts.JSON)
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

	details := buildObjectiveStakeholderClassificationQuoteDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderObjectiveStakeholderClassificationQuoteDetails(cmd, details)
}

func parseObjectiveStakeholderClassificationQuotesShowOptions(cmd *cobra.Command) (objectiveStakeholderClassificationQuotesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveStakeholderClassificationQuotesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildObjectiveStakeholderClassificationQuoteDetails(resp jsonAPISingleResponse) objectiveStakeholderClassificationQuoteDetails {
	attrs := resp.Data.Attributes
	details := objectiveStakeholderClassificationQuoteDetails{
		ID:          resp.Data.ID,
		Content:     stringAttr(attrs, "content"),
		IsGenerated: boolAttr(attrs, "is-generated"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["objective-stakeholder-classification"]; ok && rel.Data != nil {
		details.ObjectiveStakeholderClassificationID = rel.Data.ID
	}

	return details
}

func renderObjectiveStakeholderClassificationQuotesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), objectiveStakeholderClassificationQuoteDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Objective stakeholder classification quote not found.")
	return nil
}

func renderObjectiveStakeholderClassificationQuoteDetails(cmd *cobra.Command, details objectiveStakeholderClassificationQuoteDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Content: %s\n", formatOptional(details.Content))
	if details.IsGenerated {
		fmt.Fprintln(out, "Generated: yes")
	} else {
		fmt.Fprintln(out, "Generated: no")
	}
	fmt.Fprintf(out, "Classification: %s\n", formatOptional(details.ObjectiveStakeholderClassificationID))
	fmt.Fprintf(out, "Created At: %s\n", formatOptional(details.CreatedAt))
	fmt.Fprintf(out, "Updated At: %s\n", formatOptional(details.UpdatedAt))
	return nil
}
