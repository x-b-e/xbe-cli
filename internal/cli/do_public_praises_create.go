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

type doPublicPraisesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Description      string
	CultureValueIDs  string
	RecipientIDs     string
	GivenByID        string
	ReceivedByID     string
	OrganizationType string
	OrganizationID   string
}

func newDoPublicPraisesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new public praise",
		Long: `Create a new public praise.

Required flags:
  --description          Description of the praise (required)
  --given-by             User ID who is giving the praise (required)
  --received-by          User ID who is receiving the praise (required)
  --organization-type    Organization type (required)
  --organization-id      Organization ID (required)

Optional flags:
  --recipient-ids        Comma-separated list of additional recipient user IDs
  --culture-value-ids    Comma-separated list of culture value IDs`,
		Example: `  # Create a basic public praise
  xbe do public-praises create \
    --description "Great job on the project!" \
    --given-by 123 \
    --received-by 789 \
    --organization-type brokers \
    --organization-id 456

  # Create with culture values
  xbe do public-praises create \
    --description "Outstanding teamwork" \
    --given-by 123 \
    --received-by 789 \
    --organization-type brokers \
    --organization-id 456 \
    --culture-value-ids "1,2,3"`,
		Args: cobra.NoArgs,
		RunE: runDoPublicPraisesCreate,
	}
	initDoPublicPraisesCreateFlags(cmd)
	return cmd
}

func init() {
	doPublicPraisesCmd.AddCommand(newDoPublicPraisesCreateCmd())
}

func initDoPublicPraisesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description of the praise (required)")
	cmd.Flags().String("culture-value-ids", "", "Comma-separated list of culture value IDs")
	cmd.Flags().String("recipient-ids", "", "Comma-separated list of additional recipient user IDs")
	cmd.Flags().String("given-by", "", "User ID who is giving the praise (required)")
	cmd.Flags().String("received-by", "", "User ID who is receiving the praise (required)")
	cmd.Flags().String("organization-type", "", "Organization type (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPublicPraisesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPublicPraisesCreateOptions(cmd)
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

	if opts.Description == "" {
		err := fmt.Errorf("--description is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.GivenByID == "" {
		err := fmt.Errorf("--given-by is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ReceivedByID == "" {
		err := fmt.Errorf("--received-by is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"description": opts.Description,
	}

	if opts.CultureValueIDs != "" {
		ids := strings.Split(opts.CultureValueIDs, ",")
		attributes["culture-value-ids"] = ids
	}

	if opts.RecipientIDs != "" {
		ids := strings.Split(opts.RecipientIDs, ",")
		attributes["recipient-ids"] = ids
	}

	relationships := map[string]any{
		"given-by": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.GivenByID,
			},
		},
		"received-by": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ReceivedByID,
			},
		},
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "public-praises",
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

	body, _, err := client.Post(cmd.Context(), "/v1/public-praises", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created public praise %s\n", row.ID)
	return nil
}

func parseDoPublicPraisesCreateOptions(cmd *cobra.Command) (doPublicPraisesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	cultureValueIDs, _ := cmd.Flags().GetString("culture-value-ids")
	recipientIDs, _ := cmd.Flags().GetString("recipient-ids")
	givenByID, _ := cmd.Flags().GetString("given-by")
	receivedByID, _ := cmd.Flags().GetString("received-by")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPublicPraisesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Description:      description,
		CultureValueIDs:  cultureValueIDs,
		RecipientIDs:     recipientIDs,
		GivenByID:        givenByID,
		ReceivedByID:     receivedByID,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}

func buildPublicPraiseRowFromSingle(resp jsonAPISingleResponse) publicPraiseRow {
	attrs := resp.Data.Attributes

	row := publicPraiseRow{
		ID:          resp.Data.ID,
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["given-by"]; ok && rel.Data != nil {
		row.GivenByID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
