import { SidebarProvider } from "@/components/ui/sidebar";
import { DashboardSidebar } from "@/components/dashboard/DashboardSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Linkedin, Twitter, Instagram, Facebook, MessageCircle, Eye, Pencil, Trash2, Plus, Loader2, Copy, Check, Key, ExternalLink } from "lucide-react";
import { toast } from "@/hooks/use-toast";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { twitterAccountSchema, type TwitterAccountRequest } from "@/lib/schemas";
import { twitterApi, whatsappApi } from "@/lib/api";
import { useEffect, useRef } from "react";

interface ConnectedAccount {
  id: string;
  name: string;
  handle: string;
  followers: number;
}

const platforms = [
  { 
    name: "LinkedIn", 
    icon: Linkedin, 
    color: "text-blue-600",
    bgColor: "bg-blue-50",
    accounts: [
      { id: "1", name: "John Doe", handle: "@johndoe", followers: 5420 },
      { id: "2", name: "Jane Smith", handle: "@janesmith", followers: 8932 },
    ]
  },
  { 
    name: "Twitter", 
    icon: Twitter, 
    color: "text-sky-500",
    bgColor: "bg-sky-50",
    accounts: [
      { id: "3", name: "TechUser", handle: "@techuser", followers: 12500 },
    ]
  },
  { 
    name: "Instagram", 
    icon: Instagram, 
    color: "text-pink-600",
    bgColor: "bg-pink-50",
    accounts: [
      { id: "4", name: "CreativeStudio", handle: "@creativestudio", followers: 25800 },
      { id: "5", name: "PhotoDaily", handle: "@photodaily", followers: 15200 },
    ]
  },
  { 
    name: "Facebook", 
    icon: Facebook, 
    color: "text-blue-700",
    bgColor: "bg-blue-50",
    accounts: [
      { id: "6", name: "Business Page", handle: "@businesspage", followers: 34500 },
    ]
  },
  { 
    name: "WhatsApp", 
    icon: MessageCircle, 
    color: "text-green-600",
    bgColor: "bg-green-50",
    accounts: [
      { id: "7", name: "Support Bot", handle: "@supportbot", followers: 0 },
    ]
  },
];

