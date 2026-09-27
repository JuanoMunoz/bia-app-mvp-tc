import { useLayoutEffect, type RefObject } from 'react'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

export function useGsapReveal(rootRef: RefObject<HTMLElement | null>, selector = '[data-reveal]') {
    useLayoutEffect(() => {
        const root = rootRef.current
        if (!root) return

        const context = gsap.context(() => {
            const targets = [
                ...(root.matches(selector) ? [root] : []),
                ...gsap.utils.toArray<HTMLElement>(selector),
            ]
            const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

            if (reduceMotion) {
                gsap.set(targets, { clearProps: 'all' })
                return
            }

            gsap.fromTo(targets, { autoAlpha: 0, y: 12 }, {
                autoAlpha: 1,
                y: 0,
                duration: 0.5,
                ease: 'power2.out',
                stagger: 0.06,
                clearProps: 'transform,opacity,visibility',
                scrollTrigger: {
                    trigger: root,
                    start: 'top 88%',
                    once: true,
                },
            })
        }, root)

        return () => context.revert()
    }, [rootRef, selector])
}