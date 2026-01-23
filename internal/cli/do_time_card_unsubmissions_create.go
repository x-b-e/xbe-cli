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

type doTimeCardUnsubmissionsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TimeCardID string
	Comment    string
}

type timeCardUnsubmissionRow struct {
	ID         string `json:"id"`
	TimeCardID string `json:"time_card_id,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newDoTimeCardUnsubmissionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unsubmit a time card",
		Long: `Unsubmit a time card.

The time card must currently be in submitted status.

Required:
  --time-card  Time card ID (required)

Optional:
  --comment    Status change comment`,
		Example: `  # Unsubmit a time card
  xbe do time-card-unsubmissions create --time-card 123

  # Unsubmit with a comment
  xbe do time-card-unsubmissions create --time-card 123 --comment "Needs edits"

  # Output as JSON
  xbe do time-card-unsubmissions create --time-card 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardUnsubmissionsCreate,
	}
	initDoTimeCardUnsubmissionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardUnsubmissionsCmd.AddCommand(newDoTimeCardUnsubmissionsCreateCmd())
}

func initDoTimeCardUnsubmissionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID (required)")
	cmd.Flags().String("comment", "", "Status change comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-card")
}

func runDoTimeCardUnsubmissionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardUnsubmissionsCreateOptions(cmd)
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

	if opts.TimeCardID == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   opts.TimeCardID,
			},
		},
	}

	data := map[string]any{
		"type":          "time-card-unsubmissions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-unsubmissions", jsonBody)
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

	row := buildTimeCardUnsubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card unsubmission %s\n", row.ID)
	return nil
}

func parseDoTimeCardUnsubmissionsCreateOptions(cmd *cobra.Command) (doTimeCardUnsubmissionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardUnsubmissionsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TimeCardID: timeCardID,
		Comment:    comment,
	}, nil
}

func buildTimeCardUnsubmissionRowFromSingle(resp jsonAPISingleResponse) timeCardUnsubmissionRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeCardUnsubmissionRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}
	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}
	return row
}
