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

type doEmailAddressStatusesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	EmailAddress string
}

type emailAddressStatusDetails struct {
	ID             string `json:"id"`
	EmailAddress   string `json:"email_address,omitempty"`
	IsRejected     bool   `json:"is_rejected,omitempty"`
	RejectReason   string `json:"reject_reason,omitempty"`
	LastRejectedAt string `json:"last_rejected_at,omitempty"`
	RejectDetail   string `json:"reject_detail,omitempty"`
	Details        any    `json:"details,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

func newDoEmailAddressStatusesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Check an email address",
		Long: `Check an email address status.

Email address status lookups are restricted to admin users.

Required flags:
  --email-address  Email address to check`,
		Example: `  # Check an email address
  xbe do email-address-statuses create --email-address "user@example.com"

  # JSON output
  xbe do email-address-statuses create --email-address "user@example.com" --json`,
		Args: cobra.NoArgs,
		RunE: runDoEmailAddressStatusesCreate,
	}
	initDoEmailAddressStatusesCreateFlags(cmd)
	return cmd
}

func init() {
	doEmailAddressStatusesCmd.AddCommand(newDoEmailAddressStatusesCreateCmd())
}

func initDoEmailAddressStatusesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("email-address", "", "Email address to check (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("email-address")
}

func runDoEmailAddressStatusesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEmailAddressStatusesCreateOptions(cmd)
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

	email := strings.TrimSpace(opts.EmailAddress)
	if email == "" {
		err := errors.New("email address is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"email-address": email,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "email-address-statuses",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/email-address-statuses", jsonBody)
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

	details := buildEmailAddressStatusDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEmailAddressStatusDetails(cmd, details)
}

func parseDoEmailAddressStatusesCreateOptions(cmd *cobra.Command) (doEmailAddressStatusesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEmailAddressStatusesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		EmailAddress: emailAddress,
	}, nil
}

func buildEmailAddressStatusDetails(resp jsonAPISingleResponse) emailAddressStatusDetails {
	attrs := resp.Data.Attributes
	details := emailAddressStatusDetails{
		ID:             resp.Data.ID,
		EmailAddress:   stringAttr(attrs, "email-address"),
		IsRejected:     boolAttr(attrs, "is-rejected"),
		RejectReason:   stringAttr(attrs, "reject-reason"),
		LastRejectedAt: formatDateTime(stringAttr(attrs, "last-rejected-at")),
		RejectDetail:   stringAttr(attrs, "reject-detail"),
		Details:        anyAttr(attrs, "details"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	if details.EmailAddress == "" {
		details.EmailAddress = resp.Data.ID
	}

	return details
}

func renderEmailAddressStatusDetails(cmd *cobra.Command, details emailAddressStatusDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EmailAddress != "" {
		fmt.Fprintf(out, "Email Address: %s\n", details.EmailAddress)
	}
	fmt.Fprintf(out, "Rejected: %s\n", formatBool(details.IsRejected))
	if details.RejectReason != "" {
		fmt.Fprintf(out, "Reject Reason: %s\n", details.RejectReason)
	}
	if details.LastRejectedAt != "" {
		fmt.Fprintf(out, "Last Rejected At: %s\n", details.LastRejectedAt)
	}
	if details.RejectDetail != "" {
		fmt.Fprintf(out, "Reject Detail: %s\n", details.RejectDetail)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.Details != nil {
		fmt.Fprintf(out, "Details: %s\n", formatAny(details.Details))
	}

	return nil
}
