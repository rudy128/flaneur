import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
  RefreshCw,
  Filter,
  Download,
} from "lucide-react";
import { logsApi } from "@/lib/api";
import type { ApiCallLog } from "@/lib/schemas";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export const ApiLogs = () => {
  const [logs, setLogs] = useState<ApiCallLog[]>([]);
  const [filteredLogs, setFilteredLogs] = useState<ApiCallLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState<"all" | "success" | "error">("all");
  const [filterEndpoint, setFilterEndpoint] = useState<string>("all");

  const fetchLogs = async () => {
    setLoading(true);
    setError(null);
    try {
      const token = localStorage.getItem("jwt_token");
      if (!token) {
        throw new Error("No authentication token found");
      }

      const data = await logsApi.getLogs(token, 100);
      setLogs(data);
      setFilteredLogs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch logs");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, []);

  useEffect(() => {
    let filtered = logs;

    // Filter by status
    if (filterStatus === "success") {
      filtered = filtered.filter((log) => log.success);
    } else if (filterStatus === "error") {
      filtered = filtered.filter((log) => !log.success);
    }

    // Filter by endpoint
    if (filterEndpoint !== "all") {
      filtered = filtered.filter((log) => log.endpoint === filterEndpoint);
    }

    setFilteredLogs(filtered);
  }, [logs, filterStatus, filterEndpoint]);

  const getEndpoints = () => {
    const endpoints = new Set(logs.map((log) => log.endpoint));
    return Array.from(endpoints);
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  const formatResponseTime = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const exportLogs = () => {
    const csv = [
      ["Timestamp", "Endpoint", "Method", "Status", "Response Time", "Success", "Error"],
      ...filteredLogs.map((log) => [
        log.created_at,
        log.endpoint,
        log.method,
        log.status_code.toString(),
        log.response_time.toString(),
        log.success ? "Yes" : "No",
        log.error_message,
      ]),
    ]
      .map((row) => row.join(","))
      .join("\n");

    const blob = new Blob([csv], { type: "text/csv" });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `api-logs-${new Date().toISOString()}.csv`;
    a.click();
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Recent Call Logs
            </CardTitle>
            <CardDescription>
              View detailed logs of all your API calls
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={fetchLogs}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={exportLogs}
              disabled={filteredLogs.length === 0}
            >
              <Download className="h-4 w-4 mr-2" />
              Export
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {/* Filters */}
        <div className="flex gap-4 mb-4">
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <Select value={filterStatus} onValueChange={(value) => setFilterStatus(value as "all" | "success" | "error")}>
              <SelectTrigger className="w-[140px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="error">Errors</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Select value={filterEndpoint} onValueChange={setFilterEndpoint}>
            <SelectTrigger className="w-[200px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Endpoints</SelectItem>
              {getEndpoints().map((endpoint) => (
                <SelectItem key={endpoint} value={endpoint}>
                  {endpoint}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Logs Table */}
        {error ? (
          <div className="text-center py-8 text-destructive">
            <XCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>{error}</p>
            <Button onClick={fetchLogs} className="mt-4" variant="outline">
              Try Again
            </Button>
          </div>
        ) : loading ? (
          <div className="text-center py-8">
            <RefreshCw className="h-12 w-12 mx-auto mb-4 animate-spin opacity-50" />
            <p className="text-muted-foreground">Loading logs...</p>
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>No API calls logged yet</p>
          </div>
        ) : (
          <ScrollArea className="h-fit max-h-[500px]">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Endpoint</TableHead>
                  <TableHead>Method</TableHead>
                  <TableHead>Username</TableHead>
                  <TableHead>Response Time</TableHead>
                  <TableHead>Code</TableHead>
                  <TableHead>Error</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredLogs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell>
                      {log.success ? (
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      ) : (
                        <XCircle className="h-4 w-4 text-red-500" />
                      )}
                    </TableCell>
                    <TableCell className="text-sm">
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3 opacity-50" />
                        {formatTime(log.created_at)}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {log.endpoint}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{log.method}</Badge>
                    </TableCell>
                    <TableCell className="font-medium">
                      {log.twitter_username || "-"}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={log.response_time < 1000 ? "secondary" : "outline"}
                      >
                        {formatResponseTime(log.response_time)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          log.status_code >= 200 && log.status_code < 300
                            ? "default"
                            : "destructive"
                        }
                      >
                        {log.status_code}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-xs text-muted-foreground">
                      {log.error_message || "-"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        )}

        {/* Summary */}
        <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
          <div>
            Showing {filteredLogs.length} of {logs.length} logs
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <CheckCircle2 className="h-4 w-4 text-green-500" />
              {filteredLogs.filter((log) => log.success).length} successful
            </div>
            <div className="flex items-center gap-1">
              <XCircle className="h-4 w-4 text-red-500" />
              {filteredLogs.filter((log) => !log.success).length} failed
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
