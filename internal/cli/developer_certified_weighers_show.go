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

type developerCertifiedWeighersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type developerCertifiedWeigherDetails struct {
	ID            string `json:"id"`
	Number        string `json:"number,omitempty"`
	IsActive      bool   `json:"is_active"`
	DeveloperID   string `json:"developer_id,omitempty"`
	DeveloperName string `json:"developer_name,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserName      string `json:"user_name,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
}

func newDeveloperCertifiedWeighersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show developer certified weigher details",
		Long: `Show the full details of a developer certified weigher.

Output Fields:
  ID
  NUMBER
  ACTIVE
  DEVELOPER (name + ID)
  USER (name + email + ID)

Arguments:
  <id>    The developer certified weigher ID (required). Use the list command to find IDs.`,
		Example: `  # Show a developer certified weigher
  xbe view developer-certified-weighers show 123

  # Show as JSON
  xbe view developer-certified-weighers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeveloperCertifiedWeighersShow,
	}
	initDeveloperCertifiedWeighersShowFlags(cmd)
	return cmd
}

func init() {
	developerCertifiedWeighersCmd.AddCommand(newDeveloperCertifiedWeighersShowCmd())
}

func initDeveloperCertifiedWeighersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperCertifiedWeighersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDeveloperCertifiedWeighersShowOptions(cmd)
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
		return fmt.Errorf("developer certified weigher id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[developer-certified-weighers]", "number,is-active,developer,user")
	query.Set("include", "developer,user")
	query.Set("fields[developers]", "name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/developer-certified-weighers/"+id, query)
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

	details := buildDeveloperCertifiedWeigherDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeveloperCertifiedWeigherDetails(cmd, details)
}

func parseDeveloperCertifiedWeighersShowOptions(cmd *cobra.Command) (developerCertifiedWeighersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerCertifiedWeighersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeveloperCertifiedWeigherDetails(resp jsonAPISingleResponse) developerCertifiedWeigherDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	details := developerCertifiedWeigherDetails{
		ID:       resource.ID,
		Number:   stringAttr(resource.Attributes, "number"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}

	if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
		details.DeveloperID = rel.Data.ID
		if developer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DeveloperName = stringAttr(developer.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return details
}

func renderDeveloperCertifiedWeigherDetails(cmd *cobra.Command, details developerCertifiedWeigherDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Number != "" {
		fmt.Fprintf(out, "Number: %s\n", details.Number)
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)

	if details.DeveloperName != "" {
		fmt.Fprintf(out, "Developer Name: %s\n", details.DeveloperName)
	}
	if details.DeveloperID != "" {
		fmt.Fprintf(out, "Developer ID: %s\n", details.DeveloperID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}

	return nil
}
