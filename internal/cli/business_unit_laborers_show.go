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

type businessUnitLaborersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type businessUnitLaborerDetails struct {
	ID               string `json:"id"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	LaborerID        string `json:"laborer_id,omitempty"`
	LaborerName      string `json:"laborer_name,omitempty"`
}

func newBusinessUnitLaborersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show business unit laborer details",
		Long: `Show the full details of a business unit laborer link.

Output Fields:
  ID             Link identifier
  Business Unit  Linked business unit
  Laborer        Linked laborer

Arguments:
  <id>    Business unit laborer ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a business unit laborer link
  xbe view business-unit-laborers show 123

  # JSON output
  xbe view business-unit-laborers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBusinessUnitLaborersShow,
	}
	initBusinessUnitLaborersShowFlags(cmd)
	return cmd
}

func init() {
	businessUnitLaborersCmd.AddCommand(newBusinessUnitLaborersShowCmd())
}

func initBusinessUnitLaborersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitLaborersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBusinessUnitLaborersShowOptions(cmd)
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
		return fmt.Errorf("business unit laborer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[business-unit-laborers]", "business-unit,laborer")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[laborers]", "nickname")
	query.Set("include", "business-unit,laborer")

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-laborers/"+id, query)
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

	details := buildBusinessUnitLaborerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBusinessUnitLaborerDetails(cmd, details)
}

func parseBusinessUnitLaborersShowOptions(cmd *cobra.Command) (businessUnitLaborersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitLaborersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBusinessUnitLaborerDetails(resp jsonAPISingleResponse) businessUnitLaborerDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := businessUnitLaborerDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
		if unit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BusinessUnitName = strings.TrimSpace(stringAttr(unit.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["laborer"]; ok && rel.Data != nil {
		details.LaborerID = rel.Data.ID
		if laborer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LaborerName = strings.TrimSpace(stringAttr(laborer.Attributes, "nickname"))
		}
	}

	return details
}

func renderBusinessUnitLaborerDetails(cmd *cobra.Command, details businessUnitLaborerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Business Unit", details.BusinessUnitName, details.BusinessUnitID)
	writeLabelWithID(out, "Laborer", details.LaborerName, details.LaborerID)

	return nil
}
