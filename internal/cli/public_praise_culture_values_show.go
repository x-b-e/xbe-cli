package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type publicPraiseCultureValuesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type publicPraiseCultureValueDetails struct {
	ID             string `json:"id"`
	PublicPraiseID string `json:"public_praise_id,omitempty"`
	CultureValueID string `json:"culture_value_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

func newPublicPraiseCultureValuesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show public praise culture value details",
		Long: `Show the full details of a public praise culture value.

Output Fields:
  ID
  Public Praise
  Culture Value
  Created At
  Updated At

Arguments:
  <id>    The public praise culture value ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a public praise culture value
  xbe view public-praise-culture-values show 123

  # Get JSON output
  xbe view public-praise-culture-values show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPublicPraiseCultureValuesShow,
	}
	initPublicPraiseCultureValuesShowFlags(cmd)
	return cmd
}

func init() {
	publicPraiseCultureValuesCmd.AddCommand(newPublicPraiseCultureValuesShowCmd())
}

func initPublicPraiseCultureValuesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPublicPraiseCultureValuesShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePublicPraiseCultureValuesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("public praise culture value id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[public-praise-culture-values]", "created-at,updated-at,public-praise,culture-value")

	body, _, err := client.Get(cmd.Context(), "/v1/public-praise-culture-values/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPublicPraiseCultureValueDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPublicPraiseCultureValueDetails(cmd, details)
}

func parsePublicPraiseCultureValuesShowOptions(cmd *cobra.Command) (publicPraiseCultureValuesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return publicPraiseCultureValuesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return publicPraiseCultureValuesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return publicPraiseCultureValuesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return publicPraiseCultureValuesShowOptions{}, err
	}

	return publicPraiseCultureValuesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPublicPraiseCultureValueDetails(resp jsonAPISingleResponse) publicPraiseCultureValueDetails {
	row := buildPublicPraiseCultureValueRow(resp.Data)
	return publicPraiseCultureValueDetails{
		ID:             row.ID,
		PublicPraiseID: row.PublicPraiseID,
		CultureValueID: row.CultureValueID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func renderPublicPraiseCultureValueDetails(cmd *cobra.Command, details publicPraiseCultureValueDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PublicPraiseID != "" {
		fmt.Fprintf(out, "Public Praise: %s\n", details.PublicPraiseID)
	}
	if details.CultureValueID != "" {
		fmt.Fprintf(out, "Culture Value: %s\n", details.CultureValueID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
