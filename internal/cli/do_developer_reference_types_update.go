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

type doDeveloperReferenceTypesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	SubjectTypes []string
}

func newDoDeveloperReferenceTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing developer reference type",
		Long: `Update an existing developer reference type.

Provide the reference type ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name          The reference type name
  --subject-types Subject types (comma-separated or repeated)`,
		Example: `  # Update name
  xbe do developer-reference-types update 123 --name "Updated Name"

  # Update subject types
  xbe do developer-reference-types update 123 --subject-types "Job,Project,Invoice"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperReferenceTypesUpdate,
	}
	initDoDeveloperReferenceTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperReferenceTypesCmd.AddCommand(newDoDeveloperReferenceTypesUpdateCmd())
}

func initDoDeveloperReferenceTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Reference type name")
	cmd.Flags().StringSlice("subject-types", nil, "Subject types (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperReferenceTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperReferenceTypesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("subject-types") {
		attributes["subject-types"] = opts.SubjectTypes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --subject-types")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "developer-reference-types",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-reference-types/"+opts.ID, jsonBody)
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

	row := buildDeveloperReferenceTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer reference type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDeveloperReferenceTypesUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperReferenceTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	subjectTypes, _ := cmd.Flags().GetStringSlice("subject-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperReferenceTypesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Name:         name,
		SubjectTypes: subjectTypes,
	}, nil
}
