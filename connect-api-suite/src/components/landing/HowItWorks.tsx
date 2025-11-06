import { UserPlus, Key, Rocket } from "lucide-react";

const steps = [
  {
    icon: UserPlus,
    title: "Sign Up",
    description: "Create your free developer account in seconds",
  },
  {
    icon: Key,
    title: "Get API Key",
    description: "Receive your unique API credentials instantly",
  },
  {
    icon: Rocket,
    title: "Start Building",
    description: "Integrate with comprehensive docs and examples",
  },
];

export const HowItWorks = () => {
  return (
    <section className="py-16 sm:py-24 bg-secondary/30">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl mb-4">
            Get Started in Minutes
          </h2>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Simple, straightforward integration process
          </p>
        </div>
        
        <div className="grid gap-8 md:grid-cols-3 max-w-5xl mx-auto">
          {steps.map((step, index) => {
            const Icon = step.icon;
            return (
              <div key={step.title} className="text-center">
                <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground">
                  <Icon className="h-8 w-8" />
                </div>
                <div className="mb-2 text-sm font-semibold text-muted-foreground">
                  Step {index + 1}
                </div>
                <h3 className="mb-2 text-xl font-semibold text-foreground">
                  {step.title}
                </h3>
                <p className="text-muted-foreground">
                  {step.description}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};
