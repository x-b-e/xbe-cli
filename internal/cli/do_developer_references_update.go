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

type doDeveloperReferencesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Value       string
	SubjectType string
	SubjectID   string
}

func newDoDeveloperReferencesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a developer reference",
		Long: `Update a developer reference.

Optional:
  --value         Reference value
  --subject-type  Subject type (must be used with --subject-id)
  --subject-id    Subject ID (must be used with --subject-type)`,
		Example: `  # Update value
  xbe do developer-references update 123 --value "EXT-67890"

  # Update subject
  xbe do developer-references update 123 --subject-type customers --subject-id 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperReferencesUpdate,
	}
	initDoDeveloperReferencesUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperReferencesCmd.AddCommand(newDoDeveloperReferencesUpdateCmd())
}

func initDoDeveloperReferencesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("value", "", "Reference value")
	cmd.Flags().String("subject-type", "", "Subject type")
	cmd.Flags().String("subject-id", "", "Subject ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperReferencesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperReferencesUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("value") {
		attributes["value"] = opts.Value
	}

	if cmd.Flags().Changed("subject-type") && cmd.Flags().Changed("subject-id") {
		relationships["subject"] = map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "developer-references",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-references/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := developerReferenceRow{
			ID:    resp.Data.ID,
			Value: stringAttr(resp.Data.Attributes, "value"),
		}
		if rel, ok := resp.Data.Relationships["developer-reference-type"]; ok && rel.Data != nil {
			row.DeveloperReferenceTypeID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer reference %s\n", resp.Data.ID)
	return nil
}

func parseDoDeveloperReferencesUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperReferencesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	value, _ := cmd.Flags().GetString("value")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperReferencesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Value:       value,
		SubjectType: subjectType,
		SubjectID:   subjectID,
	}, nil
}
