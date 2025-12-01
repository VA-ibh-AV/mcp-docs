import { Hero } from "@/components/landing/Hero";
import Navbar from "@/components/layout/Navbar";
import Features from "@/components/landing/Features";
import Pricing from "@/components/landing/Pricing";
import Footer from "@/components/layout/Footer";

export default function Home() {
  return (
    <> 
      <Navbar />
      <Hero />
      <Features />
      <Pricing />
      <Footer />
    </>
  );
}
