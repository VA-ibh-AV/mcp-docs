import { type ClassValue } from "clsx"
import { cn } from "./utils"

/**
 * Reusable link class utilities
 * Use these throughout your project for consistent link styling
 */

// Base link styles
export const linkBase = "transition-colors"

// Logo/Brand link styles
export const linkLogo = cn(
  linkBase,
  "font-semibold text-xl text-foreground hover:text-primary"
)

// Navigation menu link styles
export const linkNav = cn(
  linkBase,
  "text-foreground hover:text-primary hover:underline"
)

// Navigation menu link (active state)
export const linkNavActive = cn(
  linkNav,
  "text-primary underline"
)

// Footer link styles (for footer navigation)
export const linkFooter = cn(
  linkBase,
  "text-muted-foreground hover:text-foreground"
)

// Text link (inline content links)
export const linkText = cn(
  linkBase,
  "text-primary hover:text-primary-dark underline underline-offset-4"
)

// Helper function to combine link classes with custom classes
export function getLinkClasses(
  variant: "logo" | "nav" | "nav-active" | "footer" | "text",
  className?: ClassValue
) {
  const variantClasses = {
    logo: linkLogo,
    nav: linkNav,
    "nav-active": linkNavActive,
    footer: linkFooter,
    text: linkText,
  }
  
  return cn(variantClasses[variant], className)
}