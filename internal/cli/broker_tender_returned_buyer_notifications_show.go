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

type brokerTenderReturnedBuyerNotificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerTenderReturnedBuyerNotificationDetails struct {
	ID                       string `json:"id"`
	TenderID                 string `json:"tender_id,omitempty"`
	UserID                   string `json:"user_id,omitempty"`
	Read                     bool   `json:"read"`
	NotificationType         string `json:"notification_type,omitempty"`
	DeliveryDecisionApproach string `json:"delivery_decision_approach,omitempty"`
	IsReadyForDelivery       bool   `json:"is_ready_for_delivery"`
	DeliverAt                string `json:"deliver_at,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
	Details                  any    `json:"details,omitempty"`
}

func newBrokerTenderReturnedBuyerNotificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker tender returned buyer notification details",
		Long: `Show the full details of a broker tender returned buyer notification.

Output Fields:
  ID, tender, and user identifiers
  Read status and delivery decision metadata
  Delivery timestamps and notification details

Arguments:
  <id>    The notification ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a notification
  xbe view broker-tender-returned-buyer-notifications show 123

  # JSON output
  xbe view broker-tender-returned-buyer-notifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerTenderReturnedBuyerNotificationsShow,
	}
	initBrokerTenderReturnedBuyerNotificationsShowFlags(cmd)
	return cmd
}

func init() {
	brokerTenderReturnedBuyerNotificationsCmd.AddCommand(newBrokerTenderReturnedBuyerNotificationsShowCmd())
}

func initBrokerTenderReturnedBuyerNotificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerTenderReturnedBuyerNotificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerTenderReturnedBuyerNotificationsShowOptions(cmd)
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
		return fmt.Errorf("notification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-tender-returned-buyer-notifications]", strings.Join([]string{
		"read",
		"details",
		"notification-type",
		"deliver-at",
		"delivery-decision-approach",
		"is-ready-for-delivery",
		"created-at",
		"updated-at",
		"user",
		"tender",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/broker-tender-returned-buyer-notifications/"+id, query)
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

	details := buildBrokerTenderReturnedBuyerNotificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerTenderReturnedBuyerNotificationDetails(cmd, details)
}

func parseBrokerTenderReturnedBuyerNotificationsShowOptions(cmd *cobra.Command) (brokerTenderReturnedBuyerNotificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerTenderReturnedBuyerNotificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerTenderReturnedBuyerNotificationDetails(resp jsonAPISingleResponse) brokerTenderReturnedBuyerNotificationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return brokerTenderReturnedBuyerNotificationDetails{
		ID:                       resource.ID,
		TenderID:                 relationshipIDFromMap(resource.Relationships, "tender"),
		UserID:                   relationshipIDFromMap(resource.Relationships, "user"),
		Read:                     boolAttr(attrs, "read"),
		NotificationType:         stringAttr(attrs, "notification-type"),
		DeliveryDecisionApproach: stringAttr(attrs, "delivery-decision-approach"),
		IsReadyForDelivery:       boolAttr(attrs, "is-ready-for-delivery"),
		DeliverAt:                formatDateTime(stringAttr(attrs, "deliver-at")),
		CreatedAt:                formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                formatDateTime(stringAttr(attrs, "updated-at")),
		Details:                  anyAttr(attrs, "details"),
	}
}

func renderBrokerTenderReturnedBuyerNotificationDetails(cmd *cobra.Command, details brokerTenderReturnedBuyerNotificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderID != "" {
		fmt.Fprintf(out, "Tender ID: %s\n", details.TenderID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	fmt.Fprintf(out, "Read: %t\n", details.Read)
	if details.NotificationType != "" {
		fmt.Fprintf(out, "Notification Type: %s\n", details.NotificationType)
	}
	if details.DeliveryDecisionApproach != "" {
		fmt.Fprintf(out, "Delivery Decision Approach: %s\n", details.DeliveryDecisionApproach)
	}
	fmt.Fprintf(out, "Is Ready For Delivery: %t\n", details.IsReadyForDelivery)
	if details.DeliverAt != "" {
		fmt.Fprintf(out, "Deliver At: %s\n", details.DeliverAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Details != nil {
		fmt.Fprintln(out, "Details:")
		if err := writeJSON(out, details.Details); err != nil {
			return err
		}
	}

	return nil
}
