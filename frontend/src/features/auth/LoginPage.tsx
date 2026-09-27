import { useState } from 'react'
import { ChevronRight, Zap } from 'lucide-react'
import { BrandLogo } from '../../components/BrandLogo'

const mockUser = import.meta.env.VITE_MOCK_LOGIN_USER ?? 'demo@bia.energy'
const mockPassword = import.meta.env.VITE_MOCK_LOGIN_PASSWORD ?? 'BiaDemo2026!'

export function LoginPage({ onLogin }: { onLogin: () => void }) {
    const [credentialsError, setCredentialsError] = useState(false)
    return (
        <main className="login-page">
            <div className="login-art" aria-hidden="true">
                <span className="art-grid" />
                <div className="art-orbit orbit-one" />
                <div className="art-orbit orbit-two" />
                <div className="art-readout"><Zap size={16} /> PUNTOS DE CONSUMO</div>
                <div className="art-mark"><BrandLogo className="brand-logo-art" /><span className="art-mark-dot" /></div>
                <div className="art-caption"><span>GESTIÓN DE ENERGÍA</span><strong>Datos claros para decidir.</strong></div>
            </div>
            <div className="login-side">
                <div className="login-brand"><BrandLogo className="brand-logo-login" /><strong>Bia</strong></div>
                    <form className="login-form" onSubmit={(event) => {
                        event.preventDefault()
                        const form = new FormData(event.currentTarget)
                        const isValid = form.get('email') === mockUser && form.get('password') === mockPassword
                        setCredentialsError(!isValid)
                        if (isValid) onLogin()
                    }}>
                    <div className="eyebrow">INGRESO A TU CUENTA</div>
                    <h1>Bienvenido de nuevo</h1>
                    <p>Consulta y gestiona el consumo de tu empresa.</p>
                    <label htmlFor="email">Correo corporativo</label>
                    <input id="email" name="email" type="email" autoComplete="username" placeholder="nombre@empresa.com" required />
                    <label htmlFor="password">Contraseña</label>
                    <input id="password" name="password" type="password" autoComplete="current-password" placeholder="Ingresa tu contraseña" required />
                    <p className="login-error" style={{ color: "#ff1111" }} role="alert" hidden={!credentialsError}>Usuario o contraseña incorrectos.</p>
                    <button className="button button-primary login-submit" type="submit">Ingresar <ChevronRight size={17} /></button>
                    <div className="login-foot">Acceso seguro a tu cuenta</div>
                </form>
                <span className="login-copyright">BIA ENERGY · 2026</span>
            </div>
        </main>
    )
}