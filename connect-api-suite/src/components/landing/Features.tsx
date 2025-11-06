import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Linkedin, Twitter, Instagram, MessageCircle } from "lucide-react";

const features = [
  {
    icon: Linkedin,
    title: "LinkedIn API",
    description: "Post updates, fetch analytics, manage connections, and automate your LinkedIn presence.",
  },
  {
    icon: Twitter,
    title: "Twitter/X API",
    description: "Tweet, retweet, manage timelines, and analyze engagement with powerful Twitter integration.",
  },
  {
    icon: Instagram,
    title: "Instagram API",
    description: "Post photos, manage stories, fetch insights, and automate your Instagram workflow.",
  },
  {
    icon: MessageCircle,
    title: "WhatsApp API",
    description: "Send messages, manage chats, automate responses, and integrate WhatsApp Business.",
  },
];

export const Features = () => {
  return (
    <section className="py-16 sm:py-24 bg-background">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl mb-4">
            Powerful APIs for Every Platform
          </h2>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Unified access to the world's leading social media platforms
          </p>
        </div>
        
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((feature) => {
            const Icon = feature.icon;
            return (
              <Card 
                key={feature.title} 
                className="hover-lift border-border/50 bg-card"
              >
                <CardHeader>
                  <div className="mb-2 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                    <Icon className="h-6 w-6 text-primary" />
                  </div>
                  <CardTitle className="text-xl">{feature.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-sm leading-relaxed">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </section>
  );
};
