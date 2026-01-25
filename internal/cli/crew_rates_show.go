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

type crewRatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type crewRateDetails struct {
	ID                         string `json:"id"`
	Description                string `json:"description,omitempty"`
	PricePerUnit               string `json:"price_per_unit,omitempty"`
	StartOn                    string `json:"start_on,omitempty"`
	EndOn                      string `json:"end_on,omitempty"`
	IsActive                   bool   `json:"is_active,omitempty"`
	BrokerID                   string `json:"broker_id,omitempty"`
	ResourceClassificationType string `json:"resource_classification_type,omitempty"`
	ResourceClassificationID   string `json:"resource_classification_id,omitempty"`
	ResourceType               string `json:"resource_type,omitempty"`
	ResourceID                 string `json:"resource_id,omitempty"`
	CraftClassID               string `json:"craft_class_id,omitempty"`
}

func newCrewRatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show crew rate details",
		Long: `Show the full details of a specific crew rate.

Output Fields:
  ID                     Crew rate identifier
  Description            Description
  Price Per Unit         Price per unit
  Start On               Start date
  End On                 End date
  Active                 Active status
  Broker                 Broker ID
  Resource Classification Resource classification type and ID
  Resource               Resource type and ID
  Craft Class            Craft class ID

Arguments:
  <id>    The crew rate ID (required). You can find IDs using the list command.`,
		Example: `  # View a crew rate by ID
  xbe view crew-rates show 123

  # Get crew rate as JSON
  xbe view crew-rates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCrewRatesShow,
	}
	initCrewRatesShowFlags(cmd)
	return cmd
}

func init() {
	crewRatesCmd.AddCommand(newCrewRatesShowCmd())
}

func initCrewRatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewRatesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCrewRatesShowOptions(cmd)
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
		return fmt.Errorf("crew rate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[crew-rates]", "description,price-per-unit,start-on,end-on,is-active,broker,resource,resource-classification,craft-class")

	body, _, err := client.Get(cmd.Context(), "/v1/crew-rates/"+id, query)
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

	details := buildCrewRateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCrewRateDetails(cmd, details)
}

func parseCrewRatesShowOptions(cmd *cobra.Command) (crewRatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewRatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCrewRateDetails(resp jsonAPISingleResponse) crewRateDetails {
	attrs := resp.Data.Attributes

	details := crewRateDetails{
		ID:           resp.Data.ID,
		Description:  strings.TrimSpace(stringAttr(attrs, "description")),
		PricePerUnit: stringAttr(attrs, "price-per-unit"),
		StartOn:      formatDate(stringAttr(attrs, "start-on")),
		EndOn:        formatDate(stringAttr(attrs, "end-on")),
		IsActive:     boolAttr(attrs, "is-active"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["resource-classification"]; ok && rel.Data != nil {
		details.ResourceClassificationType = rel.Data.Type
		details.ResourceClassificationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["resource"]; ok && rel.Data != nil {
		details.ResourceType = rel.Data.Type
		details.ResourceID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["craft-class"]; ok && rel.Data != nil {
		details.CraftClassID = rel.Data.ID
	}

	return details
}

func renderCrewRateDetails(cmd *cobra.Command, details crewRateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.PricePerUnit != "" {
		fmt.Fprintf(out, "Price Per Unit: %s\n", details.PricePerUnit)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.ResourceClassificationType != "" && details.ResourceClassificationID != "" {
		fmt.Fprintf(out, "Resource Classification: %s/%s\n", details.ResourceClassificationType, details.ResourceClassificationID)
	}
	if details.ResourceType != "" && details.ResourceID != "" {
		fmt.Fprintf(out, "Resource: %s/%s\n", details.ResourceType, details.ResourceID)
	}
	if details.CraftClassID != "" {
		fmt.Fprintf(out, "Craft Class: %s\n", details.CraftClassID)
	}

	return nil
}
