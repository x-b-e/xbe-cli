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

type doIncidentSubscriptionsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	Customer         string
	Broker           string
	MaterialSupplier string
	Kind             string
	ContactMethod    string
	IncidentType     string
	IncidentID       string
}

func newDoIncidentSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an incident subscription",
		Long: `Update an incident subscription.

Optional flags:
  --kind              Update incident kind filter
  --contact-method    Update contact method (email_address, mobile_number)
  --broker            Update broker scope
  --customer          Update customer scope
  --material-supplier Update material supplier scope
  --incident-type     Update incident type (requires --incident-id)
  --incident-id       Update incident ID (requires --incident-type)

Notes:
  If an incident scope is provided, organization scope and kind must be omitted.`,
		Example: `  # Update contact method
  xbe do incident-subscriptions update 123 --contact-method mobile_number

  # Update kind
  xbe do incident-subscriptions update 123 --kind safety`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentSubscriptionsUpdate,
	}
	initDoIncidentSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doIncidentSubscriptionsCmd.AddCommand(newDoIncidentSubscriptionsUpdateCmd())
}

func initDoIncidentSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("kind", "", "Incident kind filter")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("material-supplier", "", "Material supplier ID")
	cmd.Flags().String("incident-type", "", "Incident type (e.g., safety-incidents)")
	cmd.Flags().String("incident-id", "", "Incident ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentSubscriptionsUpdateOptions(cmd, args)
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

	if (opts.IncidentType == "") != (opts.IncidentID == "") {
		err := fmt.Errorf("--incident-type and --incident-id must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	scopeCount := 0
	if opts.Broker != "" {
		scopeCount++
	}
	if opts.Customer != "" {
		scopeCount++
	}
	if opts.MaterialSupplier != "" {
		scopeCount++
	}
	if scopeCount > 1 {
		err := fmt.Errorf("only one of --broker, --customer, or --material-supplier may be specified")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.IncidentType != "" && (opts.Kind != "" || scopeCount > 0) {
		err := fmt.Errorf("incident-scoped subscriptions cannot include organization scope or kind")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("contact-method") {
		attributes["contact-method"] = opts.ContactMethod
	}

	relationships := map[string]any{}
	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if opts.Customer != "" {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		}
	}
	if opts.MaterialSupplier != "" {
		relationships["material-supplier"] = map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.MaterialSupplier,
			},
		}
	}
	if opts.IncidentType != "" {
		relationships["incident"] = map[string]any{
			"data": map[string]any{
				"type": opts.IncidentType,
				"id":   opts.IncidentID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "incident-subscriptions",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/incident-subscriptions/"+opts.ID, jsonBody)
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

	row := buildIncidentSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated incident subscription %s\n", row.ID)
	return nil
}

func parseDoIncidentSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doIncidentSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	kind, _ := cmd.Flags().GetString("kind")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentSubscriptionsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		Kind:             kind,
		ContactMethod:    contactMethod,
		Broker:           broker,
		Customer:         customer,
		MaterialSupplier: materialSupplier,
		IncidentType:     incidentType,
		IncidentID:       incidentID,
	}, nil
}
