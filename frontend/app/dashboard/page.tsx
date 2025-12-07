"use client"

import React from "react"
import { WelcomeCard } from "@/components/dashboard/WelcomeCard"
import { SubscriptionCard } from "@/components/dashboard/SubscriptionCard"
import { MetricCard } from "@/components/dashboard/MetricCard"
import { EndpointsSection } from "@/components/dashboard/EndpointsSection"
import { useAppSelector } from "@/lib/hooks"

const DashboardPage = () => {
  const user = useAppSelector((state) => state.auth.user)
  const userName = user?.name || user?.email?.split("@")[0] || "Alex"

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      {/* Top Row: Welcome and Subscription */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <WelcomeCard userName={userName} />
        </div>
        <div className="lg:col-span-1">
          <SubscriptionCard />
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <MetricCard
          title="Documents Indexed"
          value={1420}
          change="+5.2%"
          changeType="positive"
        />
        <MetricCard
          title="API Calls This Month"
          value={8765}
          change="+12.8%"
          changeType="positive"
        />
        <MetricCard
          title="Active Endpoints"
          value={6}
          change="+1 this month"
          changeType="positive"
        />
      </div>

      {/* Endpoints Section */}
      <EndpointsSection />
    </div>
  )
}

export default DashboardPage