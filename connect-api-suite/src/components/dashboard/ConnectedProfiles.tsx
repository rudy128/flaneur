import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Linkedin, Twitter, Instagram, MessageCircle } from "lucide-react";

const profiles = [
  { platform: "LinkedIn", count: 3, icon: Linkedin, color: "text-[#0A66C2]" },
  { platform: "Twitter", count: 5, icon: Twitter, color: "text-[#1DA1F2]" },
  { platform: "Instagram", count: 2, icon: Instagram, color: "text-[#E4405F]" },
  { platform: "WhatsApp", count: 1, icon: MessageCircle, color: "text-[#25D366]" },
];

export function ConnectedProfiles() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {profiles.map((profile) => {
        const Icon = profile.icon;
        return (
          <Card key={profile.platform}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {profile.platform}
              </CardTitle>
              <Icon className={`h-4 w-4 ${profile.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{profile.count}</div>
              <p className="text-xs text-muted-foreground">
                Connected {profile.count === 1 ? 'profile' : 'profiles'}
              </p>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
