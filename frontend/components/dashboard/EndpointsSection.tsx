"use client"

import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { EndpointCard } from "./EndpointCard"

interface Endpoint {
  id: string
  name: string
  status: "active" | "stopped" | "error" | "starting"
  url: string
  createdAt: string
}

interface EndpointsSectionProps {
  endpoints?: Endpoint[]
}

export function EndpointsSection({ endpoints = [] }: EndpointsSectionProps) {
  const defaultEndpoints: Endpoint[] = [
    {
      id: "1",
      name: "Production API",
      status: "active",
      url: "https://api.mcp.app/prod/v1",
      createdAt: "Aug 12, 2024",
    },
    {
      id: "2",
      name: "Staging Environment",
      status: "active",
      url: "https://api.mcp.app/staging/v1",
      createdAt: "Jul 28, 2024",
    },
  ]

  const displayEndpoints = endpoints.length > 0 ? endpoints : defaultEndpoints

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">MCP Endpoints</h2>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create New Endpoint
        </Button>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {displayEndpoints.map((endpoint) => (
          <EndpointCard
            key={endpoint.id}
            name={endpoint.name}
            status={endpoint.status}
            url={endpoint.url}
            createdAt={endpoint.createdAt}
            onEdit={() => console.log("Edit", endpoint.id)}
            onDelete={() => console.log("Delete", endpoint.id)}
          />
        ))}
      </div>
    </div>
  )
}
