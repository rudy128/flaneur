import { SidebarProvider } from "@/components/ui/sidebar";
import { DashboardSidebar } from "@/components/dashboard/DashboardSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { ConnectedProfiles } from "@/components/dashboard/ConnectedProfiles";
import { ApiUsage } from "@/components/dashboard/ApiUsage";
import { ApiLogs } from "@/components/dashboard/ApiLogs";

const Dashboard = () => {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <DashboardSidebar />
        
        <div className="flex-1 flex flex-col">
          <DashboardHeader />
          
          <main className="flex-1 p-6 space-y-6">
            <div>
              <h1 className="text-3xl font-bold tracking-tight mb-2">Dashboard</h1>
              <p className="text-muted-foreground">
                Monitor your API usage and manage connected profiles
              </p>
            </div>

            <ConnectedProfiles />
            
            <div className="grid gap-6 lg:grid-cols-2">
              <ApiUsage />
            </div>


            <ApiLogs />

          </main>
        </div>
      </div>
    </SidebarProvider>
  );
};

export default Dashboard;
