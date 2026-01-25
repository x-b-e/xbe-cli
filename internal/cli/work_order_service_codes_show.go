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

type workOrderServiceCodesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type workOrderServiceCodeDetails struct {
	ID          string `json:"id"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	BrokerName  string `json:"broker_name,omitempty"`
}

func newWorkOrderServiceCodesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show work order service code details",
		Long: `Show the full details of a work order service code.

Output Fields:
  ID
  Code
  Description
  Broker (name)
  Broker ID

Arguments:
  <id>    The work order service code ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a work order service code
  xbe view work-order-service-codes show 123

  # Output as JSON
  xbe view work-order-service-codes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runWorkOrderServiceCodesShow,
	}
	initWorkOrderServiceCodesShowFlags(cmd)
	return cmd
}

func init() {
	workOrderServiceCodesCmd.AddCommand(newWorkOrderServiceCodesShowCmd())
}

func initWorkOrderServiceCodesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrderServiceCodesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseWorkOrderServiceCodesShowOptions(cmd)
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
		return fmt.Errorf("work order service code id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[work-order-service-codes]", "code,description,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/work-order-service-codes/"+id, query)
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

	details := buildWorkOrderServiceCodeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderWorkOrderServiceCodeDetails(cmd, details)
}

func parseWorkOrderServiceCodesShowOptions(cmd *cobra.Command) (workOrderServiceCodesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrderServiceCodesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildWorkOrderServiceCodeDetails(resp jsonAPISingleResponse) workOrderServiceCodeDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := workOrderServiceCodeDetails{
		ID:          resource.ID,
		Code:        stringAttr(attrs, "code"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	return details
}

func renderWorkOrderServiceCodeDetails(cmd *cobra.Command, details workOrderServiceCodeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Code != "" {
		fmt.Fprintf(out, "Code: %s\n", details.Code)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
