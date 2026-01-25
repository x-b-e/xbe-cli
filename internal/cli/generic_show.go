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

type genericShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newGenericShowCmd(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: fmt.Sprintf("Show %s", resource),
		Long: fmt.Sprintf(`Show %s details.

Arguments:
  <id>  The %s ID (required).`, resource, resource),
		Example: fmt.Sprintf("  xbe view %s show 12345", resource),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenericShow(cmd, resource, args[0])
		},
	}
	initGenericShowFlags(cmd)
	return cmd
}

func initGenericShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func parseGenericShowOptions(cmd *cobra.Command) (genericShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return genericShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return genericShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return genericShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return genericShowOptions{}, err
	}

	return genericShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func runGenericShow(cmd *cobra.Command, resource string, id string) error {
	opts, err := parseGenericShowOptions(cmd)
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

	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%s id is required", resource)
	}

	client := api.NewClient(opts.BaseURL, opts.Token)
	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/"+resource+"/"+id, query)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), resp.Data)
	}
	return writeJSON(cmd.OutOrStdout(), resp.Data)
}
