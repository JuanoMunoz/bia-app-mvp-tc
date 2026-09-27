import logoJpg from '../assets/logo.jpg'
import logoWebp from '../assets/logo.webp'

export function BrandLogo({ className = '', alt = 'Bia Energy' }: { className?: string; alt?: string }) {
    return <picture className={`brand-logo ${className}`}><source srcSet={logoWebp} type="image/webp" /><img src={logoJpg} alt={alt} /></picture>
}