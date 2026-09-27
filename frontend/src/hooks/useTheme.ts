import { useEffect, useState } from 'react'

export type Theme = 'dark' | 'light'

const preferenceKey = 'bia-theme-v1'

export function useTheme() {
    const [theme, setTheme] = useState<Theme>(() => {
        const preference = localStorage.getItem(preferenceKey)
        return preference === 'light' ? 'light' : 'dark'
    })

    useEffect(() => {
        document.documentElement.dataset.theme = theme
        localStorage.setItem(preferenceKey, theme)
    }, [theme])

    function toggleTheme() {
        setTheme((current) => current === 'dark' ? 'light' : 'dark')
    }

    return { theme, toggleTheme }
}