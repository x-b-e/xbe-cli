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

type doTractorTrailerCredentialClassificationsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Name        string
	Description string
	IssuerName  string
	ExternalID  string
}

func newDoTractorTrailerCredentialClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing tractor/trailer credential classification",
		Long: `Update an existing tractor/trailer credential classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name         The classification name
  --description  Classification description
  --issuer-name  Issuing authority name
  --external-id  External identifier`,
		Example: `  # Update name
  xbe do tractor-trailer-credential-classifications update 123 --name "Updated Name"

  # Update multiple fields
  xbe do tractor-trailer-credential-classifications update 123 --name "New Name" --description "New description"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorTrailerCredentialClassificationsUpdate,
	}
	initDoTractorTrailerCredentialClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTractorTrailerCredentialClassificationsCmd.AddCommand(newDoTractorTrailerCredentialClassificationsUpdateCmd())
}

func initDoTractorTrailerCredentialClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name")
	cmd.Flags().String("description", "", "Classification description")
	cmd.Flags().String("issuer-name", "", "Issuing authority name")
	cmd.Flags().String("external-id", "", "External identifier")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorTrailerCredentialClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorTrailerCredentialClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("issuer-name") {
		attributes["issuer-name"] = opts.IssuerName
	}
	if cmd.Flags().Changed("external-id") {
		attributes["external-id"] = opts.ExternalID
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --description, --issuer-name, --external-id")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tractor-trailer-credential-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/tractor-trailer-credential-classifications/"+opts.ID, jsonBody)
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

	row := buildTractorTrailerCredentialClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tractor/trailer credential classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTractorTrailerCredentialClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doTractorTrailerCredentialClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	issuerName, _ := cmd.Flags().GetString("issuer-name")
	externalID, _ := cmd.Flags().GetString("external-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorTrailerCredentialClassificationsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Name:        name,
		Description: description,
		IssuerName:  issuerName,
		ExternalID:  externalID,
	}, nil
}
