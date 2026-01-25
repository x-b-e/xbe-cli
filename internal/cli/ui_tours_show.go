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

type uiToursShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type uiTourDetails struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Abbreviation  string   `json:"abbreviation"`
	Description   string   `json:"description,omitempty"`
	UiTourStepIDs []string `json:"ui_tour_step_ids,omitempty"`
	UserUiTourIDs []string `json:"user_ui_tour_ids,omitempty"`
}

func newUiToursShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show UI tour details",
		Long: `Show the full details of a UI tour.

Output Fields:
  ID
  Name
  Abbreviation
  Description
  UI Tour Step IDs
  User UI Tour IDs

Arguments:
  <id>  The UI tour ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a UI tour
  xbe view ui-tours show 123

  # Get JSON output
  xbe view ui-tours show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUiToursShow,
	}
	initUiToursShowFlags(cmd)
	return cmd
}

func init() {
	uiToursCmd.AddCommand(newUiToursShowCmd())
}

func initUiToursShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUiToursShow(cmd *cobra.Command, args []string) error {
	opts, err := parseUiToursShowOptions(cmd)
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
		return fmt.Errorf("ui tour id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ui-tours]", "name,abbreviation,description,ui-tour-steps,user-ui-tours")

	body, _, err := client.Get(cmd.Context(), "/v1/ui-tours/"+id, query)
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

	details := buildUiTourDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUiTourDetails(cmd, details)
}

func parseUiToursShowOptions(cmd *cobra.Command) (uiToursShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return uiToursShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUiTourDetails(resp jsonAPISingleResponse) uiTourDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return uiTourDetails{
		ID:            resource.ID,
		Name:          stringAttr(attrs, "name"),
		Abbreviation:  stringAttr(attrs, "abbreviation"),
		Description:   strings.TrimSpace(stringAttr(attrs, "description")),
		UiTourStepIDs: relationshipIDsFromMap(resource.Relationships, "ui-tour-steps"),
		UserUiTourIDs: relationshipIDsFromMap(resource.Relationships, "user-ui-tours"),
	}
}

func renderUiTourDetails(cmd *cobra.Command, details uiTourDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Abbreviation != "" {
		fmt.Fprintf(out, "Abbreviation: %s\n", details.Abbreviation)
	}
	if details.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Description)
	}
	if len(details.UiTourStepIDs) > 0 {
		fmt.Fprintf(out, "UI Tour Step IDs: %s\n", strings.Join(details.UiTourStepIDs, ", "))
	}
	if len(details.UserUiTourIDs) > 0 {
		fmt.Fprintf(out, "User UI Tour IDs: %s\n", strings.Join(details.UserUiTourIDs, ", "))
	}

	return nil
}
