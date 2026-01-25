package cli

type truckerInvoiceRow struct {
	ID             string `json:"id"`
	Status         string `json:"status,omitempty"`
	InvoiceDate    string `json:"invoice_date,omitempty"`
	DueOn          string `json:"due_on,omitempty"`
	TotalAmount    string `json:"total_amount,omitempty"`
	TimeCardAmount string `json:"time_card_amount,omitempty"`
	CurrencyCode   string `json:"currency_code,omitempty"`
	QuickbooksID   string `json:"quickbooks_id,omitempty"`
	BuyerID        string `json:"buyer_id,omitempty"`
	BuyerType      string `json:"buyer_type,omitempty"`
	SellerID       string `json:"seller_id,omitempty"`
	SellerType     string `json:"seller_type,omitempty"`
}

func buildTruckerInvoiceRow(resource jsonAPIResource) truckerInvoiceRow {
	attrs := resource.Attributes
	row := truckerInvoiceRow{
		ID:             resource.ID,
		Status:         stringAttr(attrs, "status"),
		InvoiceDate:    formatDate(stringAttr(attrs, "invoice-date")),
		DueOn:          formatDate(stringAttr(attrs, "due-on")),
		TotalAmount:    stringAttr(attrs, "total-amount"),
		TimeCardAmount: stringAttr(attrs, "time-card-amount"),
		CurrencyCode:   stringAttr(attrs, "currency-code"),
		QuickbooksID:   stringAttr(attrs, "quickbooks-id"),
	}

	row.BuyerID, row.BuyerType = relationshipRefFromMap(resource.Relationships, "buyer")
	row.SellerID, row.SellerType = relationshipRefFromMap(resource.Relationships, "seller")

	return row
}

func buildTruckerInvoiceRows(resp jsonAPIResponse) []truckerInvoiceRow {
	rows := make([]truckerInvoiceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTruckerInvoiceRow(resource))
	}
	return rows
}

func buildTruckerInvoiceRowFromSingle(resp jsonAPISingleResponse) truckerInvoiceRow {
	return buildTruckerInvoiceRow(resp.Data)
}
