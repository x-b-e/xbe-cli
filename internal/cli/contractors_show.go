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

type contractorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type contractorDetails struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BrokerID    string   `json:"broker_id,omitempty"`
	BrokerName  string   `json:"broker_name,omitempty"`
	IncidentIDs []string `json:"incident_ids,omitempty"`
}

func newContractorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show contractor details",
		Long: `Show the full details of a contractor.

Output Fields:
  ID           Contractor identifier
  Name         Contractor name
  Broker       Broker organization
  Broker ID    Broker identifier
  Incident IDs Related incident IDs

Arguments:
  <id>  Contractor ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a contractor
  xbe view contractors show 123

  # JSON output
  xbe view contractors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runContractorsShow,
	}
	initContractorsShowFlags(cmd)
	return cmd
}

func init() {
	contractorsCmd.AddCommand(newContractorsShowCmd())
}

func initContractorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runContractorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseContractorsShowOptions(cmd)
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
		return fmt.Errorf("contractor id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[contractors]", "name,broker,incidents")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	body, _, err := client.Get(cmd.Context(), "/v1/contractors/"+id, query)
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

	details := buildContractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderContractorDetails(cmd, details)
}

func parseContractorsShowOptions(cmd *cobra.Command) (contractorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return contractorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildContractorDetails(resp jsonAPISingleResponse) contractorDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := contractorDetails{
		ID:   resp.Data.ID,
		Name: strings.TrimSpace(stringAttr(attrs, "name")),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["incidents"]; ok {
		details.IncidentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderContractorDetails(cmd *cobra.Command, details contractorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if len(details.IncidentIDs) > 0 {
		fmt.Fprintf(out, "Incident IDs: %s\n", strings.Join(details.IncidentIDs, ", "))
	}

	return nil
}
