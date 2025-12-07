import DashboardSideNav from '@/components/dashboard/DashboardSideNav'
import { SidebarProvider, SidebarTrigger, SidebarInset } from '@/components/ui/sidebar'
import React from 'react'

const DashboardLayout = ({ children }: { children: React.ReactNode }) => {
  return (
    <SidebarProvider>
      <DashboardSideNav />
      <SidebarInset>
        {children}
      </SidebarInset>
    </SidebarProvider>
  )
}

export default DashboardLayout