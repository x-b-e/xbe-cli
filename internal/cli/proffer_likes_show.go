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

type profferLikesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type profferLikeDetails struct {
	ID        string `json:"id"`
	ProfferID string `json:"proffer_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newProfferLikesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show proffer like details",
		Long: `Show the full details of a proffer like.

Output Fields:
  ID
  Proffer
  User
  Created At
  Updated At

Arguments:
  <id>    The proffer like ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a proffer like
  xbe view proffer-likes show 123

  # Get JSON output
  xbe view proffer-likes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProfferLikesShow,
	}
	initProfferLikesShowFlags(cmd)
	return cmd
}

func init() {
	profferLikesCmd.AddCommand(newProfferLikesShowCmd())
}

func initProfferLikesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfferLikesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProfferLikesShowOptions(cmd)
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
		return fmt.Errorf("proffer like id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[proffer-likes]", "created-at,updated-at,proffer,user")

	body, _, err := client.Get(cmd.Context(), "/v1/proffer-likes/"+id, query)
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

	details := buildProfferLikeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProfferLikeDetails(cmd, details)
}

func parseProfferLikesShowOptions(cmd *cobra.Command) (profferLikesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return profferLikesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return profferLikesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return profferLikesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return profferLikesShowOptions{}, err
	}

	return profferLikesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProfferLikeDetails(resp jsonAPISingleResponse) profferLikeDetails {
	resource := resp.Data
	row := buildProfferLikeRow(resource)
	return profferLikeDetails{
		ID:        row.ID,
		ProfferID: row.ProfferID,
		UserID:    row.UserID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func renderProfferLikeDetails(cmd *cobra.Command, details profferLikeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProfferID != "" {
		fmt.Fprintf(out, "Proffer: %s\n", details.ProfferID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
