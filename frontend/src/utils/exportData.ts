export type ExportRow = Record<string, string | number>

export async function downloadExcel(filename: string, rows: ExportRow[]) {
    const XLSX = await import('xlsx')
    const worksheet = XLSX.utils.json_to_sheet(rows)
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, 'Datos')
    XLSX.writeFile(workbook, `${filename}.xlsx`)
}

export async function downloadPdf(filename: string, title: string, columns: string[], rows: ExportRow[]) {
    const [{ default: jsPDF }, { default: autoTable }] = await Promise.all([import('jspdf'), import('jspdf-autotable')])
    const document = new jsPDF({ orientation: 'landscape' })
    document.setFontSize(16)
    document.text(title, 14, 15)
    autoTable(document, {
        startY: 22,
        head: [columns],
        body: rows.map((row) => columns.map((column) => String(row[column] ?? ''))),
        styles: { fontSize: 8 },
        headStyles: { fillColor: [0, 16, 53] },
    })
    document.save(`${filename}.pdf`)
}
