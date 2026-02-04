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

type brokerProjectTransportEventTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerProjectTransportEventTypeDetails struct {
	ID                            string `json:"id"`
	Code                          string `json:"code,omitempty"`
	BrokerID                      string `json:"broker_id,omitempty"`
	BrokerName                    string `json:"broker_name,omitempty"`
	ProjectTransportEventTypeID   string `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventType     string `json:"project_transport_event_type,omitempty"`
	ProjectTransportEventTypeCode string `json:"project_transport_event_type_code,omitempty"`
}

func newBrokerProjectTransportEventTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker project transport event type details",
		Long: `Show the full details of a broker project transport event type.

Output Fields:
  ID                         Broker project transport event type identifier
  Code                       Broker-specific event type code
  Broker                     Broker name
  Project Transport Event Type Event type name
  Event Type Code            Event type code

Arguments:
  <id>    Broker project transport event type ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a broker project transport event type
  xbe view broker-project-transport-event-types show 123

  # JSON output
  xbe view broker-project-transport-event-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerProjectTransportEventTypesShow,
	}
	initBrokerProjectTransportEventTypesShowFlags(cmd)
	return cmd
}

func init() {
	brokerProjectTransportEventTypesCmd.AddCommand(newBrokerProjectTransportEventTypesShowCmd())
}

func initBrokerProjectTransportEventTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerProjectTransportEventTypesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBrokerProjectTransportEventTypesShowOptions(cmd)
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
		return fmt.Errorf("broker project transport event type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-project-transport-event-types]", "code,broker,project-transport-event-type")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-transport-event-types]", "code,name")
	query.Set("include", "broker,project-transport-event-type")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-project-transport-event-types/"+id, query)
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

	details := buildBrokerProjectTransportEventTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerProjectTransportEventTypeDetails(cmd, details)
}

func parseBrokerProjectTransportEventTypesShowOptions(cmd *cobra.Command) (brokerProjectTransportEventTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerProjectTransportEventTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerProjectTransportEventTypeDetails(resp jsonAPISingleResponse) brokerProjectTransportEventTypeDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := brokerProjectTransportEventTypeDetails{
		ID:   resp.Data.ID,
		Code: stringAttr(resp.Data.Attributes, "code"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		details.ProjectTransportEventTypeID = rel.Data.ID
		if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportEventType = strings.TrimSpace(stringAttr(eventType.Attributes, "name"))
			details.ProjectTransportEventTypeCode = strings.TrimSpace(stringAttr(eventType.Attributes, "code"))
		}
	}

	return details
}

func renderBrokerProjectTransportEventTypeDetails(cmd *cobra.Command, details brokerProjectTransportEventTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Code != "" {
		fmt.Fprintf(out, "Code: %s\n", details.Code)
	}

	writeLabelWithID(out, "Broker", firstNonEmpty(details.BrokerName), details.BrokerID)

	eventTypeDisplay := firstNonEmpty(details.ProjectTransportEventType, details.ProjectTransportEventTypeCode)
	writeLabelWithID(out, "Project Transport Event Type", eventTypeDisplay, details.ProjectTransportEventTypeID)
	if details.ProjectTransportEventType != "" && details.ProjectTransportEventTypeCode != "" {
		fmt.Fprintf(out, "Event Type Code: %s\n", details.ProjectTransportEventTypeCode)
	}

	return nil
}
