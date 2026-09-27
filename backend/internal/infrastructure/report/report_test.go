package report

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"learning/internal/billing"
)

func sampleData() billing.BillingReportData {
	return billing.BillingReportData{
		From:        time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		To:          time.Date(2026, 9, 26, 23, 59, 0, 0, time.UTC),
		Currency:    "COP",
		TotalCost:   12795222,
		PeriodLabel: "2026-08-28 a 2026-09-26",
		Rows: []billing.MeterBillingReportRow{
			{MeterID: "M-104", KWh: 18543.8, TariffName: "bia / industrial", Cost: 12795222, Currency: "COP", Calculated: true},
		},
	}
}

func TestGenerateBillingReportPDFReturnsValidPDF(t *testing.T) {
	content, err := NewGenerator().GenerateBillingReportPDF(sampleData())
	if err != nil {
		t.Fatalf("generate pdf returned error: %v", err)
	}
	if !bytes.HasPrefix(content, []byte("%PDF-")) {
		t.Fatalf("pdf content does not start with %%PDF- header")
	}
	if !bytes.Contains(content, []byte("M-104")) {
		t.Fatalf("pdf content does not contain meter id")
	}
}

func TestGenerateBillingReportExcelReturnsValidXLSX(t *testing.T) {
	content, err := NewGenerator().GenerateBillingReportExcel(sampleData())
	if err != nil {
		t.Fatalf("generate excel returned error: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("excel content is not a valid zip archive: %v", err)
	}
	names := map[string]bool{}
	for _, file := range reader.File {
		names[file.Name] = true
	}
	for _, required := range []string{"[Content_Types].xml", "xl/workbook.xml", "xl/worksheets/sheet1.xml"} {
		if !names[required] {
			t.Fatalf("xlsx archive is missing %s", required)
		}
	}
}