const ApiManagement = () => {
  const [viewAccount, setViewAccount] = useState<{ platform: string; account: ConnectedAccount } | null>(null);
  const [deleteAccount, setDeleteAccount] = useState<{ platform: string; account: ConnectedAccount } | null>(null);
  const [editAccount, setEditAccount] = useState<{ platform: string; account: ConnectedAccount } | null>(null);
  const [editHandle, setEditHandle] = useState("");
  const [connectModalOpen, setConnectModalOpen] = useState(false);
  const [selectedPlatform, setSelectedPlatform] = useState<string>("twitter");
  const [isConnecting, setIsConnecting] = useState(false);
  const [copiedKey, setCopiedKey] = useState(false);
  
  // WhatsApp QR Code state
  const [whatsappQRCode, setWhatsappQRCode] = useState<string | null>(null);
  const [whatsappSessionId, setWhatsappSessionId] = useState<string | null>(null);
  const [whatsappStatus, setWhatsappStatus] = useState<string>("pending");
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const queryClient = useQueryClient();
  const token = localStorage.getItem("jwt_token");

  // Fetch connected Twitter accounts
  const { data: accountsData, isLoading: isLoadingAccounts } = useQuery({
    queryKey: ["twitter-accounts"],
    queryFn: () => twitterApi.getAccounts(token || ""),
    enabled: !!token,
  });

  // Fetch connected WhatsApp accounts
  const { data: whatsappAccountsData, isLoading: isLoadingWhatsAppAccounts } = useQuery({
    queryKey: ["whatsapp-accounts"],
    queryFn: () => whatsappApi.getAccounts(token || ""),
    enabled: !!token,
  });

  const { register, handleSubmit, formState: { errors }, reset: resetForm } = useForm<TwitterAccountRequest>({
    resolver: zodResolver(twitterAccountSchema),
  });

  const connectAccountMutation = useMutation({
    mutationFn: (data: TwitterAccountRequest) => 
      twitterApi.addAccount(token || "", { username: data.username, password: data.password }),
    onSuccess: (data) => {
      toast({ 
        title: "Account Connected!", 
        description: `Twitter account @${data.username} has been connected successfully.` 
      });
      setConnectModalOpen(false);
      resetForm();
      setIsConnecting(false);
      queryClient.invalidateQueries({ queryKey: ["twitter-accounts"] });
    },
    onError: (error: Error) => {
      toast({ 
        title: "Connection Failed", 
        description: error.message || "Failed to connect account. Please try again.", 
        variant: "destructive" 
      });
      setIsConnecting(false);
    },
  });

  const onSubmit = (data: TwitterAccountRequest) => {
    setIsConnecting(true);
    connectAccountMutation.mutate(data);
  };

  // Delete WhatsApp account mutation
  const deleteWhatsAppAccountMutation = useMutation({
    mutationFn: (accountId: string) => whatsappApi.deleteAccount(token || "", accountId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["whatsapp-accounts"] });
      toast({
        title: "Account Deleted",
        description: "WhatsApp account has been removed successfully.",
      });
      setDeleteAccount(null);
    },
    onError: (error: Error) => {
      toast({
        title: "Delete Failed",
        description: error.message || "Failed to delete account. Please try again.",
        variant: "destructive",
      });
    },
  });

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    setCopiedKey(true);
    toast({ title: "API Key Copied!", description: "The API key has been copied to your clipboard." });
    setTimeout(() => setCopiedKey(false), 2000);
  };

  // WhatsApp QR Code generation
  const generateWhatsAppQR = async () => {
    if (!token) return;
    
  // Removed client-side WhatsApp account limit
    
    try {
      setIsConnecting(true);
      const response = await whatsappApi.generateQR(token);
      setWhatsappQRCode(response.qr_code);
      setWhatsappSessionId(response.session_id);
      setWhatsappStatus(response.status);
      
      // Start polling for authentication status
      startPolling(response.session_id);
      
      toast({
        title: "QR Code Generated",
        description: "Scan the QR code with WhatsApp to connect your account.",
      });
    } catch (error) {
      toast({
        title: "Failed to Generate QR Code",
        description: error instanceof Error ? error.message : "Please try again.",
        variant: "destructive",
      });
      setIsConnecting(false);
    }
  };

  // Poll session status
  const startPolling = (sessionId: string) => {
    console.log("🔄 Starting polling for session:", sessionId);
    
    // Clear any existing interval
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
    }

    pollingIntervalRef.current = setInterval(async () => {
      if (!token) {
        console.log("❌ No token available for polling");
        return;
      }

      try {
        console.log("📡 Polling session status for:", sessionId);
        const status = await whatsappApi.checkSessionStatus(token, sessionId);
        console.log("✅ Session status response:", status);
        setWhatsappStatus(status.status);

        if (status.status === "authenticated") {
          // Success! Stop polling
          console.log("🎉 Authentication successful!");
          if (pollingIntervalRef.current) {
            clearInterval(pollingIntervalRef.current);
          }
          
          toast({
            title: "WhatsApp Connected!",
            description: `Account ${status.phone_number} has been connected successfully.`,
          });
          
          setConnectModalOpen(false);
          setWhatsappQRCode(null);
          setWhatsappSessionId(null);
          setIsConnecting(false);
          queryClient.invalidateQueries({ queryKey: ["whatsapp-accounts"] });
        } else if (status.status === "failed" || status.status === "expired") {
          // Failed/Expired - stop polling
          console.log("❌ Authentication failed/expired");
          if (pollingIntervalRef.current) {
            clearInterval(pollingIntervalRef.current);
          }
          
          toast({
            title: "Authentication Failed",
            description: status.message,
            variant: "destructive",
          });
          
          setWhatsappQRCode(null);
          setWhatsappSessionId(null);
          setIsConnecting(false);
        }
      } catch (error) {
        console.error("❌ Error polling session status:", error);
      }
    }, 2000); // Poll every 2 seconds
  };

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, []);

  // Handle platform selection
  const handlePlatformChange = (platform: string) => {
    setSelectedPlatform(platform);
    // Reset WhatsApp state when switching platforms
    if (platform !== "whatsapp") {
      setWhatsappQRCode(null);
      setWhatsappSessionId(null);
      setWhatsappStatus("pending");
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    }
  };

  const twitterAccounts = accountsData?.accounts || [];
  const whatsappAccounts = whatsappAccountsData?.accounts || [];
  const hasWhatsAppAccount = whatsappAccounts.length > 0;

  const handleEdit = (platform: string, account: ConnectedAccount) => {
    setEditAccount({ platform, account });
    setEditHandle(account.handle);
  };

  const handleSaveEdit = () => {
    if (!editHandle.startsWith("@")) {
      toast({
        title: "Invalid Handle",
        description: "Handle must start with @",
        variant: "destructive",
      });
      return;
    }
    
    toast({
      title: "Success",
      description: "Handle updated successfully",
    });
    setEditAccount(null);
  };

  const handleDelete = () => {
    if (!deleteAccount) return;

    if (deleteAccount.platform === "WhatsApp") {
      // Delete WhatsApp account
      deleteWhatsAppAccountMutation.mutate(deleteAccount.account.id);
    } else {
      // For other platforms, just show toast (not implemented yet)
      toast({
        title: "Account Deleted",
        description: "Connected account has been removed",
      });
      setDeleteAccount(null);
    }
  };

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <DashboardSidebar />
        
        <div className="flex-1 flex flex-col">
          <DashboardHeader />
          
          <main className="flex-1 p-6 space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold tracking-tight mb-2">Manage APIs & Connected Accounts</h1>
                <p className="text-muted-foreground">
                  View, edit, or remove accounts for each social media platform.
                </p>
              </div>
              <Button onClick={() => setConnectModalOpen(true)} size="lg">
                <Plus className="h-5 w-5 mr-2" />
                Connect Account
              </Button>
            </div>

            <Card className="p-6">
              <Accordion type="single" collapsible className="w-full">
                {/* Twitter Platform with Real Data */}
                <AccordionItem value="platform-twitter">
                  <AccordionTrigger className="hover:no-underline">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-sky-50">
                        <Twitter className="h-5 w-5 text-sky-500" />
                      </div>
                      <div className="text-left">
                        <div className="font-semibold">Twitter</div>
                        <div className="text-sm text-muted-foreground">
                          {isLoadingAccounts ? (
                            <span className="flex items-center gap-2">
                              <Loader2 className="h-3 w-3 animate-spin" />
                              Loading...
                            </span>
                          ) : (
                            `${twitterAccounts.length} connected account${twitterAccounts.length !== 1 ? 's' : ''}`
                          )}
                        </div>
                      </div>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    {isLoadingAccounts ? (
                      <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
                      </div>
                    ) : twitterAccounts.length === 0 ? (
                      <div className="text-center py-8">
                        <Twitter className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                        <h3 className="text-sm font-semibold text-gray-900 mb-1">No Twitter Accounts</h3>
                        <p className="text-sm text-muted-foreground mb-4">Connect your first Twitter account to get started</p>
                        <Button size="sm" onClick={() => setConnectModalOpen(true)}>
                          <Plus className="h-4 w-4 mr-2" />
                          Connect Twitter Account
                        </Button>
                      </div>
                    ) : (
                      <div className="space-y-3 pt-4">
                        {twitterAccounts.map((account) => (
                          <div
                            key={account.id}
                            className="flex items-center justify-between p-4 rounded-lg border bg-card hover:bg-accent/50 transition-colors cursor-pointer"
                            onClick={() => setViewAccount({ platform: "Twitter", account: { id: account.id, name: account.username, handle: `@${account.username}`, followers: 0 } })}
                          >
                            <div className="flex items-center gap-3">
                              <div className="h-10 w-10 rounded-full bg-sky-100 flex items-center justify-center">
                                <Twitter className="h-5 w-5 text-sky-500" />
                              </div>
                              <div className="flex-1">
                                <div className="font-medium">@{account.username}</div>
                                <div className="text-sm text-muted-foreground">Twitter Account</div>
                              </div>
                            </div>
                            <Button size="sm" variant="ghost" onClick={(e) => {
                              e.stopPropagation();
                              setViewAccount({ platform: "Twitter", account: { id: account.id, name: account.username, handle: `@${account.username}`, followers: 0 } });
                            }}>
                              <Key className="h-4 w-4 mr-2" />
                              View API Key
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </AccordionContent>
                </AccordionItem>

                {/* Other Platforms (No Data - Show Empty State) */}
                {platforms.filter(p => p.name !== "Twitter" && p.name !== "WhatsApp").map((platform, idx) => {
                  const Icon = platform.icon;
                  return (
                    <AccordionItem key={platform.name} value={`platform-${idx}`}>
                      <AccordionTrigger className="hover:no-underline">
                        <div className="flex items-center gap-3">
                          <div className={`p-2 rounded-lg ${platform.bgColor}`}>
                            <Icon className={`h-5 w-5 ${platform.color}`} />
                          </div>
                          <div className="text-left">
                            <div className="font-semibold">{platform.name}</div>
                            <div className="text-sm text-muted-foreground">
                              0 connected accounts
                            </div>
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div className="text-center py-8">
                          <Icon className={`h-12 w-12 mx-auto mb-3 ${platform.color} opacity-30`} />
                          <h3 className="text-sm font-semibold text-gray-900 mb-1">No {platform.name} Accounts</h3>
                          <p className="text-sm text-muted-foreground">No {platform.name} accounts are connected here</p>
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  );
                })}

                {/* WhatsApp Platform with Real Data */}
                <AccordionItem value="platform-whatsapp">
                  <AccordionTrigger className="hover:no-underline">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-green-50">
                        <MessageCircle className="h-5 w-5 text-green-600" />
                      </div>
                      <div className="text-left">
                        <div className="font-semibold">WhatsApp</div>
                        <div className="text-sm text-muted-foreground">
                          {isLoadingWhatsAppAccounts ? (
                            <span className="flex items-center gap-2">
                              <Loader2 className="h-3 w-3 animate-spin" />
                              Loading...
                            </span>
                          ) : (
                            `${whatsappAccountsData?.accounts?.length || 0} connected account${whatsappAccountsData?.accounts?.length !== 1 ? 's' : ''}`
                          )}
                        </div>
                      </div>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    {isLoadingWhatsAppAccounts ? (
                      <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
                      </div>
                    ) : (whatsappAccountsData?.accounts?.length || 0) === 0 ? (
                      <div className="text-center py-8">
                        <MessageCircle className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                        <h3 className="text-sm font-semibold text-gray-900 mb-1">No WhatsApp Accounts</h3>
                        <p className="text-sm text-muted-foreground mb-4">Connect your WhatsApp account to get started</p>
                        <Button 
                          size="sm" 
                          onClick={() => {
                            setSelectedPlatform("whatsapp");
                            setConnectModalOpen(true);
                          }}
                        >
                          <Plus className="h-4 w-4 mr-2" />
                          Connect WhatsApp Account
                        </Button>
                      </div>
                    ) : (
                      <div className="space-y-3 pt-4">
                        {whatsappAccountsData?.accounts?.map((account) => (
                          <div
                            key={account.id}
                            className="flex items-center justify-between p-4 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                          >
                            <div 
                              className="flex items-center gap-3 flex-1 cursor-pointer"
                              onClick={() => setViewAccount({ 
                                platform: "WhatsApp", 
                                account: { 
                                  id: account.id, 
                                  name: account.phone_number, 
                                  handle: account.phone_number, 
                                  followers: 0 
                                } 
                              })}
                            >
                              <div className="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center">
                                <MessageCircle className="h-5 w-5 text-green-600" />
                              </div>
                              <div className="flex-1">
                                <div className="font-medium">{account.phone_number}</div>
                                <div className="text-sm text-muted-foreground">
                                  {account.name || "WhatsApp Account"} • {account.status}
                                </div>
                              </div>
                            </div>
                            <div className="flex gap-2">
                              <Button 
                                size="sm" 
                                variant="ghost"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setViewAccount({ 
                                    platform: "WhatsApp", 
                                    account: { 
                                      id: account.id, 
                                      name: account.phone_number, 
                                      handle: account.phone_number, 
                                      followers: 0 
                                    } 
                                  });
                                }}
                              >
                                <Key className="h-4 w-4 mr-2" />
                                View Session
                              </Button>
                              <Button 
                                size="sm" 
                                variant="ghost"
                                className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setDeleteAccount({ 
                                    platform: "WhatsApp", 
                                    account: { 
                                      id: account.id, 
                                      name: account.phone_number, 
                                      handle: account.phone_number, 
                                      followers: 0 
                                    } 
                                  });
                                }}
                              >
                                <Trash2 className="h-4 w-4 mr-2" />
                                Delete
                              </Button>
                            </div>
                          </div>
                        ))}
                        <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                          <p className="text-xs text-blue-800">
                            <span className="font-semibold">Note:</span> You can only connect one WhatsApp account per user. To connect a different account, please remove the existing one first.
                          </p>
                        </div>
                      </div>
                    )}
                  </AccordionContent>
                </AccordionItem>
              </Accordion>
            </Card>
          </main>
        </div>
      </div>

      {/* View Modal - Discord Style with API Key */}
      <Dialog open={!!viewAccount} onOpenChange={() => setViewAccount(null)}>
        <DialogContent className="sm:max-w-3xl max-h-[90vh] p-0 overflow-y-auto">
          {viewAccount && (() => {
            const platform = platforms.find(p => p.name === viewAccount.platform) || { name: "Twitter", icon: Twitter, color: "text-sky-500", bgColor: "bg-sky-50" };
            const Icon = platform.icon;
            const getBannerColor = () => {
              switch(viewAccount.platform) {
                case "LinkedIn": return "bg-blue-600";
                case "Twitter": return "bg-gradient-to-r from-sky-400 to-blue-500";
                case "Instagram": return "bg-gradient-to-br from-purple-600 via-pink-600 to-orange-500";
                case "Facebook": return "bg-blue-700";
                case "WhatsApp": return "bg-green-600";
                default: return "bg-primary";
              }
            };

            // Get the real Twitter account data if it's a Twitter account
            const twitterAccountData = viewAccount.platform === "Twitter" 
              ? twitterAccounts.find(acc => acc.id === viewAccount.account.id)
              : null;
            
            // Get the real WhatsApp account data if it's a WhatsApp account
            const whatsappAccountData = viewAccount.platform === "WhatsApp"
              ? whatsappAccountsData?.accounts?.find(acc => acc.id === viewAccount.account.id)
              : null;
            
            return (
              <>
                {/* Banner with faint platform logo */}
                <div className={`relative h-24 ${getBannerColor()}`}>
                  <div className="absolute inset-0 flex items-center justify-center opacity-10">
                    <Icon className="h-20 w-20 text-white" />
                  </div>
                </div>
                
                {/* Profile Picture */}
                <div className="relative px-6 pb-6">
                  <div className="absolute -top-12 left-6">
                    <div className={`h-24 w-24 rounded-full border-4 border-background flex items-center justify-center ${platform.bgColor}`}>
                      <Icon className={`h-12 w-12 ${platform.color}`} />
                    </div>
                  </div>
                  
                  {/* Content */}
                  <div className="pt-16 space-y-4">
                    <div>
                      <h2 className="text-2xl font-bold">{viewAccount.account.handle}</h2>
                      <p className="text-muted-foreground">{viewAccount.platform} Account</p>
                    </div>
                    
                    {twitterAccountData ? (
                      <>
                        <div className="border-t pt-4 space-y-4">
                          <div>
                            <Label className="text-sm font-medium">API Key (Token)</Label>
                            <div className="mt-2 flex gap-2">
                              <Input
                                value={twitterAccountData.token}
                                readOnly
                                className="font-mono text-xs flex-1"
                              />
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleCopyKey(twitterAccountData.token)}
                                className="shrink-0"
                              >
                                {copiedKey ? (
                                  <Check className="h-4 w-4 text-green-600" />
                                ) : (
                                  <Copy className="h-4 w-4" />
                                )}
                              </Button>
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                              Use this API key in the Authorization header as "Bearer {twitterAccountData.token.substring(0, 20)}..."
                            </p>
                          </div>

                          <div className="border-t pt-4">
                            <h4 className="text-sm font-semibold mb-3">Example Usage</h4>
                            <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
                              <pre className="text-xs text-gray-100 whitespace-pre-wrap break-all">
                                <code>{`curl -X POST http://localhost:8080/twitter/post \\
  -H "Authorization: Bearer ${twitterAccountData.token}" \\
  -H "Content-Type: application/json" \\
  -d '{"url": "https://twitter.com/user/status/123"}'`}</code>
                              </pre>
                            </div>
                          </div>

                          <div className="border-t pt-4">
                            <h4 className="text-sm font-semibold mb-3">Available Endpoints</h4>
                            <div className="space-y-2">
                              {[
                                { method: "POST", path: "/twitter/post", desc: "Get tweet data with media" },
                                { method: "POST", path: "/twitter/post/likes", desc: "Get users who liked a tweet" },
                                { method: "POST", path: "/twitter/post/quotes", desc: "Get quote tweets" },
                                { method: "POST", path: "/twitter/post/comments", desc: "Get tweet replies" },
                                { method: "POST", path: "/twitter/post/reposts", desc: "Get users who reposted" },
                              ].map((endpoint, i) => (
                                <div key={i} className="flex flex-col sm:flex-row sm:items-center gap-2 text-sm">
                                  <div className="flex items-center gap-2">
                                    <span className="px-2 py-0.5 text-xs font-semibold text-blue-700 bg-blue-100 rounded shrink-0">
                                      {endpoint.method}
                                    </span>
                                    <code className="text-xs font-mono break-all">{endpoint.path}</code>
                                  </div>
                                  <span className="text-muted-foreground text-xs">{endpoint.desc}</span>
                                </div>
                              ))}
                            </div>
                          </div>
                        </div>
                      </>
                    ) : whatsappAccountData ? (
                      <>
                        <div className="border-t pt-4 space-y-4">
                          <div>
                            <Label className="text-sm font-medium">Session ID</Label>
                            <div className="mt-2 flex gap-2">
                              <Input
                                value={whatsappAccountData.session_id}
                                readOnly
                                className="font-mono text-xs flex-1"
                              />
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleCopyKey(whatsappAccountData.session_id)}
                                className="shrink-0"
                              >
                                {copiedKey ? (
                                  <Check className="h-4 w-4 text-green-600" />
                                ) : (
                                  <Copy className="h-4 w-4" />
                                )}
                              </Button>
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                              Phone Number: {whatsappAccountData.phone_number}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              Status: <span className="text-green-600 font-medium">{whatsappAccountData.status}</span>
                            </p>
                          </div>

                          <div className="border-t pt-4">
                            <h4 className="text-sm font-semibold mb-3">Example Usage</h4>
                            <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
                              <pre className="text-xs text-gray-100 whitespace-pre-wrap break-all">
                                <code>{`curl -X POST http://localhost:8080/whatsapp/send-message \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "session_id": "${whatsappAccountData.session_id}",
    "phone": "1234567890@s.whatsapp.net",
    "message": "Hello from API!",
    "reply": false
  }'`}</code>
                              </pre>
                            </div>
                          </div>

                          <div className="border-t pt-4">
                            <h4 className="text-sm font-semibold mb-3">Available Endpoints</h4>
                            <div className="space-y-2">
                              {[
                                { method: "POST", path: "/whatsapp/send-message", desc: "Send WhatsApp message" },
                                { method: "GET", path: "/whatsapp/", desc: "Get connected WhatsApp accounts" },
                                { method: "GET", path: "/whatsapp/session-status/:id", desc: "Check session status" },
                              ].map((endpoint, i) => (
                                <div key={i} className="flex flex-col sm:flex-row sm:items-center gap-2 text-sm">
                                  <div className="flex items-center gap-2">
                                    <span className="px-2 py-0.5 text-xs font-semibold text-green-700 bg-green-100 rounded shrink-0">
                                      {endpoint.method}
                                    </span>
                                    <code className="text-xs font-mono break-all">{endpoint.path}</code>
                                  </div>
                                  <span className="text-muted-foreground text-xs">{endpoint.desc}</span>
                                </div>
                              ))}
                            </div>
                          </div>
                        </div>
                      </>
                    ) : null}

                    <div className="flex gap-2 pt-4 border-t">
                      <Button onClick={() => setViewAccount(null)} className="flex-1">
                        Close
                      </Button>
                    </div>
                  </div>
                </div>
              </>
            );
          })()}
        </DialogContent>
      </Dialog>

      {/* Edit Modal */}
      <Dialog open={!!editAccount} onOpenChange={() => setEditAccount(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Edit Account Handle</DialogTitle>
            <DialogDescription>
              Update the Telegram handle for this account.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label htmlFor="handle" className="text-sm font-medium">
                Telegram Handle
              </label>
              <Input
                id="handle"
                value={editHandle}
                onChange={(e) => setEditHandle(e.target.value)}
                placeholder="@username"
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                Handle must start with @
              </p>
            </div>
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => setEditAccount(null)}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit}>
              Save Changes
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Alert */}
      <AlertDialog open={!!deleteAccount} onOpenChange={() => setDeleteAccount(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              Do you want to delete this connected account? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>Yes, Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Connect Account Modal */}
      <Dialog open={connectModalOpen} onOpenChange={setConnectModalOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Connect Social Media Account</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="platform">Platform</Label>
              <Select value={selectedPlatform} onValueChange={handlePlatformChange}>
                <SelectTrigger id="platform">
                  <SelectValue placeholder="Select a platform" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="twitter">
                    <div className="flex items-center gap-2">
                      <Twitter className="h-4 w-4" />
                      <span>Twitter</span>
                    </div>
                  </SelectItem>
                  <SelectItem value="whatsapp">
                    <div className="flex items-center gap-2">
                      <MessageCircle className="h-4 w-4" />
                      <span>WhatsApp</span>
                      {/* Allow multiple WhatsApp accounts, no limit message */}
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              {/* Removed WhatsApp account limit info message */}
            </div>

            {selectedPlatform === "twitter" && (
              <>
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    placeholder="Enter your Twitter username"
                    {...register("username")}
                    disabled={isConnecting}
                  />
                  {errors.username && (
                    <p className="text-sm text-destructive">{errors.username.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="Enter your Twitter password"
                    {...register("password")}
                    disabled={isConnecting}
                  />
                  {errors.password && (
                    <p className="text-sm text-destructive">{errors.password.message}</p>
                  )}
                </div>
              </>
            )}

            {selectedPlatform === "whatsapp" && (
              <div className="space-y-4">
                {!whatsappQRCode ? (
                  <div className="text-center py-4">
                    <p className="text-sm text-muted-foreground mb-4">
                      Click the button below to generate a QR code for WhatsApp login
                    </p>
                    <Button 
                      type="button"
                      onClick={generateWhatsAppQR}
                      disabled={isConnecting}
                      className="w-full"
                    >
                      {isConnecting ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Generating QR Code...
                        </>
                      ) : (
                        <>
                          <MessageCircle className="mr-2 h-4 w-4" />
                          Generate QR Code
                        </>
                      )}
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="bg-white p-4 rounded-lg border-2 border-dashed border-gray-300 flex items-center justify-center">
                      <img
                        src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(whatsappQRCode)}`}
                        alt="WhatsApp QR Code"
                        className="w-48 h-48"
                      />
                    </div>
                    <div className="text-center space-y-2">
                      <p className="text-sm font-medium">
                        {whatsappStatus === "pending" && "Scan QR code with WhatsApp"}
                        {whatsappStatus === "authenticated" && "✓ Authenticated Successfully!"}
                        {whatsappStatus === "failed" && "✗ Authentication Failed"}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        Open WhatsApp on your phone → Settings → Linked Devices → Link a Device
                      </p>
                      {whatsappStatus === "pending" && (
                        <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
                          <Loader2 className="h-3 w-3 animate-spin" />
                          Waiting for scan...
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setConnectModalOpen(false);
                  resetForm();
                  setWhatsappQRCode(null);
                  setWhatsappSessionId(null);
                  setWhatsappStatus("pending");
                  if (pollingIntervalRef.current) {
                    clearInterval(pollingIntervalRef.current);
                  }
                }}
                disabled={isConnecting}
              >
                Cancel
              </Button>
              {selectedPlatform === "twitter" && (
                <Button type="submit" disabled={isConnecting}>
                  {isConnecting ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Connecting...
                    </>
                  ) : (
                    "Connect Account"
                  )}
                </Button>
              )}
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
};

export default ApiManagement;