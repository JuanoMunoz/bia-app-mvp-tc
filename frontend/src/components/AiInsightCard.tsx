import { useEffect, useState } from 'react'
import { MessageCircleQuestion, Sparkles } from 'lucide-react'
import type { AnalysisInsight } from '../types'

export function AiInsightCard({ insight, title = 'Asistente de energía' }: { insight: AnalysisInsight; title?: string }) {
    const [activeIndex, setActiveIndex] = useState<number | null>(null)
    const selectedText = activeIndex === null ? insight.answer : insight.suggested_questions[activeIndex]?.answer ?? insight.answer

    return (
        <section className="ai-chat-card panel" aria-label="Asistente de IA para análisis de energía">
            <div className="insight-header ai-chat-header">
                <span className="insight-icon ai-chat-icon"><Sparkles size={17} /></span>
                <div>
                    <div className="eyebrow">ASISTENTE IA</div>
                    <h2>{title}</h2>
                </div>
            </div>

            <div className="ai-chat-bubble" aria-live="polite">
                <div className="ai-bubble-label">{activeIndex === null ? 'Resumen ejecutivo' : 'Respuesta'}</div>
                <TypedText key={selectedText} text={selectedText} />
            </div>

            <div className="insight-questions ai-chat-questions">
                <div className="eyebrow">3 preguntas sugeridas</div>
                <div className="question-list ai-question-list">
                    {insight.suggested_questions.map((item, index) => (
                        <button className={activeIndex === index ? 'active' : ''} key={item.question} onClick={() => setActiveIndex(index)}>
                            <MessageCircleQuestion size={15} />
                            {item.question}
                        </button>
                    ))}
                </div>
            </div>

            <details className="insight-explanation">
                <summary>Ver explicación técnica</summary>
                <p>{insight.explanation}</p>
            </details>
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
            if (position >= text.length) {
                window.clearInterval(interval)
            }
        }, 18)

        return () => window.clearInterval(interval)
    }, [text])

    return <p className="ai-chat-answer">{visibleText}<span className="typing-caret" aria-hidden="true" /></p>
}

export function InsightChip({ text }: { text: string }) {
    return <span className="ai-chip"><Sparkles size={12} />{text}</span>
}
