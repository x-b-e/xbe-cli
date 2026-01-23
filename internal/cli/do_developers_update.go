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

type doDevelopersUpdateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	Name                            string
	WeigherSealLabel                string
	IsPrevailingWageExplicit        bool
	IsCertificationRequiredExplicit bool
}

func newDoDevelopersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a developer",
		Long: `Update an existing developer.

Specify the developer ID and the fields to update.`,
		Example: `  # Update developer name
  xbe do developers update 123 --name "New Name"

  # Update prevailing wage setting
  xbe do developers update 123 --is-prevailing-wage-explicit

  # Update multiple fields
  xbe do developers update 123 --name "New Name" --weigher-seal-label "NEW"

  # Get JSON output
  xbe do developers update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDevelopersUpdate,
	}
	initDoDevelopersUpdateFlags(cmd)
	return cmd
}

func init() {
	doDevelopersCmd.AddCommand(newDoDevelopersUpdateCmd())
}

func initDoDevelopersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Developer name")
	cmd.Flags().String("weigher-seal-label", "", "Weigher seal label")
	cmd.Flags().Bool("is-prevailing-wage-explicit", false, "Explicitly set prevailing wage requirement")
	cmd.Flags().Bool("is-certification-required-explicit", false, "Explicitly set certification requirement")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDevelopersUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	opts, err := parseDoDevelopersUpdateOptions(cmd)
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

	// Build attributes - only include changed fields
	attributes := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
		hasChanges = true
	}
	if cmd.Flags().Changed("weigher-seal-label") {
		attributes["weigher-seal-label"] = opts.WeigherSealLabel
		hasChanges = true
	}
	if cmd.Flags().Changed("is-prevailing-wage-explicit") {
		attributes["is-prevailing-wage-explicit"] = opts.IsPrevailingWageExplicit
		hasChanges = true
	}
	if cmd.Flags().Changed("is-certification-required-explicit") {
		attributes["is-certification-required-explicit"] = opts.IsCertificationRequiredExplicit
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update - specify at least one field to change")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "developers",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/developers/"+id, jsonBody)
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

	row := developerRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDevelopersUpdateOptions(cmd *cobra.Command) (doDevelopersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	weigherSealLabel, _ := cmd.Flags().GetString("weigher-seal-label")
	isPrevailingWageExplicit, _ := cmd.Flags().GetBool("is-prevailing-wage-explicit")
	isCertificationRequiredExplicit, _ := cmd.Flags().GetBool("is-certification-required-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDevelopersUpdateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		Name:                            name,
		WeigherSealLabel:                weigherSealLabel,
		IsPrevailingWageExplicit:        isPrevailingWageExplicit,
		IsCertificationRequiredExplicit: isCertificationRequiredExplicit,
	}, nil
}
