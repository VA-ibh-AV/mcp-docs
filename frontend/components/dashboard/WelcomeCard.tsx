"use client"

import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

interface WelcomeCardProps {
  userName?: string
}

export function WelcomeCard({ userName = "Alex" }: WelcomeCardProps) {
  return (
    <Card className="bg-gradient-to-r from-blue-600 to-blue-400 text-white border-0 shadow-lg h-full">
      <CardContent className="flex items-center justify-between p-6">
        <div className="flex flex-col gap-2">
          <h2 className="text-2xl font-bold">Welcome back, {userName}!</h2>
          <p className="text-blue-50 text-sm">
            Ready to get started? Index your first set of documents to begin.
          </p>
        </div>
        <Button
          variant="secondary"
          className="bg-white text-blue-600 hover:bg-blue-50 font-medium"
        >
          <Plus className="mr-2 h-4 w-4" />
          Index New Document...
        </Button>
      </CardContent>
    </Card>
  )
}
