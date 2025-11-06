import { Button } from "@/components/ui/button";
import { Code2 } from "lucide-react";
import { useNavigate } from "react-router-dom";

export const Header = () => {
  const navigate = useNavigate();

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center gap-2 cursor-pointer" onClick={() => navigate('/')}>
            <Code2 className="h-6 w-6 text-primary" />
            <span className="text-xl font-bold text-foreground">SocialLink API</span>
          </div>
          
          <nav className="hidden md:flex items-center gap-6 text-sm font-medium">
            <a href="#" className="text-muted-foreground hover:text-foreground transition-colors">
              Docs
            </a>
            <a href="#" className="text-muted-foreground hover:text-foreground transition-colors">
              Pricing
            </a>
            <a href="#" className="text-muted-foreground hover:text-foreground transition-colors">
              API Reference
            </a>
          </nav>
          
          <div className="flex items-center gap-3">
            <Button variant="ghost" onClick={() => navigate('/login')}>
              Login
            </Button>
            <Button onClick={() => navigate('/signup')}>
              Sign Up
            </Button>
          </div>
        </div>
      </div>
    </header>
  );
};
