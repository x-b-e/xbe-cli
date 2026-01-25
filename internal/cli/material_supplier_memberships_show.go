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

type materialSupplierMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMaterialSupplierMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material supplier membership details",
		Long: `Show the full details of a material supplier membership.

Retrieves and displays comprehensive information about a membership including
user information, material supplier, role settings, and configuration options.

Arguments:
  <id>    The material supplier membership ID (required).`,
		Example: `  # View a material supplier membership by ID
  xbe view material-supplier-memberships show 686

  # Get membership as JSON
  xbe view material-supplier-memberships show 686 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSupplierMembershipsShow,
	}
	initMaterialSupplierMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	materialSupplierMembershipsCmd.AddCommand(newMaterialSupplierMembershipsShowCmd())
}

func initMaterialSupplierMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSupplierMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSupplierMembershipsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material supplier membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker,project-office")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[project-offices]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/material-supplier-memberships/"+id, query)
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

	details := buildMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMembershipDetails(cmd, details)
}

func parseMaterialSupplierMembershipsShowOptions(cmd *cobra.Command) (materialSupplierMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSupplierMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}
