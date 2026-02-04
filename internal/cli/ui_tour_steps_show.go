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

type uiTourStepsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type uiTourStepDetails struct {
	ID                 string `json:"id"`
	Name               string `json:"name,omitempty"`
	Abbreviation       string `json:"abbreviation,omitempty"`
	Sequence           string `json:"sequence,omitempty"`
	Content            string `json:"content,omitempty"`
	UiTourID           string `json:"ui_tour_id,omitempty"`
	UiTourName         string `json:"ui_tour_name,omitempty"`
	UiTourAbbreviation string `json:"ui_tour_abbreviation,omitempty"`
	UiTourDescription  string `json:"ui_tour_description,omitempty"`
}

func newUiTourStepsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show UI tour step details",
		Long: `Show the full details of a UI tour step.

Output Fields:
  ID                   UI tour step identifier
  Name                 Step name
  Abbreviation         Step abbreviation
  Sequence             Step sequence order
  UI Tour ID           Parent UI tour ID
  UI Tour Name         Parent UI tour name
  UI Tour Abbreviation Parent UI tour abbreviation
  UI Tour Description  Parent UI tour description
  Content              Step content

Arguments:
  <id>  The UI tour step ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a UI tour step
  xbe view ui-tour-steps show 123

  # Output as JSON
  xbe view ui-tour-steps show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUiTourStepsShow,
	}
	initUiTourStepsShowFlags(cmd)
	return cmd
}

func init() {
	uiTourStepsCmd.AddCommand(newUiTourStepsShowCmd())
}

func initUiTourStepsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUiTourStepsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseUiTourStepsShowOptions(cmd)
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
		return fmt.Errorf("ui tour step id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ui-tour-steps]", "name,content,abbreviation,sequence,ui-tour")
	query.Set("include", "ui-tour")
	query.Set("fields[ui-tours]", "name,abbreviation,description")

	body, _, err := client.Get(cmd.Context(), "/v1/ui-tour-steps/"+id, query)
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

	details := buildUiTourStepDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUiTourStepDetails(cmd, details)
}

func parseUiTourStepsShowOptions(cmd *cobra.Command) (uiTourStepsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return uiTourStepsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUiTourStepDetails(resp jsonAPISingleResponse) uiTourStepDetails {
	attrs := resp.Data.Attributes
	details := uiTourStepDetails{
		ID:           resp.Data.ID,
		Name:         strings.TrimSpace(stringAttr(attrs, "name")),
		Abbreviation: strings.TrimSpace(stringAttr(attrs, "abbreviation")),
		Sequence:     strings.TrimSpace(stringAttr(attrs, "sequence")),
		Content:      strings.TrimSpace(stringAttr(attrs, "content")),
	}

	if rel, ok := resp.Data.Relationships["ui-tour"]; ok && rel.Data != nil {
		details.UiTourID = rel.Data.ID
	}

	for _, inc := range resp.Included {
		if inc.Type != "ui-tours" {
			continue
		}
		if inc.ID != details.UiTourID {
			continue
		}
		details.UiTourName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		details.UiTourAbbreviation = strings.TrimSpace(stringAttr(inc.Attributes, "abbreviation"))
		details.UiTourDescription = strings.TrimSpace(stringAttr(inc.Attributes, "description"))
		break
	}

	return details
}

func renderUiTourStepDetails(cmd *cobra.Command, details uiTourStepDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Abbreviation != "" {
		fmt.Fprintf(out, "Abbreviation: %s\n", details.Abbreviation)
	}
	if details.Sequence != "" {
		fmt.Fprintf(out, "Sequence: %s\n", details.Sequence)
	}
	if details.UiTourID != "" {
		fmt.Fprintf(out, "UI Tour ID: %s\n", details.UiTourID)
	}
	if details.UiTourName != "" {
		fmt.Fprintf(out, "UI Tour Name: %s\n", details.UiTourName)
	}
	if details.UiTourAbbreviation != "" {
		fmt.Fprintf(out, "UI Tour Abbrev: %s\n", details.UiTourAbbreviation)
	}
	if details.UiTourDescription != "" {
		fmt.Fprintf(out, "UI Tour Description: %s\n", details.UiTourDescription)
	}
	if details.Content != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Content:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Content)
	}

	return nil
}
