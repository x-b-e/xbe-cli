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

type textMessagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type textMessageDetails struct {
	ID                  string `json:"id"`
	Status              string `json:"status,omitempty"`
	Direction           string `json:"direction,omitempty"`
	From                string `json:"from,omitempty"`
	To                  string `json:"to,omitempty"`
	DateSent            string `json:"date_sent,omitempty"`
	DateCreated         string `json:"date_created,omitempty"`
	DateUpdated         string `json:"date_updated,omitempty"`
	ErrorCode           string `json:"error_code,omitempty"`
	ErrorMessage        string `json:"error_message,omitempty"`
	Body                string `json:"body,omitempty"`
	AccountSID          string `json:"account_sid,omitempty"`
	MessagingServiceSID string `json:"messaging_service_sid,omitempty"`
	NumMedia            int    `json:"num_media,omitempty"`
	NumSegments         int    `json:"num_segments,omitempty"`
	Price               string `json:"price,omitempty"`
	PriceUnit           string `json:"price_unit,omitempty"`
	SubresourceURIs     any    `json:"subresource_uris,omitempty"`
	URI                 string `json:"uri,omitempty"`
	Webhooks            any    `json:"webhooks,omitempty"`
}

func newTextMessagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <sid>",
		Short: "Show text message details",
		Long: `Show full details of a text message.

Output Fields:
  Message metadata (status, direction, timestamps)
  Sender and recipient numbers
  Message body and error details
  Twilio identifiers and pricing
  Webhook events (if available)

Arguments:
  <sid>    The Twilio message SID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show text message details
  xbe view text-messages show SM123

  # JSON output
  xbe view text-messages show SM123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTextMessagesShow,
	}
	initTextMessagesShowFlags(cmd)
	return cmd
}

func init() {
	textMessagesCmd.AddCommand(newTextMessagesShowCmd())
}

func initTextMessagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTextMessagesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTextMessagesShowOptions(cmd)
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

	sid := strings.TrimSpace(args[0])
	if sid == "" {
		return fmt.Errorf("text message sid is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[text-messages]", "date-sent,status,error-code,error-message,date-created,date-updated,body,direction,from,to,account-sid,messaging-service-sid,num-media,num-segments,price,price-unit,subresource-uris,uri,webhooks")

	body, _, err := client.Get(cmd.Context(), "/v1/text-messages/"+sid, query)
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

	details := buildTextMessageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTextMessageDetails(cmd, details)
}

func parseTextMessagesShowOptions(cmd *cobra.Command) (textMessagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return textMessagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTextMessageDetails(resp jsonAPISingleResponse) textMessageDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return textMessageDetails{
		ID:                  resource.ID,
		Status:              stringAttr(attrs, "status"),
		Direction:           stringAttr(attrs, "direction"),
		From:                stringAttr(attrs, "from"),
		To:                  stringAttr(attrs, "to"),
		DateSent:            formatDateTime(stringAttr(attrs, "date-sent")),
		DateCreated:         formatDateTime(stringAttr(attrs, "date-created")),
		DateUpdated:         formatDateTime(stringAttr(attrs, "date-updated")),
		ErrorCode:           stringAttr(attrs, "error-code"),
		ErrorMessage:        stringAttr(attrs, "error-message"),
		Body:                stringAttr(attrs, "body"),
		AccountSID:          stringAttr(attrs, "account-sid"),
		MessagingServiceSID: stringAttr(attrs, "messaging-service-sid"),
		NumMedia:            intAttr(attrs, "num-media"),
		NumSegments:         intAttr(attrs, "num-segments"),
		Price:               stringAttr(attrs, "price"),
		PriceUnit:           stringAttr(attrs, "price-unit"),
		SubresourceURIs:     anyAttr(attrs, "subresource-uris"),
		URI:                 stringAttr(attrs, "uri"),
		Webhooks:            anyAttr(attrs, "webhooks"),
	}
}

func renderTextMessageDetails(cmd *cobra.Command, details textMessageDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Direction != "" {
		fmt.Fprintf(out, "Direction: %s\n", details.Direction)
	}
	if details.From != "" {
		fmt.Fprintf(out, "From: %s\n", details.From)
	}
	if details.To != "" {
		fmt.Fprintf(out, "To: %s\n", details.To)
	}
	if details.DateSent != "" {
		fmt.Fprintf(out, "Date Sent: %s\n", details.DateSent)
	}
	if details.DateCreated != "" {
		fmt.Fprintf(out, "Date Created: %s\n", details.DateCreated)
	}
	if details.DateUpdated != "" {
		fmt.Fprintf(out, "Date Updated: %s\n", details.DateUpdated)
	}
	if details.ErrorCode != "" {
		fmt.Fprintf(out, "Error Code: %s\n", details.ErrorCode)
	}
	if details.ErrorMessage != "" {
		fmt.Fprintf(out, "Error Message: %s\n", details.ErrorMessage)
	}
	if details.Body != "" {
		fmt.Fprintf(out, "Body: %s\n", details.Body)
	}
	if details.AccountSID != "" {
		fmt.Fprintf(out, "Account SID: %s\n", details.AccountSID)
	}
	if details.MessagingServiceSID != "" {
		fmt.Fprintf(out, "Messaging Service SID: %s\n", details.MessagingServiceSID)
	}
	if details.NumMedia != 0 {
		fmt.Fprintf(out, "Num Media: %d\n", details.NumMedia)
	}
	if details.NumSegments != 0 {
		fmt.Fprintf(out, "Num Segments: %d\n", details.NumSegments)
	}
	if details.Price != "" {
		fmt.Fprintf(out, "Price: %s\n", details.Price)
	}
	if details.PriceUnit != "" {
		fmt.Fprintf(out, "Price Unit: %s\n", details.PriceUnit)
	}
	if details.URI != "" {
		fmt.Fprintf(out, "URI: %s\n", details.URI)
	}
	if details.SubresourceURIs != nil {
		fmt.Fprintln(out, "Subresource URIs:")
		if err := writeJSON(out, details.SubresourceURIs); err != nil {
			return err
		}
	}
	if details.Webhooks != nil {
		fmt.Fprintln(out, "Webhooks:")
		if err := writeJSON(out, details.Webhooks); err != nil {
			return err
		}
	}

	return nil
}
