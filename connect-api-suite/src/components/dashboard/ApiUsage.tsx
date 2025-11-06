import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Bar, BarChart, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { useQuery } from "@tanstack/react-query";
import { logsApi } from "@/lib/api";
import { Loader2 } from "lucide-react";
import { apiStatsSchema, type ApiStats } from "@/lib/schemas";

export function ApiUsage() {
  const token = localStorage.getItem("jwt_token");

  const { data: stats, isLoading, error } = useQuery({
    queryKey: ["api-stats"],
    queryFn: () => logsApi.getStats(token!),
    enabled: !!token,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  if (isLoading) {
    return (
      <Card className="col-span-full">
        <CardHeader>
          <CardTitle>API Usage Analytics</CardTitle>
          <CardDescription>Visualizing your API usage patterns</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-[400px]">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error || !stats) {
    return (
      <Card className="col-span-full">
        <CardHeader>
          <CardTitle>API Usage Analytics</CardTitle>
          <CardDescription>Visualizing your API usage patterns</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-[400px] text-muted-foreground">
            No usage data available yet. Make some API calls to see analytics.
          </div>
        </CardContent>
      </Card>
    );
  }

  // Type assertion for stats
  const validatedStats = stats as ApiStats;

  // Prepare data for endpoint usage bar chart
  const endpointData = Array.isArray(validatedStats.calls_by_endpoint)
    ? validatedStats.calls_by_endpoint.map((item) => ({
        endpoint: (item.Endpoint || "").replace("/twitter/post", "").replace("/", "") || "tweets",
        count: item.Count || 0,
      }))
    : [];

  // Prepare data for success/failure pie chart
  const successData = [
    { name: "Successful", value: validatedStats.successful_calls || 0, color: "hsl(var(--chart-1))" },
    { name: "Failed", value: validatedStats.failed_calls || 0, color: "hsl(var(--chart-2))" },
  ].filter(item => item.value > 0);

  const totalCalls = validatedStats.total_calls || 0;
  const successRate = totalCalls > 0 ? (((validatedStats.successful_calls || 0) / totalCalls) * 100) : 0;

  return (
    <Card className="col-span-full">
      <CardHeader>
        <CardTitle>API Usage Analytics</CardTitle>
        <CardDescription>Visualizing your API usage patterns and performance</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-6 md:grid-cols-2">
          {/* Endpoint Usage Bar Chart */}
          <div>
            <h3 className="text-sm font-medium mb-4">Calls by Endpoint</h3>
            {endpointData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={endpointData}>
                  <XAxis 
                    dataKey="endpoint" 
                    stroke="#888888"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    stroke="#888888"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    allowDecimals={false}
                  />
                  <Tooltip 
                    contentStyle={{ 
                      backgroundColor: "hsl(var(--background))",
                      border: "1px solid hsl(var(--border))",
                      borderRadius: "8px"
                    }}
                  />
                  <Bar dataKey="count" fill="hsl(var(--primary))" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex items-center justify-center h-[250px] text-sm text-muted-foreground">
                No endpoint data yet
              </div>
            )}
          </div>

          {/* Success/Failure Pie Chart */}
          <div>
            <h3 className="text-sm font-medium mb-4">Success Rate</h3>
            {successData.length > 0 ? (
              <div className="flex items-center justify-between">
                <ResponsiveContainer width="60%" height={250}>
                  <PieChart>
                    <Pie
                      data={successData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={80}
                      paddingAngle={5}
                      dataKey="value"
                    >
                      {successData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip 
                      contentStyle={{ 
                        backgroundColor: "hsl(var(--background))",
                        border: "1px solid hsl(var(--border))",
                        borderRadius: "8px"
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
                <div className="space-y-4">
                  {successData.map((entry) => (
                    <div key={entry.name} className="flex items-center gap-2">
                      <div 
                        className="w-3 h-3 rounded-full" 
                        style={{ backgroundColor: entry.color }}
                      />
                      <div>
                        <p className="text-sm font-medium">{entry.name}</p>
                        <p className="text-2xl font-bold">{entry.value}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center h-[250px] text-sm text-muted-foreground">
                No call data yet
              </div>
            )}
          </div>
        </div>

        {/* Summary Stats */}
        <div className="mt-6 grid gap-4 md:grid-cols-4">
          <div className="p-4 rounded-lg border bg-card">
            <p className="text-sm font-medium text-muted-foreground">Total Calls</p>
            <p className="text-2xl font-bold">{totalCalls}</p>
          </div>
          <div className="p-4 rounded-lg border bg-card">
            <p className="text-sm font-medium text-muted-foreground">Success Rate</p>
            <p className="text-2xl font-bold">{successRate.toFixed(1)}%</p>
          </div>
          <div className="p-4 rounded-lg border bg-card">
            <p className="text-sm font-medium text-muted-foreground">Avg Response Time</p>
            <p className="text-2xl font-bold">{validatedStats.average_response_time || 0}ms</p>
          </div>
          <div className="p-4 rounded-lg border bg-card">
            <p className="text-sm font-medium text-muted-foreground">Endpoints Used</p>
            <p className="text-2xl font-bold">{endpointData.length}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
