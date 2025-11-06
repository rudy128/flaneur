import { Button } from "@/components/ui/button";
import { ArrowRight, Code2 } from "lucide-react";
import { useNavigate } from "react-router-dom";

export const Hero = () => {
  const navigate = useNavigate();

  return (
    <section className="relative overflow-hidden bg-gradient-to-b from-background to-secondary/30 pt-20 pb-16 sm:pt-32 sm:pb-24">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-4xl text-center">
          <div className="mb-6 inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2 text-sm font-medium text-primary">
            <Code2 className="h-4 w-4" />
            <span>Powerful Social Media APIs for Developers</span>
          </div>
          
          <h1 className="mb-6 text-4xl font-bold tracking-tight text-foreground sm:text-5xl md:text-6xl">
            Connect, Automate, and Scale <br />
            <span className="text-primary">with Unified Social APIs</span>
          </h1>
          
          <p className="mb-10 text-lg text-muted-foreground sm:text-xl max-w-2xl mx-auto">
            Easily integrate LinkedIn, Twitter, Instagram, and WhatsApp into your apps 
            with one powerful API platform.
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Button 
              size="lg" 
              className="gap-2 w-full sm:w-auto"
              onClick={() => navigate('/signup')}
            >
              Get API Access
              <ArrowRight className="h-4 w-4" />
            </Button>
            <Button 
              size="lg" 
              variant="outline" 
              className="w-full sm:w-auto"
              onClick={() => document.getElementById('developer-section')?.scrollIntoView({ behavior: 'smooth' })}
            >
              View Docs
            </Button>
          </div>
        </div>
      </div>
      
      {/* Decorative background elements */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-6xl h-full opacity-30 pointer-events-none">
        <div className="absolute top-20 left-10 w-72 h-72 bg-primary/20 rounded-full blur-3xl"></div>
        <div className="absolute bottom-20 right-10 w-96 h-96 bg-primary/10 rounded-full blur-3xl"></div>
      </div>
    </section>
  );
};
