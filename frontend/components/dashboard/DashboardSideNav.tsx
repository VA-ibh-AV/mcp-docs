"use client"

import { LayoutDashboard, Folder, Network, Settings, User, LogOut, CreditCard } from "lucide-react"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useEffect } from "react"
 
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Logo } from "../ui/logo"
import { Separator } from "../ui/separator"
import { useAppSelector, useAppDispatch } from "@/lib/hooks"
import { logout } from "@/features/auth/authSlice"
import { Button } from "@/components/ui/button"

// Menu items matching the design
const items = [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: LayoutDashboard,
    },
    {
      title: "Endpoints",
      url: "/dashboard/endpoints",
      icon: Network,
    },
    {
      title: "Billing",
      url: "/dashboard/billing",
      icon: CreditCard,
    },
  ]

export default function DashboardSideNav() {
    const pathname = usePathname()
    const router = useRouter()
    const dispatch = useAppDispatch()

    const user = useAppSelector((state) => state.auth.user)
    const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated)

    // Handle authentication redirect using useEffect to avoid hooks rule violations
    useEffect(() => {
        if (!isAuthenticated || !user) {
            router.push("/login")
        }
    }, [isAuthenticated, user, router])

    // Don't render if not authenticated to avoid errors
    if (!isAuthenticated || !user) {
        return null
    }

    const userName = user?.name || user?.email || "User"
    const userEmail = user?.email || "user@example.com"

    const handleLogout = () => {
        dispatch(logout())
        router.push("/login")
    }
    return (
        <Sidebar>
          <SidebarHeader className="px-6 py-4">
            <div className="flex justify-center">
                <Logo href="/dashboard" showText={true} />
            </div>
            <Separator />
          </SidebarHeader>
          <SidebarContent className="px-4">
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  {items.map((item) => {
                    const isActive = pathname === item.url || pathname?.startsWith(item.url + "/")
                    return (
                      <SidebarMenuItem key={item.title}>
                        <SidebarMenuButton 
                          asChild 
                          isActive={isActive}
                          className={isActive 
                            ? "bg-gray-100 text-gray-900 hover:bg-gray-100 hover:text-gray-900 rounded-l-none rounded-r-md font-medium" 
                            : "text-gray-700 hover:bg-gray-50"
                          }
                        >
                          <Link href={item.url}>
                            <item.icon className={isActive ? "text-gray-900" : "text-gray-600"} size={20} />
                            <span className={isActive ? "text-gray-900 font-medium" : "text-gray-700"}>{item.title}</span>
                          </Link>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    )
                  })}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter className="px-6 py-4 border-t border-gray-200 space-y-3">
            <div className="flex items-center gap-3">
              <Avatar className="h-10 w-10">
                <AvatarFallback className="bg-orange-100 text-orange-600">
                  <User className="h-5 w-5" />
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col flex-1 min-w-0">
                <span className="text-sm font-medium text-gray-900 truncate">{userName}</span>
                <span className="text-xs text-gray-500 truncate">{userEmail}</span>
              </div>
            </div>
            <Button
              variant="ghost"
              className="w-full justify-start text-gray-700 hover:text-gray-900 hover:bg-gray-50"
              onClick={handleLogout}
            >
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </SidebarFooter>
        </Sidebar>
      )
}