"use client"

import React from "react"
import { ArrowUp, Download, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface BillingRecord {
  date: string
  description: string
  amount: string
  status: "paid" | "pending" | "failed"
}

const mockBillingHistory: BillingRecord[] = [
  {
    date: "Oct 26, 2023",
    description: "Pro Plan - Monthly",
    amount: "$49.00",
    status: "paid",
  },
  {
    date: "Sep 26, 2023",
    description: "Pro Plan - Monthly",
    amount: "$49.00",
    status: "paid",
  },
  {
    date: "Aug 26, 2023",
    description: "Pro Plan - Monthly",
    amount: "$49.00",
    status: "paid",
  },
  {
    date: "Jul 26, 2023",
    description: "Pro Plan - Monthly",
    amount: "$49.00",
    status: "paid",
  },
]

const plans = [
  {
    name: "Free",
    price: "$0/month",
    features: [
      "1 Endpoint",
      "1,000 API Calls/month",
      "Community Support",
    ],
    action: "Downgrade",
    actionVariant: "outline" as const,
    isCurrent: false,
  },
  {
    name: "Pro",
    price: "$49/month",
    features: [
      "10 Endpoints",
      "100,000 API Calls/month",
      "Email Support",
      "Advanced Analytics",
    ],
    action: "Your Current Plan",
    actionVariant: "secondary" as const,
    isCurrent: true,
  },
  {
    name: "Enterprise",
    price: "$199/month",
    features: [
      "Unlimited Endpoints",
      "1,000,000 API Calls/month",
      "Priority Phone & Email Support",
      "Dedicated Account Manager",
    ],
    action: "Upgrade to Enterprise",
    actionVariant: "default" as const,
    isCurrent: false,
  },
]

export default function BillingPage() {
  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      {/* Page Title */}
      <h1 className="text-3xl font-bold text-gray-900">Subscription Management</h1>

      {/* Top Section: Current Subscription and Usage Statistics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Current Subscription Card */}
        <Card>
          <CardHeader>
            <CardTitle>Current Subscription</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <h3 className="text-2xl font-semibold">Pro Plan</h3>
                <p className="text-sm text-gray-600">
                  Your plan renews on November 26, 2024.
                </p>
              </div>
              <div className="text-2xl font-bold text-blue-600">$49/month</div>
            </div>
            <div className="flex gap-3">
              <Button>
                <ArrowUp className="mr-2 h-4 w-4" />
                Change Plan
              </Button>
              <Button variant="outline">Cancel Subscription</Button>
            </div>
          </CardContent>
        </Card>

        {/* Usage Statistics Card */}
        <Card>
          <CardHeader>
            <CardTitle>Usage Statistics</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">API Calls</span>
                <span className="font-medium">75,000 / 100,000</span>
              </div>
              <Progress value={75} className="h-2" />
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Endpoints</span>
                <span className="font-medium">8 / 10</span>
              </div>
              <Progress value={80} className="h-2" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Available Plans Section */}
      <div className="space-y-4">
        <h2 className="text-2xl font-bold text-gray-900">Available Plans</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {plans.map((plan) => (
            <Card
              key={plan.name}
              className={plan.isCurrent ? "border-blue-500 border-2" : ""}
            >
              {plan.isCurrent && (
                <div className="px-6 pt-6">
                  <Badge className="bg-blue-600 text-white">Current Plan</Badge>
                </div>
              )}
              <CardHeader className={plan.isCurrent ? "pt-4" : ""}>
                <CardTitle className="text-xl">{plan.name}</CardTitle>
                <p className="text-2xl font-bold text-gray-900">{plan.price}</p>
              </CardHeader>
              <CardContent className="space-y-4">
                <ul className="space-y-2">
                  {plan.features.map((feature, index) => (
                    <li key={index} className="flex items-start gap-2 text-sm">
                      <Check className="h-4 w-4 text-green-600 mt-0.5 shrink-0" />
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
                <Button
                  variant={plan.actionVariant}
                  className="w-full"
                  disabled={plan.isCurrent}
                >
                  {plan.action}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Billing History Section */}
      <div className="space-y-4">
        <h2 className="text-2xl font-bold text-gray-900">Billing History</h2>
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Invoice</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockBillingHistory.map((record, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-medium">{record.date}</TableCell>
                    <TableCell>{record.description}</TableCell>
                    <TableCell>{record.amount}</TableCell>
                    <TableCell>
                      <Badge variant="success">{record.status}</Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="link" className="text-blue-600" asChild>
                        <a href="#" className="flex items-center gap-1">
                          <Download className="h-4 w-4" />
                          Download
                        </a>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
