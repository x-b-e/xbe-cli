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

type doIncidentSubscriptionsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	User             string
	Customer         string
	Broker           string
	MaterialSupplier string
	Kind             string
	ContactMethod    string
	IncidentType     string
	IncidentID       string
}

func newDoIncidentSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new incident subscription",
		Long: `Create a new incident subscription.

Required flags:
  --user  User ID (required)

Scope flags (at least one required for non-admin users):
  --broker             Broker ID
  --customer           Customer ID
  --material-supplier  Material supplier ID
  --incident-type      Incident type (e.g., safety-incidents)
  --incident-id        Incident ID

Optional flags:
  --kind            Incident kind filter (e.g., safety, equipment)
  --contact-method  Contact method (email_address, mobile_number)

Notes:
  If an incident scope is provided, organization scope and kind must be omitted.`,
		Example: `  # Subscribe a user to broker incidents by kind
  xbe do incident-subscriptions create --user 123 --broker 456 --kind safety

  # Subscribe to a specific incident
  xbe do incident-subscriptions create --user 123 --incident-type safety-incidents --incident-id 789

  # Create with explicit contact method
  xbe do incident-subscriptions create --user 123 --broker 456 --contact-method email_address --json`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentSubscriptionsCreate,
	}
	initDoIncidentSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentSubscriptionsCmd.AddCommand(newDoIncidentSubscriptionsCreateCmd())
}

func initDoIncidentSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("material-supplier", "", "Material supplier ID")
	cmd.Flags().String("kind", "", "Incident kind filter")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("incident-type", "", "Incident type (e.g., safety-incidents)")
	cmd.Flags().String("incident-id", "", "Incident ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("user")
}

func runDoIncidentSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentSubscriptionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}
	if opts.ContactMethod != "" {
		attributes["contact-method"] = opts.ContactMethod
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-subscriptions",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/incident-subscriptions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident subscription %s\n", row.ID)
	return nil
}

func parseDoIncidentSubscriptionsCreateOptions(cmd *cobra.Command) (doIncidentSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	kind, _ := cmd.Flags().GetString("kind")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentSubscriptionsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		User:             user,
		Broker:           broker,
		Customer:         customer,
		MaterialSupplier: materialSupplier,
		Kind:             kind,
		ContactMethod:    contactMethod,
		IncidentType:     incidentType,
		IncidentID:       incidentID,
	}, nil
}
