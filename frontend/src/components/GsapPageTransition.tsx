import { useRef, type ReactNode } from 'react'
import { useGsapReveal } from '../hooks/useGsapReveal'

export function GsapPageTransition({ children }: { children: ReactNode }) {
    const rootRef = useRef<HTMLDivElement>(null)
    useGsapReveal(rootRef, '[data-page-reveal]')

    return <div ref={rootRef} className="page-transition" data-page-reveal>{children}</div>
}
