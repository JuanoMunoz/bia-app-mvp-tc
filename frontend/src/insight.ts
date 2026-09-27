import type { AnalysisInsight } from './types'

export function buildFriendlyInsight(answer: string, explanation: string, questions: Array<{ question: string; answer: string }>): AnalysisInsight {
    return {
        answer,
        explanation,
        suggested_questions: questions,
    }
}
