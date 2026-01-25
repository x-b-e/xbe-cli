package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type communicationsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	User             string
	SubjectType      string
	SubjectID        string
	DeliveryStatus   string
	MessageType      string
	MessageID        string
	MessageSentAtMin string
	MessageSentAtMax string
	IsMessageSentAt  string
	IsAddressed      string
	IsRetried        string
	CreatedAtMin     string
	CreatedAtMax     string
	UpdatedAtMin     string
	UpdatedAtMax     string
	IsCreatedAt      string
	IsUpdatedAt      string
}

type communicationRow struct {
	ID             string `json:"id"`
	MessageType    string `json:"message_type,omitempty"`
	MessageStatus  string `json:"message_status,omitempty"`
	DeliveryStatus string `json:"delivery_status,omitempty"`
	MessageSentAt  string `json:"message_sent_at,omitempty"`
	IsRetried      bool   `json:"is_retried"`
	UserID         string `json:"user_id,omitempty"`
	SubjectType    string `json:"subject_type,omitempty"`
	SubjectID      string `json:"subject_id,omitempty"`
}

func newCommunicationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List communications",
		Long: `List communications with filtering and pagination.

Output Columns:
  ID              Communication identifier
  TYPE            Message type (e.g., TextMessage, Email)
  DELIVERY        Delivery status
  MESSAGE STATUS  Provider message status
  SENT AT         When the message was sent
  RETRIED         Whether the message was retried
  USER            User ID
  SUBJECT         Subject type and ID

Filters:
  --user                   Filter by user ID
  --subject-type           Filter by subject type
  --subject-id             Filter by subject ID (requires --subject-type)
  --delivery-status        Filter by delivery status
  --message-type           Filter by message type
  --message-id             Filter by message ID
  --message-sent-at-min    Filter by message-sent-at on/after (ISO 8601)
  --message-sent-at-max    Filter by message-sent-at on/before (ISO 8601)
  --is-message-sent-at     Filter by presence of message-sent-at (true/false)
  --is-addressed           Filter by addressed status (true/false)
  --is-retried             Filter by retried status (true/false)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-created-at          Filter by presence of created-at (true/false)
  --is-updated-at          Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List communications
  xbe view communications list

  # Filter by subject
  xbe view communications list --subject-type Project --subject-id 123

  # Filter by delivery status
  xbe view communications list --delivery-status incoming_received

  # Output as JSON
  xbe view communications list --json`,
		Args: cobra.NoArgs,
		RunE: runCommunicationsList,
	}
	initCommunicationsListFlags(cmd)
	return cmd
}

func init() {
	communicationsCmd.AddCommand(newCommunicationsListCmd())
}

func initCommunicationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("subject-type", "", "Filter by subject type")
	cmd.Flags().String("subject-id", "", "Filter by subject ID (requires --subject-type)")
	cmd.Flags().String("delivery-status", "", "Filter by delivery status")
	cmd.Flags().String("message-type", "", "Filter by message type")
	cmd.Flags().String("message-id", "", "Filter by message ID")
	cmd.Flags().String("message-sent-at-min", "", "Filter by message-sent-at on/after (ISO 8601)")
	cmd.Flags().String("message-sent-at-max", "", "Filter by message-sent-at on/before (ISO 8601)")
	cmd.Flags().String("is-message-sent-at", "", "Filter by presence of message-sent-at (true/false)")
	cmd.Flags().String("is-addressed", "", "Filter by addressed status (true/false)")
	cmd.Flags().String("is-retried", "", "Filter by retried status (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommunicationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommunicationsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[communications]", "message-type,message-status,delivery-status,message-sent-at,is-retried,subject-type,user,subject")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user]", opts.User)

	if opts.SubjectType != "" && opts.SubjectID != "" {
		query.Set("filter[subject]", opts.SubjectType+"|"+opts.SubjectID)
	} else {
		setFilterIfPresent(query, "filter[subject-type]", opts.SubjectType)
	}

	setFilterIfPresent(query, "filter[delivery-status]", opts.DeliveryStatus)
	setFilterIfPresent(query, "filter[message-type]", opts.MessageType)
	setFilterIfPresent(query, "filter[message-id]", opts.MessageID)
	setFilterIfPresent(query, "filter[message-sent-at-min]", opts.MessageSentAtMin)
	setFilterIfPresent(query, "filter[message-sent-at-max]", opts.MessageSentAtMax)
	setFilterIfPresent(query, "filter[is-message-sent-at]", opts.IsMessageSentAt)
	setFilterIfPresent(query, "filter[is-addressed]", opts.IsAddressed)
	setFilterIfPresent(query, "filter[is-retried]", opts.IsRetried)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/communications", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildCommunicationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommunicationsTable(cmd, rows)
}

