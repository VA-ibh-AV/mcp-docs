"use client"

import { Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"

interface SubscriptionCardProps {
  planName?: string
  renewalDate?: string
}

export function SubscriptionCard({
  planName = "Pro Plan",
  renewalDate = "Oct 24, 2024",
}: SubscriptionCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg font-semibold">Subscription Status</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4 pt-0">
        <div className="flex items-center gap-2 -mt-4">
          <Badge variant="default" className="bg-blue-600 text-white">
            <Star className="mr-1 h-3 w-3" />
            {planName}
          </Badge>
        </div>
        <p className="text-sm text-gray-600">
          Next renewal: {renewalDate}
        </p>
        <Button variant="outline" className="w-full" asChild>
          <Link href="/dashboard/subscription">Manage Subscription</Link>
        </Button>
      </CardContent>
    </Card>
  )
}
