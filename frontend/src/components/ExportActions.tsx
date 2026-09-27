import { FileDown, FileSpreadsheet, FileText } from 'lucide-react'
import { downloadExcel, downloadPdf, type ExportRow } from '../utils/exportData'

export function ExportActions({ filename, title, columns, rows }: { filename: string; title: string; columns: string[]; rows: ExportRow[] }) {
    return <div className="export-actions" aria-label="Descargar datos">
        <button className="icon-button" title="Descargar Excel" aria-label="Descargar Excel" onClick={() => downloadExcel(filename, rows)}><FileSpreadsheet size={16} /></button>
        <button className="icon-button" title="Descargar PDF" aria-label="Descargar PDF" onClick={() => downloadPdf(filename, title, columns, rows)}><FileText size={16} /></button>
        <FileDown size={14} className="export-hint" aria-hidden="true" />
    </div>
}