func parseCommunicationsListOptions(cmd *cobra.Command) (communicationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	deliveryStatus, _ := cmd.Flags().GetString("delivery-status")
	messageType, _ := cmd.Flags().GetString("message-type")
	messageID, _ := cmd.Flags().GetString("message-id")
	messageSentAtMin, _ := cmd.Flags().GetString("message-sent-at-min")
	messageSentAtMax, _ := cmd.Flags().GetString("message-sent-at-max")
	isMessageSentAt, _ := cmd.Flags().GetString("is-message-sent-at")
	isAddressed, _ := cmd.Flags().GetString("is-addressed")
	isRetried, _ := cmd.Flags().GetString("is-retried")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return communicationsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		User:             user,
		SubjectType:      subjectType,
		SubjectID:        subjectID,
		DeliveryStatus:   deliveryStatus,
		MessageType:      messageType,
		MessageID:        messageID,
		MessageSentAtMin: messageSentAtMin,
		MessageSentAtMax: messageSentAtMax,
		IsMessageSentAt:  isMessageSentAt,
		IsAddressed:      isAddressed,
		IsRetried:        isRetried,
		CreatedAtMin:     createdAtMin,
		CreatedAtMax:     createdAtMax,
		UpdatedAtMin:     updatedAtMin,
		UpdatedAtMax:     updatedAtMax,
		IsCreatedAt:      isCreatedAt,
		IsUpdatedAt:      isUpdatedAt,
	}, nil
}

func buildCommunicationRows(resp jsonAPIResponse) []communicationRow {
	rows := make([]communicationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		subjectID, subjectType := relationshipRefFromMap(resource.Relationships, "subject")
		subjectTypeAttr := stringAttr(attrs, "subject-type")
		if subjectTypeAttr != "" {
			subjectType = subjectTypeAttr
		}
		row := communicationRow{
			ID:             resource.ID,
			MessageType:    stringAttr(attrs, "message-type"),
			MessageStatus:  stringAttr(attrs, "message-status"),
			DeliveryStatus: stringAttr(attrs, "delivery-status"),
			MessageSentAt:  formatDateTime(stringAttr(attrs, "message-sent-at")),
			IsRetried:      boolAttr(attrs, "is-retried"),
			UserID:         relationshipIDFromMap(resource.Relationships, "user"),
			SubjectType:    subjectType,
			SubjectID:      subjectID,
		}
		rows = append(rows, row)
	}
	return rows
}

func renderCommunicationsTable(cmd *cobra.Command, rows []communicationRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tDELIVERY\tMESSAGE STATUS\tSENT AT\tRETRIED\tUSER\tSUBJECT")
	for _, row := range rows {
		retried := "no"
		if row.IsRetried {
			retried = "yes"
		}
		subjectLabel := formatRelationshipLabel(row.SubjectType, row.SubjectID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.MessageType, 16),
			truncateString(row.DeliveryStatus, 22),
			truncateString(row.MessageStatus, 22),
			truncateString(row.MessageSentAt, 20),
			retried,
			truncateString(row.UserID, 14),
			truncateString(subjectLabel, 28),
		)
	}
	return writer.Flush()
}
