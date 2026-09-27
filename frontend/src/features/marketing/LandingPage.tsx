import { ArrowRight, BarChart3, BellRing, Building2, Check, ChevronDown, CircleDollarSign, Clock3, CloudSun, Cpu, Gauge, LineChart, Menu, ShieldCheck, Sparkles, Zap } from 'lucide-react'
import { useState } from 'react'
import '../../marketing.css'
import { useRef } from 'react'
import { useGsapReveal } from '../../hooks/useGsapReveal'
import { BrandLogo } from '../../components/BrandLogo'

const faqs = [
    ['¿Puedo cambiar de comercializador de energía en Bogotá?', 'Sí. El cambio de comercializador se gestiona con acompañamiento y sin interrumpir el suministro eléctrico de tu empresa. Primero revisamos la información básica de tus cuentas y puntos de consumo.'],
    ['¿Qué necesito para recibir una propuesta?', 'Necesitamos información de tus sedes y facturas recientes para entender tu perfil de consumo y preparar una propuesta comercial ajustada a tu operación.'],
    ['¿Bia sirve para empresas con varias sedes?', 'Sí. Puedes centralizar la información de tus puntos de consumo, comparar sedes y consultar el desempeño energético desde una misma plataforma.'],
    ['¿Qué información puedo consultar?', 'La plataforma reúne consumo, reportes, alertas y análisis para que tu equipo pueda detectar variaciones y tomar decisiones con datos.'],
]

