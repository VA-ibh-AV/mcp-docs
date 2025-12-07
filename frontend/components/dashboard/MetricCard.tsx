"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { TrendingUp } from "lucide-react"

interface MetricCardProps {
  title: string
  value: string | number
  change?: string
  changeType?: "positive" | "negative"
}

export function MetricCard({
  title,
  value,
  change,
  changeType = "positive",
}: MetricCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium text-gray-600">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="text-3xl font-bold text-gray-900">
            {typeof value === "number" ? value.toLocaleString() : value}
          </div>
          {change && (
            <div
              className={`flex items-center gap-1 text-sm ${
                changeType === "positive" ? "text-green-600" : "text-red-600"
              }`}
            >
              <TrendingUp className="h-4 w-4" />
              <span>{change}</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
