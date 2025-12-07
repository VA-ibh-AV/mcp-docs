"use client"

import React, { useState } from "react"
import { Plus, Search, MoreVertical, ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface Endpoint {
  id: string
  name: string
  status: "active" | "inactive" | "error"
  subscriptionTier: string
  lastUpdated: string
  creationDate: string
}

const mockEndpoints: Endpoint[] = [
  {
    id: "1",
    name: "API Gateway Production",
    status: "active",
    subscriptionTier: "Enterprise",
    lastUpdated: "2023-10-26 10:05 AM",
    creationDate: "2023-01-15 09:00 AM",
  },
  {
    id: "2",
    name: "Staging Environment",
    status: "active",
    subscriptionTier: "Pro",
    lastUpdated: "2023-10-25 03:22 PM",
    creationDate: "2023-02-20 11:30 AM",
  },
  {
    id: "3",
    name: "Development Server - Auth",
    status: "inactive",
    subscriptionTier: "Free",
    lastUpdated: "2023-10-24 08:15 AM",
    creationDate: "2023-03-10 02:45 PM",
  },
  {
    id: "4",
    name: "Legacy Data Indexer",
    status: "error",
    subscriptionTier: "Pro",
    lastUpdated: "2023-10-23 05:30 PM",
    creationDate: "2022-12-05 10:20 AM",
  },
  {
    id: "5",
    name: "Q4 Reporting Endpoint",
    status: "active",
    subscriptionTier: "Enterprise",
    lastUpdated: "2023-10-26 09:00 AM",
    creationDate: "2023-09-01 08:00 AM",
  },
]

export default function EndpointsPage() {
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [subscriptionFilter, setSubscriptionFilter] = useState("all")
  const [sortBy, setSortBy] = useState("lastUpdated")
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 5

  const filteredEndpoints = mockEndpoints.filter((endpoint) => {
    const matchesSearch =
      endpoint.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      endpoint.id.includes(searchQuery)
    const matchesStatus = statusFilter === "all" || endpoint.status === statusFilter
    const matchesSubscription =
      subscriptionFilter === "all" || endpoint.subscriptionTier === subscriptionFilter
    return matchesSearch && matchesStatus && matchesSubscription
  })

  const totalPages = Math.ceil(filteredEndpoints.length / itemsPerPage)
  const startIndex = (currentPage - 1) * itemsPerPage
  const paginatedEndpoints = filteredEndpoints.slice(
    startIndex,
    startIndex + itemsPerPage
  )

  const getStatusBadge = (status: string) => {
    const config = {
      active: { label: "Active", variant: "success" as const, dotColor: "bg-green-500" },
      inactive: { label: "Inactive", variant: "secondary" as const, dotColor: "bg-gray-500" },
      error: { label: "Error", variant: "destructive" as const, dotColor: "bg-red-500" },
    }
    const statusConfig = config[status as keyof typeof config] || config.inactive
    return (
      <div className="flex items-center gap-2">
        <div className={`h-2 w-2 rounded-full ${statusConfig.dotColor}`} />
        <Badge variant={statusConfig.variant}>{statusConfig.label}</Badge>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">MCP Endpoint Management</h1>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create New Endpoint
        </Button>
      </div>

      {/* Search and Filters */}
      <div className="space-y-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <Input
            placeholder="Search endpoints by name or ID..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex gap-4">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Status: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
              <SelectItem value="error">Error</SelectItem>
            </SelectContent>
          </Select>
          <Select value={subscriptionFilter} onValueChange={setSubscriptionFilter}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Subscription: All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="Free">Free</SelectItem>
              <SelectItem value="Pro">Pro</SelectItem>
              <SelectItem value="Enterprise">Enterprise</SelectItem>
            </SelectContent>
          </Select>
          <Select value={sortBy} onValueChange={setSortBy}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Sort By: Last Updated" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="lastUpdated">Last Updated</SelectItem>
              <SelectItem value="name">Name</SelectItem>
              <SelectItem value="creationDate">Creation Date</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <input type="checkbox" className="rounded border-gray-300" />
              </TableHead>
              <TableHead>Endpoint Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Subscription Tier</TableHead>
              <TableHead>Last Updated</TableHead>
              <TableHead>Creation Date</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedEndpoints.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                  No endpoints found
                </TableCell>
              </TableRow>
            ) : (
              paginatedEndpoints.map((endpoint) => (
                <TableRow key={endpoint.id}>
                  <TableCell>
                    <input type="checkbox" className="rounded border-gray-300" />
                  </TableCell>
                  <TableCell className="font-medium">{endpoint.name}</TableCell>
                  <TableCell>{getStatusBadge(endpoint.status)}</TableCell>
                  <TableCell>{endpoint.subscriptionTier}</TableCell>
                  <TableCell>{endpoint.lastUpdated}</TableCell>
                  <TableCell>{endpoint.creationDate}</TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem>Edit</DropdownMenuItem>
                        <DropdownMenuItem>View Details</DropdownMenuItem>
                        <DropdownMenuItem className="text-red-600">Delete</DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-gray-600">
          Showing {startIndex + 1} to {Math.min(startIndex + itemsPerPage, filteredEndpoints.length)} of {filteredEndpoints.length} results
        </p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setCurrentPage((prev) => Math.max(1, prev - 1))}
            disabled={currentPage === 1}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
            <Button
              key={page}
              variant={currentPage === page ? "default" : "outline"}
              size="icon"
              onClick={() => setCurrentPage(page)}
              className="w-10"
            >
              {page}
            </Button>
          ))}
          <Button
            variant="outline"
            size="icon"
            onClick={() => setCurrentPage((prev) => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
