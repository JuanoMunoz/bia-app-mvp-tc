import { useEffect, useState } from 'react'
import { MessageCircleQuestion, Sparkles } from 'lucide-react'
import type { AnalysisInsight } from '../types'

export function InsightCard({ insight }: { insight: AnalysisInsight }) {
    const [activeQuestion, setActiveQuestion] = useState<number | null>(null)
    const sourceText = activeQuestion === null ? insight.answer : insight.suggested_questions[activeQuestion]?.answer ?? insight.answer
    return (
        <section className="insight-card panel" aria-label="Explicación del análisis">
            <div className="insight-header"><span className="insight-icon"><Sparkles size={17} /></span><div><div className="eyebrow">LECTURA DEL ANÁLISIS</div><h2>{activeQuestion === null ? 'Qué significa para tu empresa' : 'Respuesta a tu pregunta'}</h2></div></div>
            <TypedText key={sourceText} text={sourceText} />
            <details className="insight-explanation"><summary>Ver explicación</summary><p>{insight.explanation}</p></details>
            <div className="insight-questions"><div className="eyebrow">También puedes preguntar</div><div className="question-list">{insight.suggested_questions.map((item, index) => <button className={activeQuestion === index ? 'active' : ''} key={item.question} onClick={() => setActiveQuestion(index)}><MessageCircleQuestion size={15} />{item.question}</button>)}</div></div>
        </section>
    )
}

function TypedText({ text }: { text: string }) {
    const [visibleText, setVisibleText] = useState('')

    useEffect(() => {
        let position = 0
        const interval = window.setInterval(() => {
            position += 2
            setVisibleText(text.slice(0, position))
            if (position >= text.length) window.clearInterval(interval)
        }, 18)
        return () => window.clearInterval(interval)
    }, [text])

    return <p className="insight-answer">{visibleText}<span className="typing-caret" aria-hidden="true" /></p>
}