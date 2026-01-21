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

type doCertificationTypesUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	CanApplyTo         string
	RequiresExpiration bool
	CanBeRequirementOf []string
}

func newDoCertificationTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a certification type",
		Long: `Update an existing certification type.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The certification type ID (required)

Flags:
  --name                   Update the name
  --can-apply-to           Update what entity type this applies to
  --requires-expiration    Update whether certifications require expiration
  --can-be-requirement-of  Update entity types this can be a requirement of`,
		Example: `  # Update just the name
  xbe do certification-types update 456 --name "CDL Class A - Updated"

  # Update expiration requirement
  xbe do certification-types update 456 --requires-expiration

  # Update requirement-of types
  xbe do certification-types update 456 --can-be-requirement-of Job,BrokerTender

  # Get JSON output
  xbe do certification-types update 456 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCertificationTypesUpdate,
	}
	initDoCertificationTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCertificationTypesCmd.AddCommand(newDoCertificationTypesUpdateCmd())
}

func initDoCertificationTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("can-apply-to", "", "New entity type this applies to")
	cmd.Flags().Bool("requires-expiration", false, "Whether certifications require expiration")
	cmd.Flags().StringSlice("can-be-requirement-of", nil, "Entity types this can be a requirement of")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationTypesUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("certification type id is required")
	}

	// Check if at least one field is being updated
	hasUpdate := opts.Name != "" || opts.CanApplyTo != "" ||
		cmd.Flags().Changed("requires-expiration") ||
		cmd.Flags().Changed("can-be-requirement-of")

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.CanApplyTo != "" {
		attributes["can-apply-to"] = opts.CanApplyTo
	}
	if cmd.Flags().Changed("requires-expiration") {
		attributes["requires-expiration"] = opts.RequiresExpiration
	}
	if cmd.Flags().Changed("can-be-requirement-of") {
		attributes["can-be-requirement-of"] = opts.CanBeRequirementOf
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "certification-types",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/certification-types/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated certification type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCertificationTypesUpdateOptions(cmd *cobra.Command) (doCertificationTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetString("can-apply-to")
	requiresExpiration, _ := cmd.Flags().GetBool("requires-expiration")
	canBeRequirementOf, _ := cmd.Flags().GetStringSlice("can-be-requirement-of")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationTypesUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		CanApplyTo:         canApplyTo,
		RequiresExpiration: requiresExpiration,
		CanBeRequirementOf: canBeRequirementOf,
	}, nil
}
