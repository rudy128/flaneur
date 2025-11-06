import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Linkedin, Twitter, Instagram, MessageCircle, Loader2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { twitterApi } from "@/lib/api";

const platformsConfig = [
  { platform: "LinkedIn", icon: Linkedin, color: "text-[#0A66C2]", bgColor: "bg-[#0A66C2]/10" },
  { platform: "Twitter", icon: Twitter, color: "text-[#1DA1F2]", bgColor: "bg-[#1DA1F2]/10" },
  { platform: "Instagram", icon: Instagram, color: "text-[#E4405F]", bgColor: "bg-[#E4405F]/10" },
  { platform: "WhatsApp", icon: MessageCircle, color: "text-[#25D366]", bgColor: "bg-[#25D366]/10" },
];

export function ConnectedProfiles() {
  const token = localStorage.getItem("jwt_token");

  // Fetch Twitter accounts
  const { data: twitterData, isLoading } = useQuery({
    queryKey: ["twitter-accounts"],
    queryFn: () => twitterApi.getAccounts(token || ""),
    enabled: !!token,
  });

  const twitterCount = twitterData?.count || 0;

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {platformsConfig.map((config) => {
        const Icon = config.icon;
        const count = config.platform === "Twitter" ? twitterCount : 0;
        
        return (
          <Card key={config.platform}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {config.platform}
              </CardTitle>
              <div className={`h-8 w-8 rounded-lg ${config.bgColor} flex items-center justify-center`}>
                <Icon className={`h-4 w-4 ${config.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading && config.platform === "Twitter" ? (
                <div className="flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Loading...</span>
                </div>
              ) : (
                <>
                  <div className="text-2xl font-bold">{count}</div>
                  <p className="text-xs text-muted-foreground">
                    Connected {count === 1 ? 'profile' : 'profiles'}
                  </p>
                </>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
