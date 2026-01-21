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

type doCertificationTypesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	CanApplyTo         string
	RequiresExpiration bool
	CanBeRequirementOf []string
	Broker             string
}

func newDoCertificationTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new certification type",
		Long: `Create a new certification type.

Required flags:
  --name           The certification type name (required)
  --can-apply-to   What entity type this applies to (required, e.g., Trucker, User)
  --broker         The broker ID (required)

Optional flags:
  --requires-expiration    Whether certifications of this type require an expiration date
  --can-be-requirement-of  Entity types this can be a requirement of (comma-separated)`,
		Example: `  # Create a basic certification type
  xbe do certification-types create --name "CDL Class A" --can-apply-to Trucker --broker 123

  # Create with expiration requirement
  xbe do certification-types create --name "OSHA 10" --can-apply-to User --requires-expiration --broker 123

  # Create with requirement-of types
  xbe do certification-types create --name "Hazmat" --can-apply-to Trucker --can-be-requirement-of Job,CustomerTender --broker 123

  # Get JSON output
  xbe do certification-types create --name "CDL Class A" --can-apply-to Trucker --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCertificationTypesCreate,
	}
	initDoCertificationTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doCertificationTypesCmd.AddCommand(newDoCertificationTypesCreateCmd())
}

func initDoCertificationTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Certification type name (required)")
	cmd.Flags().String("can-apply-to", "", "Entity type this applies to (required, e.g., Trucker, User)")
	cmd.Flags().Bool("requires-expiration", false, "Whether certifications require an expiration date")
	cmd.Flags().StringSlice("can-be-requirement-of", nil, "Entity types this can be a requirement of (comma-separated)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationTypesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationTypesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require can-apply-to
	if opts.CanApplyTo == "" {
		err := fmt.Errorf("--can-apply-to is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require broker
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name":         opts.Name,
		"can-apply-to": opts.CanApplyTo,
	}
	if cmd.Flags().Changed("requires-expiration") {
		attributes["requires-expiration"] = opts.RequiresExpiration
	}
	if len(opts.CanBeRequirementOf) > 0 {
		attributes["can-be-requirement-of"] = opts.CanBeRequirementOf
	}

	// Build request data with relationships
	data := map[string]any{
		"type":       "certification-types",
		"attributes": attributes,
		"relationships": map[string]any{
			"broker": map[string]any{
				"data": map[string]string{
					"type": "brokers",
					"id":   opts.Broker,
				},
			},
		},
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/certification-types", jsonBody)
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

	row := buildCertificationTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created certification type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCertificationTypesCreateOptions(cmd *cobra.Command) (doCertificationTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetString("can-apply-to")
	requiresExpiration, _ := cmd.Flags().GetBool("requires-expiration")
	canBeRequirementOf, _ := cmd.Flags().GetStringSlice("can-be-requirement-of")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationTypesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		CanApplyTo:         canApplyTo,
		RequiresExpiration: requiresExpiration,
		CanBeRequirementOf: canBeRequirementOf,
		Broker:             broker,
	}, nil
}

func buildCertificationTypeRowFromSingle(resp jsonAPISingleResponse) certificationTypeRow {
	attrs := resp.Data.Attributes

	row := certificationTypeRow{
		ID:                 resp.Data.ID,
		Name:               stringAttr(attrs, "name"),
		CanApplyTo:         stringAttr(attrs, "can-apply-to"),
		RequiresExpiration: boolAttr(attrs, "requires-expiration"),
		CanDelete:          boolAttr(attrs, "can-delete"),
		CanBeRequirementOf: strings.Join(stringSliceAttr(attrs, "can-be-requirement-of"), ", "),
	}

	// Get broker ID from relationships
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