export function LandingPage() {
    const [openFaq, setOpenFaq] = useState<number | null>(null)
    const [menuOpen, setMenuOpen] = useState(false)
    const pageRef = useRef<HTMLElement>(null)
    useGsapReveal(pageRef, '[data-scroll-reveal]')

    return (
        <main ref={pageRef} className="marketing-page">
            <nav className="marketing-nav" aria-label="Navegación principal">
                <a className="marketing-logo" href="/" aria-label="Bia Energy inicio"><BrandLogo className="brand-logo-marketing" /><strong>Bia</strong><small>energy</small></a>
                <button className="marketing-menu-toggle" onClick={() => setMenuOpen(!menuOpen)} aria-label="Abrir menú"><Menu size={21} /></button>
                <div className={`marketing-links ${menuOpen ? 'open' : ''}`}>
                    <a href="#soluciones">Soluciones</a><a href="#proceso">Cómo funciona</a><a href="#tecnologia">Tecnología</a><a href="#preguntas">Preguntas frecuentes</a>
                    <a className="marketing-login" href="/app">Ingresar <ArrowRight size={15} /></a>
                </div>
            </nav>

            <section className="marketing-hero">
                <div className="hero-copy">
                    <div className="marketing-kicker"><span /> ENERGÍA PARA EMPRESAS EN BOGOTÁ</div>
                    <h1>Comercializamos energía para tu empresa</h1>
                    <p className="hero-lede">Una forma más clara de comprar, entender y gestionar la energía de tu negocio en Bogotá y Colombia.</p>
                    <div className="hero-actions"><a className="marketing-button primary" href="#contacto">Quiero una propuesta <ArrowRight size={17} /></a><a className="marketing-text-link" href="#soluciones">Conoce la solución <ChevronDown size={16} /></a></div>
                    <div className="hero-proof"><span><Check size={14} />Acompañamiento para cambiar de comercializador</span><span><Check size={14} />Información centralizada por sede</span></div>
                </div>
                <div className="energy-visual" aria-label="Panel conceptual de gestión energética">
                    <div className="visual-top"><span><span className="live-dot" />CONSUMO EN VIVO</span><span>BOGOTÁ · 24 H</span></div>
                    <div className="visual-reading"><span>Consumo consolidado</span><strong>Por sede</strong><em><span>↗</span> Comparación por período</em></div>
                    <div className="visual-chart"><div className="chart-axis"><span>Alto</span><span>Medio</span><span>Bajo</span></div><div className="chart-bars">{[38, 48, 42, 61, 54, 68, 57, 74, 66, 82, 72, 91].map((height, index) => <i key={index} style={{ height: `${height}%` }} />)}</div><div className="chart-line"><i /><i /><i /><i /><i /><i /></div></div>
                    <div className="visual-bottom"><span><Gauge size={14} /> Puntos de consumo</span><span><BellRing size={14} /> Alertas y eventos</span></div>
                    <div className="visual-orbit orbit-a" /><div className="visual-orbit orbit-b" />
                </div>
            </section>

            <section className="trust-strip" data-scroll-reveal><span>Una plataforma para tomar mejores decisiones energéticas</span><div><strong>Consumo</strong><strong>Facturación</strong><strong>Analítica</strong><strong>Gestión de sedes</strong></div></section>

            <section className="marketing-section solutions" id="soluciones" data-scroll-reveal>
                <div className="section-intro"><div className="marketing-kicker"><span />LA SOLUCIÓN ENERGÉTICA</div><h2>La solución energética que transforma tu negocio</h2><p>Conecta la comercialización de energía con herramientas que ayudan a tu equipo a entender qué está pasando en cada punto de consumo.</p></div>
                <div className="solution-grid"><SolutionCard icon={<CircleDollarSign />} title="Tarifas competitivas y eficiencia" text="Recibe una propuesta alineada con el perfil de consumo y las necesidades operativas de tu empresa." /><SolutionCard icon={<Cpu />} title="Tecnología de clase mundial" text="Convierte tus lecturas y facturas en información útil para decidir con más contexto." /><SolutionCard icon={<CloudSun />} title="Energía sostenible sin complicaciones" text="Integra una gestión energética más consciente sin añadir complejidad a la operación diaria." /><SolutionCard icon={<Building2 />} title="Control total de múltiples sedes" text="Centraliza tus puntos de consumo y visualiza el comportamiento de tu operación desde un mismo lugar." /><SolutionCard icon={<Clock3 />} title="Soporte experto" text="Cuenta con acompañamiento para resolver dudas y avanzar en la gestión energética de tu empresa." /></div>
            </section>

            <section className="process-section" id="proceso" data-scroll-reveal><div className="process-heading"><div className="marketing-kicker"><span />CAMBIAR ES MÁS SIMPLE</div><h2>3 pasos para cambiarte</h2><p>Te acompañamos durante el proceso para que tu empresa pueda empezar a gestionar mejor su energía.</p></div><div className="process-grid"><ProcessStep number="01" title="Contáctanos" text="Comparte la información básica de tu empresa y tus puntos de consumo." /><ProcessStep number="02" title="Formalizamos nuestra relación" text="Revisamos tu operación y presentamos una propuesta personalizada." /><ProcessStep number="03" title="Instalamos y conectamos" text="Coordinamos el cambio y habilitamos las herramientas para monitorear tu energía." /></div></section>

            <section className="technology-section" id="tecnologia" data-scroll-reveal><div className="tech-visual"><div className="tech-grid-lines" /><div className="tech-center"><Sparkles size={23} /><span>Análisis de consumo</span><strong>Decisiones<br />con contexto</strong></div><div className="tech-node node-one"><BarChart3 size={16} /><span>Reportes</span></div><div className="tech-node node-two"><BellRing size={16} /><span>Alertas</span></div><div className="tech-node node-three"><LineChart size={16} /><span>Pronóstico</span></div></div><div className="tech-copy"><div className="marketing-kicker"><span />INFORMACIÓN PARA DECIDIR</div><h2>Más claridad sobre tu consumo</h2><p>Deja de revisar datos aislados. Obtén una vista accionable del consumo, las alertas y las oportunidades de mejora de tu empresa.</p><div className="tech-list"><span><Check size={16} />Información útil</span><span><Check size={16} />Alertas oportunas</span><span><Check size={16} />Reportes ejecutivos</span><span><Check size={16} />Pronóstico por hora</span></div></div></section>

            <section className="features-section"><div className="section-intro"><div className="marketing-kicker"><span />UNA VISTA COMPLETA</div><h2>Todo lo que necesitas para gestionar mejor</h2></div><div className="feature-grid"><Feature icon={<Zap />} title="Consumo en vivo" text="Consulta el comportamiento de tus puntos de consumo y detecta cambios a tiempo." /><Feature icon={<ShieldCheck />} title="Calidad energética" text="Identifica irregularidades que pueden afectar el rendimiento de tu operación." /><Feature icon={<Gauge />} title="Intensidad energética" text="Relaciona el uso de energía con la actividad de tu empresa para encontrar oportunidades." /></div></section>

            <section className="faq-section" id="preguntas"><div className="section-intro"><div className="marketing-kicker"><span />PREGUNTAS FRECUENTES</div><h2>Resuelve tus dudas sobre energía empresarial</h2><p>Información clara para evaluar si Bia es el siguiente paso para tu empresa.</p></div><div className="faq-list">{faqs.map(([question, answer], index) => <div className={`faq-item ${openFaq === index ? 'open' : ''}`} key={question}><button onClick={() => setOpenFaq(openFaq === index ? null : index)} aria-expanded={openFaq === index}><span>{question}</span><ChevronDown size={18} /></button>{openFaq === index && <p>{answer}</p>}</div>)}</div></section>

            <section className="contact-section" id="contacto"><div><div className="marketing-kicker"><span />EMPIEZA A GESTIONAR MEJOR</div><h2>Haz de tu energía una oportunidad</h2><p>Conversemos sobre el consumo de tu empresa en Bogotá.</p></div><a className="marketing-button light" href="mailto:clientes@bia.app">Solicitar contacto <ArrowRight size={17} /></a></section>
            <footer className="marketing-footer"><div className="marketing-logo"><BrandLogo className="brand-logo-marketing" /><strong>Bia</strong><small>energy</small></div><p>Al fin, tecnología en energía.</p><div className="footer-links"><a href="#soluciones">Soluciones</a><a href="#preguntas">Preguntas frecuentes</a><a href="mailto:clientes@bia.app">clientes@bia.app</a><a href="/app">Acceso a plataforma</a></div><small>© 2026 Bia Energy. Bogotá, Colombia.</small></footer>
        </main>
    )
}

function SolutionCard({ icon, title, text }: { icon: React.ReactNode; title: string; text: string }) { return <article className="solution-card"><span className="card-icon">{icon}</span><h3>{title}</h3><p>{text}</p></article> }
function ProcessStep({ number, title, text }: { number: string; title: string; text: string }) { return <article className="process-step"><span>{number}</span><h3>{title}</h3><p>{text}</p></article> }
function Feature({ icon, title, text }: { icon: React.ReactNode; title: string; text: string }) { return <article className="feature-card"><span className="feature-icon">{icon}</span><h3>{title}</h3><p>{text}</p></article> }