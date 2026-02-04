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

type developerTruckerCertificationMultipliersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type developerTruckerCertificationMultiplierDetails struct {
	ID                              string  `json:"id"`
	Multiplier                      float64 `json:"multiplier,omitempty"`
	DeveloperTruckerCertificationID string  `json:"developer_trucker_certification_id,omitempty"`
	TrailerID                       string  `json:"trailer_id,omitempty"`
}

func newDeveloperTruckerCertificationMultipliersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show developer trucker certification multiplier details",
		Long: `Show the full details of a developer trucker certification multiplier.

Output Fields:
  ID
  Multiplier
  Developer Trucker Certification ID
  Trailer ID

Arguments:
  <id>    Developer trucker certification multiplier ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show multiplier details
  xbe view developer-trucker-certification-multipliers show 123

  # JSON output
  xbe view developer-trucker-certification-multipliers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeveloperTruckerCertificationMultipliersShow,
	}
	initDeveloperTruckerCertificationMultipliersShowFlags(cmd)
	return cmd
}

func init() {
	developerTruckerCertificationMultipliersCmd.AddCommand(newDeveloperTruckerCertificationMultipliersShowCmd())
}

func initDeveloperTruckerCertificationMultipliersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperTruckerCertificationMultipliersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDeveloperTruckerCertificationMultipliersShowOptions(cmd)
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
		return fmt.Errorf("developer trucker certification multiplier id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[developer-trucker-certification-multipliers]", "multiplier,developer-trucker-certification,trailer")

	body, _, err := client.Get(cmd.Context(), "/v1/developer-trucker-certification-multipliers/"+id, query)
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

	details := buildDeveloperTruckerCertificationMultiplierDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeveloperTruckerCertificationMultiplierDetails(cmd, details)
}

func parseDeveloperTruckerCertificationMultipliersShowOptions(cmd *cobra.Command) (developerTruckerCertificationMultipliersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerTruckerCertificationMultipliersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeveloperTruckerCertificationMultiplierDetails(resp jsonAPISingleResponse) developerTruckerCertificationMultiplierDetails {
	details := developerTruckerCertificationMultiplierDetails{
		ID: resp.Data.ID,
	}

	if multiplier, ok := floatAttrValue(resp.Data.Attributes, "multiplier"); ok {
		details.Multiplier = multiplier
	}

	if rel, ok := resp.Data.Relationships["developer-trucker-certification"]; ok && rel.Data != nil {
		details.DeveloperTruckerCertificationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}

	return details
}

func renderDeveloperTruckerCertificationMultiplierDetails(cmd *cobra.Command, details developerTruckerCertificationMultiplierDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Multiplier: %s\n", formatMultiplier(details.Multiplier))
	fmt.Fprintf(out, "Developer Trucker Certification ID: %s\n", details.DeveloperTruckerCertificationID)
	fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)

	return nil
}
