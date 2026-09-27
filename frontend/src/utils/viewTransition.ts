type ViewTransitionDocument = Document & {
    startViewTransition?: (update: () => void | Promise<void>) => void
}

export function startViewTransition(update: () => void | Promise<void>) {
    const documentWithTransitions = document as ViewTransitionDocument
    if (documentWithTransitions.startViewTransition) {
        documentWithTransitions.startViewTransition(update)
        return
    }
    update()
}