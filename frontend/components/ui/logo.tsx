import Link from "next/link"
import { cn } from "@/lib/utils"
import { linkLogo } from "@/lib/link-classes"

interface LogoProps {
  href?: string
  className?: string
  showText?: boolean
}

export function Logo({ 
  href = "/", 
  className,
  showText = true 
}: LogoProps) {
  const logoContent = (
    <>
      <svg 
        className="h-6 w-6 text-primary shrink-0" 
        fill="none" 
        viewBox="0 0 48 48" 
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <path 
          d="M44 4H30.6666V17.3334H17.3334V30.6666H4V44H44V4Z" 
          fill="currentColor"
        />
      </svg>
      {showText && (
        <span className={cn(linkLogo, "whitespace-nowrap")}>
          MCP-Docs
        </span>
      )}
    </>
  )

  return (
    <Link 
      href={href} 
      className={cn(
        "flex items-center gap-2", 
        className
      )}
      aria-label="MCP-Docs Home"
    >
      {logoContent}
    </Link>
  )
}