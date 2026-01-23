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

type doPublicPraisesUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	Description     string
	CultureValueIDs string
	RecipientIDs    string
}

func newDoPublicPraisesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a public praise",
		Long: `Update a public praise.

Optional flags:
  --description          Description of the praise
  --recipient-ids        Comma-separated list of recipient user IDs
  --culture-value-ids    Comma-separated list of culture value IDs`,
		Example: `  # Update description
  xbe do public-praises update 123 --description "Updated praise"

  # Update recipients
  xbe do public-praises update 123 --recipient-ids "456,789"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPublicPraisesUpdate,
	}
	initDoPublicPraisesUpdateFlags(cmd)
	return cmd
}

func init() {
	doPublicPraisesCmd.AddCommand(newDoPublicPraisesUpdateCmd())
}

func initDoPublicPraisesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description of the praise")
	cmd.Flags().String("culture-value-ids", "", "Comma-separated list of culture value IDs")
	cmd.Flags().String("recipient-ids", "", "Comma-separated list of recipient user IDs")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPublicPraisesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPublicPraisesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("culture-value-ids") {
		if opts.CultureValueIDs != "" {
			ids := strings.Split(opts.CultureValueIDs, ",")
			attributes["culture-value-ids"] = ids
		} else {
			attributes["culture-value-ids"] = []string{}
		}
	}
	if cmd.Flags().Changed("recipient-ids") {
		if opts.RecipientIDs != "" {
			ids := strings.Split(opts.RecipientIDs, ",")
			attributes["recipient-ids"] = ids
		} else {
			attributes["recipient-ids"] = []string{}
		}
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "public-praises",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/public-praises/"+opts.ID, jsonBody)
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

	row := buildPublicPraiseRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated public praise %s\n", row.ID)
	return nil
}

func parseDoPublicPraisesUpdateOptions(cmd *cobra.Command, args []string) (doPublicPraisesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	cultureValueIDs, _ := cmd.Flags().GetString("culture-value-ids")
	recipientIDs, _ := cmd.Flags().GetString("recipient-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPublicPraisesUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		Description:     description,
		CultureValueIDs: cultureValueIDs,
		RecipientIDs:    recipientIDs,
	}, nil
}
