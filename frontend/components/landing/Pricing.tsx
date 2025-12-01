'use client'

import { Button } from "@/components/ui/button"
import { Check } from "lucide-react"
import { cn } from "@/lib/utils"
import Link from "next/link"

interface PricingTier {
  name: string
  price: string
  description: string
  features: string[]
  cta: string
  popular?: boolean
  highlight?: boolean
}

const pricingTiers: PricingTier[] = [
  {
    name: "Free",
    price: "$0",
    description: "Perfect for getting started",
    features: [
      "100 SSE executions/month",
      "1 project",
      "1 concurrent indexing job",
      "50 pages per index",
      "Basic semantic search",
      "Standard MCP server",
      "Community support"
    ],
    cta: "Get Started",
    highlight: false
  },
  {
    name: "Basic",
    price: "$9.99",
    description: "For small teams and projects",
    features: [
      "1,000 SSE executions/month",
      "5 projects",
      "3 concurrent indexing jobs",
      "200 pages per index",
      "Basic semantic search",
      "Standard MCP server",
      "Email support",
      "Usage analytics"
    ],
    cta: "Start Free Trial",
    highlight: false
  },
  {
    name: "Pro",
    price: "$29.99",
    description: "For growing businesses",
    features: [
      "10,000 SSE executions/month",
      "20 projects",
      "10 concurrent indexing jobs",
      "1,000 pages per index",
      "Basic semantic search",
      "AI-powered search (CrewAI)",
      "Priority support",
      "Advanced analytics",
      "API access"
    ],
    cta: "Start Free Trial",
    popular: true,
    highlight: true
  },
  {
    name: "Advanced",
    price: "$99.99",
    description: "For enterprise needs",
    features: [
      "100,000 SSE executions/month",
      "Unlimited projects",
      "Unlimited concurrent jobs",
      "Unlimited pages per index",
      "AI-powered search (CrewAI)",
      "Enhanced MCP tools",
      "24/7 priority support",
      "Advanced analytics",
      "API access with higher limits",
      "Custom integrations"
    ],
    cta: "Contact Sales",
    highlight: false
  }
]

export default function Pricing() {
  return (
    <section className="py-24 px-4 bg-gray-50 font-sans">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
            Simple, Transparent Pricing
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Choose the plan that's right for you.
          </p>
        </div>

        {/* Pricing Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {pricingTiers.map((tier, index) => (
            <div
              key={index}
              className={cn(
                "bg-white border rounded-lg p-8 flex flex-col",
                tier.highlight
                  ? "border-primary shadow-lg ring-2 ring-primary ring-offset-2"
                  : "border-gray-200 shadow-md hover:shadow-lg transition-shadow"
              )}
            >
              {/* Popular Badge */}
              {tier.popular && (
                <div className="mb-4">
                  <span className="inline-block px-3 py-1 text-xs font-semibold text-primary bg-primary/10 rounded-full">
                    Most Popular
                  </span>
                </div>
              )}

              {/* Tier Name */}
              <h3 className="text-2xl font-bold text-gray-900 mb-2">{tier.name}</h3>
              <p className="text-gray-600 mb-6 text-sm">{tier.description}</p>

              {/* Price */}
              <div className="mb-6">
                <span className="text-4xl font-bold text-gray-900">{tier.price}</span>
                <span className="text-gray-600 ml-2">/month</span>
              </div>

              {/* Features List */}
              <ul className="flex-1 mb-8 space-y-3">
                {tier.features.map((feature, featureIndex) => (
                  <li key={featureIndex} className="flex items-start">
                    <Check className="w-5 h-5 text-primary mr-2 flex-shrink-0 mt-0.5" />
                    <span className="text-gray-600 text-sm">{feature}</span>
                  </li>
                ))}
              </ul>

              {/* CTA Button */}
              <Button
                asChild
                variant={tier.highlight ? "default" : "outline"}
                className="w-full"
                size="lg"
              >
                <Link href="/register">{tier.cta}</Link>
              </Button>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

