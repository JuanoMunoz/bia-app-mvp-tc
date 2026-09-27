import { ArrowLeft, Compass } from 'lucide-react'

export function NotFoundPage() {
    return (
        <main className="not-found-page mesh-gradient">
            <div className="not-found-mark"><Compass size={20} /></div>
            <span className="eyebrow">ERROR 404</span>
            <h1>Esta página se salió del mapa.</h1>
            <p>La ruta que buscas no existe o ya fue movida. Regresa al inicio para continuar explorando.</p>
            <a className="button button-primary" href="/"><ArrowLeft size={16} />Volver al inicio</a>
        </main>
    )
}