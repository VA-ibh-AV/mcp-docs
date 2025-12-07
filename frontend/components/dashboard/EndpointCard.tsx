"use client"

import { Edit, Trash2 } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"

interface EndpointCardProps {
  name: string
  status: "active" | "stopped" | "error" | "starting"
  url: string
  createdAt: string
  onEdit?: () => void
  onDelete?: () => void
}

export function EndpointCard({
  name,
  status,
  url,
  createdAt,
  onEdit,
  onDelete,
}: EndpointCardProps) {
  const statusConfig = {
    active: { label: "Active", variant: "success" as const, dotColor: "bg-green-500" },
    stopped: { label: "Stopped", variant: "secondary" as const, dotColor: "bg-gray-500" },
    error: { label: "Error", variant: "destructive" as const, dotColor: "bg-red-500" },
    starting: { label: "Starting", variant: "warning" as const, dotColor: "bg-amber-500" },
  }

  const config = statusConfig[status]

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">{name}</CardTitle>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={onEdit}
            >
              <Edit className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-red-600 hover:text-red-700"
              onClick={onDelete}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          <div className={`h-2 w-2 rounded-full ${config.dotColor}`} />
          <Badge variant={config.variant}>{config.label}</Badge>
        </div>
        <div className="space-y-2">
          <div>
            <p className="text-xs text-gray-500 mb-1">URL</p>
            <p className="text-sm font-mono text-gray-900">{url}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">Created: {createdAt}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
