package cli

import "strings"

type materialTransactionTicketGeneratorRow struct {
	ID               string `json:"id"`
	FormatRule       string `json:"format_rule,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
}

func buildMaterialTransactionTicketGeneratorRow(resource jsonAPIResource) materialTransactionTicketGeneratorRow {
	row := materialTransactionTicketGeneratorRow{
		ID:         resource.ID,
		FormatRule: stringAttr(resource.Attributes, "format-rule"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}

func buildMaterialTransactionTicketGeneratorRows(resp jsonAPIResponse) []materialTransactionTicketGeneratorRow {
	rows := make([]materialTransactionTicketGeneratorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionTicketGeneratorRow(resource))
	}
	return rows
}

func buildMaterialTransactionTicketGeneratorRowFromSingle(resp jsonAPISingleResponse) materialTransactionTicketGeneratorRow {
	return buildMaterialTransactionTicketGeneratorRow(resp.Data)
}

func normalizeOrganizationFilter(orgType, orgID string) string {
	if orgType == "" || orgID == "" {
		return ""
	}
	return strings.TrimSpace(orgType) + "|" + strings.TrimSpace(orgID)
}
