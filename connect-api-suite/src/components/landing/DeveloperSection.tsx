import { Code } from "lucide-react";

const codeExample = `// Authenticate and fetch LinkedIn profile
const response = await fetch('https://api.sociallink.dev/linkedin/profile', {
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});

const profile = await response.json();
console.log(profile);`;

export const DeveloperSection = () => {
  return (
    <section id="developer-section" className="py-16 sm:py-24 bg-background">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid gap-12 lg:grid-cols-2 items-center">
          <div>
            <div className="mb-4 inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2 text-sm font-medium text-primary">
              <Code className="h-4 w-4" />
              <span>Developer Friendly</span>
            </div>
            
            <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl mb-4">
              Built for Developers
            </h2>
            
            <p className="text-lg text-muted-foreground mb-6">
              Simple REST APIs with comprehensive documentation. Get started in minutes 
              with our SDKs and code examples.
            </p>
            
            <ul className="space-y-3">
              <li className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <span className="text-muted-foreground">Supports REST + Webhooks</span>
              </li>
              <li className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <span className="text-muted-foreground">Detailed API documentation</span>
              </li>
              <li className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <span className="text-muted-foreground">SDKs for popular languages</span>
              </li>
              <li className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <span className="text-muted-foreground">Sandbox environment for testing</span>
              </li>
            </ul>
          </div>
          
          <div className="relative">
            <div className="code-block">
              <pre className="text-sm leading-relaxed">
                <code>{codeExample}</code>
              </pre>
            </div>
            <div className="absolute -top-4 -right-4 h-24 w-24 bg-primary/20 rounded-full blur-2xl"></div>
            <div className="absolute -bottom-4 -left-4 h-32 w-32 bg-primary/10 rounded-full blur-2xl"></div>
          </div>
        </div>
      </div>
    </section>
  );
};
