'use client';
import { Check, Copy } from "lucide-react";
import { Button } from "../ui/button";
import { cn } from "@/lib/utils";
import { useState } from "react";

export function Hero() {

    const [copied, setCopied] = useState(false)

    const handleCopy = async () => {
      try {
        await navigator.clipboard.writeText('npm install mcp-docs-cli -g')
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      } catch (err) {
        console.error('Failed to copy:', err)
      }
    }

  return (
    <section className="bg-gradient-to-b from-white to-gray-50 py-32 px-4 font-sans">
      <div className="max-w-7xl mx-auto">
        <div className="max-w-3xl mx-auto text-center">
          {/* Main Headline */}
          <h1 className="text-5xl md:text-6xl font-extrabold text-gray-900 mb-6 font-sans tracking-tight">
            Index Your Documentation with AI
          </h1>

          {/* Subheadline */}
          <p className="text-xl text-gray-600 mb-8 font-sans">
            The ultimate platform to simplify documentation indexing and MCP
            endpoint management using the power of AI.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col md:flex-row gap-4 justify-center">
            <Button variant="default" size="lg" className="text-lg">Try Free</Button>
            <Button variant="outline" size="lg">Learn More</Button>
          </div>


          {/* Installation Code Box */}
          <div className="max-w-lg mx-auto mt-12 bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4 text-left">
              <h3 className="text-gray-900 font-sans text-base font-medium"><strong>OR</strong> Get started with our local free library</h3>
              <button
                onClick={handleCopy}
                className={cn(
                  "flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 hover:text-gray-900",
                  "border border-gray-200 rounded-md hover:bg-gray-50 transition-colors",
                  "focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-1"
                )}
                aria-label="Copy installation command"
              >
                {copied ? (
                  <>
                    <Check className="w-4 h-4" />
                    <span>Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>

            <div className="bg-gray-50 border border-gray-200 rounded-md p-4 text-left">
              <code className="text-gray-900 font-mono text-sm">pip install mcp-docs</code>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
