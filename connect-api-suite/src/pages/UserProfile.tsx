import { useState } from "react";
import { SidebarProvider } from "@/components/ui/sidebar";
import { DashboardSidebar } from "@/components/dashboard/DashboardSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { toast } from "@/hooks/use-toast";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authApi } from "@/lib/api";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { changePasswordFormSchema, type ChangePasswordFormData } from "@/lib/schemas";
import { Loader2 } from "lucide-react";

const UserProfile = () => {
  const queryClient = useQueryClient();
  const token = localStorage.getItem("jwt_token");

  // Fetch user profile
  const { data: profile, isLoading } = useQuery({
    queryKey: ["profile"],
    queryFn: () => authApi.getProfile(token || ""),
    enabled: !!token,
  });

  // Password change form
  const { register, handleSubmit, formState: { errors }, reset } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordFormSchema),
  });

  // Password change mutation
  const changePasswordMutation = useMutation({
    mutationFn: (data: ChangePasswordFormData) => 
      authApi.changePassword(token || "", {
        current_password: data.current_password,
        new_password: data.new_password,
      }),
    onSuccess: () => {
      toast({
        title: "Password Changed",
        description: "Your password has been updated successfully.",
      });
      reset();
    },
    onError: (error: Error) => {
      const apiError = error as { data?: { error?: string } };
      toast({
        title: "Error",
        description: apiError.data?.error || "Failed to change password. Please try again.",
        variant: "destructive",
      });
    },
  });

  const onSubmit = (data: ChangePasswordFormData) => {
    changePasswordMutation.mutate(data);
  };

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      year: "numeric",
    });
  };

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <DashboardSidebar />
        
        <div className="flex-1 flex flex-col">
          <DashboardHeader />
          
          <main className="flex-1 p-6 space-y-6">
            <div>
              <h1 className="text-3xl font-bold tracking-tight mb-2">User Profile</h1>
              <p className="text-muted-foreground">
                Manage your account information and settings
              </p>
            </div>

            {/* Header Section */}
            <Card className="p-6">
              {isLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
                </div>
              ) : profile ? (
                <div className="flex justify-between pr-48">

                <div className="flex items-start gap-6">
                  <Avatar className="h-24 w-24">
                    <AvatarImage src="https://github.com/shadcn.png" />
                    <AvatarFallback>{getInitials(profile.name)}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h2 className="text-2xl font-bold">{profile.name}</h2>
                      <Badge variant="secondary">Active</Badge>
                    </div>
                    <p className="text-muted-foreground mb-4">{profile.email}</p>
                    <div className="flex gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">Member since</span>
                        <span className="ml-2 font-medium">{formatDate(profile.created_at)}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <Card className="p-4">
                  <p className="text-sm font-medium text-muted-foreground">Connected Accounts</p>
                  <div className="flex items-center justify-between">
                  <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
                    <svg
                      className="h-6 w-6 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                      />
                    </svg>
                  </div>
                  <div>
                    <p className="text-3xl font-bold mt-2">{profile?.accounts_connected || 0}</p>
                  </div>
                </div>
                </Card>
                </div>
              ) : (
                <p className="text-center text-muted-foreground">Failed to load profile</p>
              )}
            </Card>

            {/* Settings Section */}
            <Card className="p-6">
              <h3 className="text-xl font-semibold mb-6">Account Settings</h3>
              
              <div className="space-y-6">
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                  <h4 className="text-sm font-medium">Change Password</h4>
                  <div className="grid gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="current_password">Current Password</Label>
                      <Input
                        id="current_password"
                        type="password"
                        {...register("current_password")}
                        disabled={changePasswordMutation.isPending}
                      />
                      {errors.current_password && (
                        <p className="text-sm text-destructive">{errors.current_password.message}</p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="new_password">New Password</Label>
                      <Input
                        id="new_password"
                        type="password"
                        {...register("new_password")}
                        disabled={changePasswordMutation.isPending}
                      />
                      {errors.new_password && (
                        <p className="text-sm text-destructive">{errors.new_password.message}</p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="confirm_password">Confirm New Password</Label>
                      <Input
                        id="confirm_password"
                        type="password"
                        {...register("confirm_password")}
                        disabled={changePasswordMutation.isPending}
                      />
                      {errors.confirm_password && (
                        <p className="text-sm text-destructive">{errors.confirm_password.message}</p>
                      )}
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <Button type="submit" disabled={changePasswordMutation.isPending}>
                      {changePasswordMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Changing Password...
                        </>
                      ) : (
                        "Change Password"
                      )}
                    </Button>
                  </div>
                </form>
              </div>
            </Card>
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
};

export default UserProfile;
