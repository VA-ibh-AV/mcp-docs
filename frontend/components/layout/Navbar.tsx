'use client'
import { NavigationMenu, NavigationMenuItem, NavigationMenuList } from "@/components/ui/navigation-menu"
import { Button } from "@/components/ui/button"
import Link from "next/link"

export default function Navbar() {
  return (
    <nav className="border-b border-gray-200 bg-white shadow-sm sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-8 h-16 flex items-center justify-between">
        
        {/* Logo */}
        <Link href="/" className="font-semibold text-xl">
          MCP-Docs
        </Link>

        {/* Menu */}
        <NavigationMenu>
          <NavigationMenuList className="gap-4">
            <NavigationMenuItem>
              <Link href="/features">Features</Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/pricing">Pricing</Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/documentation">Documentation</Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/about">About</Link>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>

        {/* Right side (Sign In / Try Free) */}
        <div className="flex items-center gap-4">
          <Button variant="link" asChild>
            <Link href="/login">Sign In</Link>
          </Button>
          <Button asChild>
            <Link href="/register">Try Free</Link>
          </Button>
        </div>
      </div>
    </nav>
  )
}
