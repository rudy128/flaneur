import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const recentCalls = [
  {
    endpoint: "/linkedin/profile",
    status: "200",
    responseTime: "124ms",
    timestamp: "2 min ago",
  },
  {
    endpoint: "/twitter/tweet",
    status: "200",
    responseTime: "98ms",
    timestamp: "5 min ago",
  },
  {
    endpoint: "/instagram/post",
    status: "200",
    responseTime: "156ms",
    timestamp: "12 min ago",
  },
  {
    endpoint: "/whatsapp/send",
    status: "429",
    responseTime: "45ms",
    timestamp: "15 min ago",
  },
  {
    endpoint: "/linkedin/connections",
    status: "200",
    responseTime: "210ms",
    timestamp: "23 min ago",
  },
];

export function RecentCalls() {
  return (
    <Card className="col-span-full">
      <CardHeader>
        <CardTitle>Recent API Calls</CardTitle>
        <CardDescription>Latest requests to your API endpoints</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Endpoint</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Response Time</TableHead>
              <TableHead className="text-right">Timestamp</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {recentCalls.map((call, index) => (
              <TableRow key={index}>
                <TableCell className="font-mono text-sm">{call.endpoint}</TableCell>
                <TableCell>
                  <Badge variant={call.status === "200" ? "default" : "destructive"}>
                    {call.status}
                  </Badge>
                </TableCell>
                <TableCell className="text-muted-foreground">{call.responseTime}</TableCell>
                <TableCell className="text-right text-muted-foreground">
                  {call.timestamp}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
