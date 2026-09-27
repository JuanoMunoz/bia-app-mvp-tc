package report

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"learning/internal/billing"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (generator *Generator) GenerateBillingReportPDF(data billing.BillingReportData) ([]byte, error) {
	return buildPDF(reportLines(data)), nil
}

func (generator *Generator) GenerateBillingReportExcel(data billing.BillingReportData) ([]byte, error) {
	return buildXLSX(data)
}

func reportLines(data billing.BillingReportData) []string {
	lines := []string{
		"Reporte de facturacion",
		"Periodo: " + data.PeriodLabel,
		fmt.Sprintf("Costo total: %.2f %s", data.TotalCost, data.Currency),
		"",
		"Medidor | kWh | Tarifa | Costo",
	}
	for _, row := range data.Rows {
		lines = append(lines, fmt.Sprintf("%s | %.2f | %s | %.2f %s",
			row.MeterID, row.KWh, row.TariffName, row.Cost, row.Currency))
	}
	return lines
}

func buildPDF(lines []string) []byte {
	var content strings.Builder
	content.WriteString("BT /F1 12 Tf 50 750 Td 15 TL\n")
	for _, line := range lines {
		content.WriteString("(" + escapePDFText(line) + ") Tj T*\n")
	}
	content.WriteString("ET")
	stream := content.String()

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	for index, body := range objects {
		offsets = append(offsets, buffer.Len())
		fmt.Fprintf(&buffer, "%d 0 obj\n%s\nendobj\n", index+1, body)
	}
	xrefStart := buffer.Len()
	fmt.Fprintf(&buffer, "xref\n0 %d\n", len(objects)+1)
	buffer.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(&buffer, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&buffer, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefStart)
	return buffer.Bytes()
}

func escapePDFText(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(value)
}

func buildXLSX(data billing.BillingReportData) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?>` +
			`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
			`<Default Extension="xml" ContentType="application/xml"/>` +
			`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
			`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>` +
			`</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8"?>` +
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
			`</Relationships>`,
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8"?>` +
			`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
			`<sheets><sheet name="Billing" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8"?>` +
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>` +
			`</Relationships>`,
		"xl/worksheets/sheet1.xml": sheetXML(data),
	}
	for name, body := range files {
		entry, err := writer.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write([]byte(body)); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func sheetXML(data billing.BillingReportData) string {
	var rows strings.Builder
	writeRow := func(cells ...string) {
		rows.WriteString("<row>")
		for _, cell := range cells {
			rows.WriteString(`<c t="inlineStr"><is><t>` + escapeXML(cell) + `</t></is></c>`)
		}
		rows.WriteString("</row>")
	}
	writeRow("Reporte de facturacion", data.PeriodLabel)
	writeRow(fmt.Sprintf("Costo total: %.2f %s", data.TotalCost, data.Currency))
	writeRow("Medidor", "kWh", "Tarifa", "Costo", "Moneda")
	for _, item := range data.Rows {
		writeRow(item.MeterID, fmt.Sprintf("%.2f", item.KWh), item.TariffName,
			fmt.Sprintf("%.2f", item.Cost), item.Currency)
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		`<sheetData>` + rows.String() + `</sheetData></worksheet>`
}

func escapeXML(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return replacer.Replace(value)
}
