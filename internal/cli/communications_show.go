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

type communicationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type communicationDetails struct {
	ID                             string   `json:"id"`
	SubjectType                    string   `json:"subject_type,omitempty"`
	SubjectResourceType            string   `json:"subject_resource_type,omitempty"`
	SubjectID                      string   `json:"subject_id,omitempty"`
	MessageType                    string   `json:"message_type,omitempty"`
	MessageStatus                  string   `json:"message_status,omitempty"`
	MessageID                      string   `json:"message_id,omitempty"`
	MessageSentAt                  string   `json:"message_sent_at,omitempty"`
	MessageDeliveryStatusUpdatedAt string   `json:"message_delivery_status_updated_at,omitempty"`
	DeliveryStatus                 string   `json:"delivery_status,omitempty"`
	IsRetried                      bool     `json:"is_retried"`
	UserID                         string   `json:"user_id,omitempty"`
	ParentID                       string   `json:"parent_id,omitempty"`
	ChildIDs                       []string `json:"child_ids,omitempty"`
	CreatedAt                      string   `json:"created_at,omitempty"`
	UpdatedAt                      string   `json:"updated_at,omitempty"`
	Details                        any      `json:"details,omitempty"`
}

func newCommunicationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show communication details",
		Long: `Show the full details of a communication.

Output Fields:
  ID, message type, status, message identifiers
  Message timestamps and delivery status
  Subject and user references
  Parent/child communication references
  Created/updated timestamps
  Details payload

Arguments:
  <id>    The communication ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show communication details
  xbe view communications show 123

  # JSON output
  xbe view communications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommunicationsShow,
	}
	initCommunicationsShowFlags(cmd)
	return cmd
}

func init() {
	communicationsCmd.AddCommand(newCommunicationsShowCmd())
}

func initCommunicationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommunicationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCommunicationsShowOptions(cmd)
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
		return fmt.Errorf("communication id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[communications]", "subject-type,message-type,message-status,message-id,message-sent-at,message-delivery-status-updated-at,delivery-status,details,is-retried,created-at,updated-at,user,subject,parent,children")
	query.Set("include", "user,subject,parent,children")

	body, _, err := client.Get(cmd.Context(), "/v1/communications/"+id, query)
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

	details := buildCommunicationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommunicationDetails(cmd, details)
}

func parseCommunicationsShowOptions(cmd *cobra.Command) (communicationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return communicationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommunicationDetails(resp jsonAPISingleResponse) communicationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	subjectID, subjectType := relationshipRefFromMap(resource.Relationships, "subject")
	details := communicationDetails{
		ID:                             resource.ID,
		SubjectType:                    stringAttr(attrs, "subject-type"),
		SubjectResourceType:            subjectType,
		SubjectID:                      subjectID,
		MessageType:                    stringAttr(attrs, "message-type"),
		MessageStatus:                  stringAttr(attrs, "message-status"),
		MessageID:                      stringAttr(attrs, "message-id"),
		MessageSentAt:                  formatDateTime(stringAttr(attrs, "message-sent-at")),
		MessageDeliveryStatusUpdatedAt: formatDateTime(stringAttr(attrs, "message-delivery-status-updated-at")),
		DeliveryStatus:                 stringAttr(attrs, "delivery-status"),
		IsRetried:                      boolAttr(attrs, "is-retried"),
		UserID:                         relationshipIDFromMap(resource.Relationships, "user"),
		ParentID:                       relationshipIDFromMap(resource.Relationships, "parent"),
		ChildIDs:                       relationshipIDsFromMap(resource.Relationships, "children"),
		CreatedAt:                      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                      formatDateTime(stringAttr(attrs, "updated-at")),
		Details:                        anyAttr(attrs, "details"),
	}

	return details
}

func renderCommunicationDetails(cmd *cobra.Command, details communicationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MessageType != "" {
		fmt.Fprintf(out, "Message Type: %s\n", details.MessageType)
	}
	if details.MessageStatus != "" {
		fmt.Fprintf(out, "Message Status: %s\n", details.MessageStatus)
	}
	if details.MessageID != "" {
		fmt.Fprintf(out, "Message ID: %s\n", details.MessageID)
	}
	if details.MessageSentAt != "" {
		fmt.Fprintf(out, "Message Sent At: %s\n", details.MessageSentAt)
	}
	if details.MessageDeliveryStatusUpdatedAt != "" {
		fmt.Fprintf(out, "Delivery Status Updated At: %s\n", details.MessageDeliveryStatusUpdatedAt)
	}
	if details.DeliveryStatus != "" {
		fmt.Fprintf(out, "Delivery Status: %s\n", details.DeliveryStatus)
	}
	fmt.Fprintf(out, "Is Retried: %t\n", details.IsRetried)

	subjectLabel := formatRelationshipLabel(details.SubjectResourceType, details.SubjectID)
	if subjectLabel != "" {
		fmt.Fprintf(out, "Subject: %s\n", subjectLabel)
	}
	if details.SubjectType != "" {
		fmt.Fprintf(out, "Subject Type: %s\n", details.SubjectType)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.ParentID != "" {
		fmt.Fprintf(out, "Parent ID: %s\n", details.ParentID)
	}
	if len(details.ChildIDs) > 0 {
		fmt.Fprintf(out, "Child IDs: %s\n", strings.Join(details.ChildIDs, ", "))
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
