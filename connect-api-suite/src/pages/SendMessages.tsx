import { useState, useEffect, useMemo } from "react";
import { Upload, Send, FileText, X, Download, Plus, UserPlus, CheckCircle2, XCircle, Clock, RefreshCw, ChevronUp, ChevronDown, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Slider } from "@/components/ui/slider";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useQuery } from "@tanstack/react-query";
import { whatsappApi } from "@/lib/api";
import { SidebarProvider } from "@/components/ui/sidebar";
import { DashboardSidebar } from "@/components/dashboard/DashboardSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface Contact {
  phone: string;
  name?: string;
  message?: string; // Individual message for this contact
}

interface MessageLog {
  id: string;
  user_id: string;
  session_id: string;
  recipient_phone: string;
  recipient_name?: string;
  message: string;
  message_type: string;
  status: string;
  scheduled_at?: string;
  sent_at?: string;
  error_message?: string;
  batch_id?: string;
  sequence_number?: number;
  retry_count?: number;
  created_at: string;
  updated_at: string;
}

export default function SendMessages() {
  const [message, setMessage] = useState("");
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const [selectedSession, setSelectedSession] = useState<string>("");
  const [useIndividualMessages, setUseIndividualMessages] = useState(false);
  const [manualPhone, setManualPhone] = useState("");
  const [manualName, setManualName] = useState("");
  const [countryCode, setCountryCode] = useState("+91"); // Default to India
  const [inputMode, setInputMode] = useState<"file" | "manual">("file");
  
  // Scheduling state - Random by default
  const [enableScheduling, setEnableScheduling] = useState(false);
  const [randomDelayMin, setRandomDelayMin] = useState(10); // Default 10 seconds
  const [randomDelayMax, setRandomDelayMax] = useState(60); // Default 60 seconds
  
  // Alert Dialog state
  const [showResultDialog, setShowResultDialog] = useState(false);
  const [resultType, setResultType] = useState<"success" | "error">("success");
  const [resultMessage, setResultMessage] = useState("");
  const [successCount, setSuccessCount] = useState(0);
  const [failCount, setFailCount] = useState(0);

  // Message History state
  const [showMessageHistory, setShowMessageHistory] = useState(false);

  // Country codes with flags
  const countryCodes = [
    { code: "+1", country: "US/Canada", flag: "🇺🇸" },
    { code: "+44", country: "UK", flag: "🇬🇧" },
    { code: "+91", country: "India", flag: "🇮🇳" },
    { code: "+86", country: "China", flag: "🇨🇳" },
    { code: "+81", country: "Japan", flag: "🇯🇵" },
    { code: "+49", country: "Germany", flag: "🇩🇪" },
    { code: "+33", country: "France", flag: "🇫🇷" },
    { code: "+61", country: "Australia", flag: "🇦🇺" },
    { code: "+55", country: "Brazil", flag: "🇧🇷" },
    { code: "+7", country: "Russia", flag: "🇷🇺" },
    { code: "+82", country: "South Korea", flag: "🇰🇷" },
    { code: "+34", country: "Spain", flag: "🇪🇸" },
    { code: "+39", country: "Italy", flag: "🇮🇹" },
    { code: "+52", country: "Mexico", flag: "🇲🇽" },
    { code: "+27", country: "South Africa", flag: "🇿🇦" },
    { code: "+971", country: "UAE", flag: "🇦🇪" },
    { code: "+966", country: "Saudi Arabia", flag: "🇸🇦" },
    { code: "+65", country: "Singapore", flag: "🇸🇬" },
    { code: "+60", country: "Malaysia", flag: "🇲🇾" },
    { code: "+62", country: "Indonesia", flag: "🇮🇩" },
    { code: "+63", country: "Philippines", flag: "🇵🇭" },
    { code: "+66", country: "Thailand", flag: "🇹🇭" },
    { code: "+84", country: "Vietnam", flag: "🇻🇳" },
    { code: "+92", country: "Pakistan", flag: "🇵🇰" },
    { code: "+880", country: "Bangladesh", flag: "🇧🇩" },
    { code: "+94", country: "Sri Lanka", flag: "🇱🇰" },
    { code: "+977", country: "Nepal", flag: "🇳🇵" },
  ];

  // Get JWT token from localStorage (same key as API Management page)
  const token = localStorage.getItem("jwt_token");

  // Debug token
  useEffect(() => {
    console.log("=== TOKEN DEBUG ===");
    console.log("Token exists:", !!token);
    console.log("Token length:", token?.length || 0);
    console.log("Token preview:", token?.substring(0, 50) + "...");
    if (token && token.length < 50) {
      console.warn("⚠️ WARNING: Token seems too short! Length:", token.length);
      console.log("Full token:", token);
    }
  }, [token]);

  // Fetch WhatsApp sessions
  const { data: whatsappAccountsData, isLoading: loadingSessions, error: sessionsError } = useQuery({
    queryKey: ["whatsapp-accounts"],
    queryFn: () => whatsappApi.getAccounts(token || ""),
    enabled: !!token,
  });

  const whatsappSessions = useMemo(() => {
    return (whatsappAccountsData?.accounts || []) as Array<{
      id: string;
      session_id: string;
      phone_number: string;
      name?: string;
      status: string;
      created_at: string;
    }>;
  }, [whatsappAccountsData]);

  // Debug logging
  useEffect(() => {
    console.log("=== WhatsApp Sessions Debug ===");
    console.log("Raw Data:", whatsappAccountsData);
    console.log("Sessions Array:", whatsappSessions);
    console.log("Sessions Count:", whatsappSessions.length);
    console.log("Selected Session:", selectedSession);
    console.log("Loading:", loadingSessions);
    console.log("Error:", sessionsError);
  }, [whatsappAccountsData, whatsappSessions, selectedSession, loadingSessions, sessionsError]);

  // Fetch message logs with auto-refresh
  const { data: messageLogsData, refetch: refetchLogs } = useQuery({
    queryKey: ["message-logs", selectedSession],
    queryFn: () => whatsappApi.getMessageLogs(token || "", { limit: 50, offset: 0 }),
    enabled: !!token && showMessageHistory,
    refetchInterval: 10000, // Auto-refresh every 10 seconds
  });

  // Fetch message log stats
  const { data: messageStatsData } = useQuery({
    queryKey: ["message-log-stats"],
    queryFn: () => whatsappApi.getMessageLogStats(token || ""),
    enabled: !!token && showMessageHistory,
    refetchInterval: 10000,
  });

  // Auto-select first session
  useEffect(() => {
    if (whatsappSessions.length > 0 && !selectedSession) {
      console.log("Auto-selecting session:", whatsappSessions[0].session_id);
      setSelectedSession(whatsappSessions[0].session_id);
    }
  }, [whatsappSessions, selectedSession]);

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setFileError(null);
    setSelectedFile(file);

    const fileExtension = file.name.split(".").pop()?.toLowerCase();
    
    if (fileExtension === "csv") {
      parseCSV(file);
    } else if (fileExtension === "xlsx" || fileExtension === "xls") {
      parseExcel(file);
    } else {
      setFileError("Please upload a CSV or XLSX file");
      setSelectedFile(null);
    }
  };

  const parseCSV = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      const lines = text.split("\n").filter(line => line.trim());
      
      // Skip header if exists
      const hasHeader = lines[0]?.toLowerCase().includes("phone") || lines[0]?.toLowerCase().includes("name");
      const dataLines = hasHeader ? lines.slice(1) : lines;
      
      const parsedContacts: Contact[] = dataLines.map(line => {
        const [phone, name, individualMessage] = line.split(",").map(cell => cell.trim());
        // Remove all non-numeric characters from phone number
        const cleanPhone = phone.replace(/\D/g, '');
        return { 
          phone: cleanPhone, 
          name: name || undefined,
          message: individualMessage || undefined 
        };
      }).filter(contact => contact.phone);

      setContacts(parsedContacts);
      
      // Check if any contact has an individual message
      const hasIndividualMessages = parsedContacts.some(c => c.message);
      setUseIndividualMessages(hasIndividualMessages);
    };
    reader.onerror = () => {
      setFileError("Error reading CSV file");
    };
    reader.readAsText(file);
  };

  const parseExcel = async (file: File) => {
    try {
      // Dynamically import xlsx library
      const XLSX = await import("xlsx");
      
      const reader = new FileReader();
      reader.onload = (e) => {
        const data = new Uint8Array(e.target?.result as ArrayBuffer);
        const workbook = XLSX.read(data, { type: "array" });
        
        // Get first sheet
        const sheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[sheetName];
        
        // Convert to JSON
        const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 }) as unknown[][];
        
        // Check if has header
        const hasHeader = jsonData[0]?.some((cell: unknown) => 
          typeof cell === "string" && (cell.toLowerCase().includes("phone") || cell.toLowerCase().includes("name"))
        );
        
        const dataRows = hasHeader ? jsonData.slice(1) : jsonData;
        
        const parsedContacts: Contact[] = dataRows.map(row => {
          const [phone, name, individualMessage] = row;
          // Remove all non-numeric characters from phone number
          const phoneStr = phone?.toString() || "";
          const cleanPhone = phoneStr.replace(/\D/g, '');
          return { 
            phone: cleanPhone, 
            name: name?.toString() || undefined,
            message: individualMessage?.toString() || undefined
          };
        }).filter(contact => contact.phone);

        setContacts(parsedContacts);
        
        // Check if any contact has an individual message
        const hasIndividualMessages = parsedContacts.some(c => c.message);
        setUseIndividualMessages(hasIndividualMessages);
      };
      reader.onerror = () => {
        setFileError("Error reading Excel file");
      };
      reader.readAsArrayBuffer(file);
    } catch (error) {
      setFileError("Please install xlsx library: npm install xlsx");
      console.error(error);
    }
  };

  const removeContact = (index: number) => {
    setContacts(contacts.filter((_, i) => i !== index));
  };

  const clearAll = () => {
    setContacts([]);
    setSelectedFile(null);
    setFileError(null);
  };

  // Render WhatsApp formatted text as HTML
  const renderFormattedText = (text: string): JSX.Element => {
    // Split by lines to handle quotes
    const lines = text.split('\n');
    
    return (
      <>
        {lines.map((line, lineIndex) => {
          // Check if line starts with > (quote)
          const isQuote = line.trim().startsWith('>');
          let processedLine = isQuote ? line.trim().substring(1).trim() : line;
          
          // Process formatting: bold, italic, strikethrough, monospace
          // Bold: *text*
          processedLine = processedLine.replace(/\*([^*]+)\*/g, '<strong>$1</strong>');
          // Italic: _text_
          processedLine = processedLine.replace(/_([^_]+)_/g, '<em>$1</em>');
          // Strikethrough: ~text~
          processedLine = processedLine.replace(/~([^~]+)~/g, '<del>$1</del>');
          // Monospace: `text`
          processedLine = processedLine.replace(/`([^`]+)`/g, '<code class="bg-gray-200 px-1 rounded text-sm font-mono">$1</code>');
          
          return (
            <div key={lineIndex}>
              {isQuote ? (
                <div className="border-l-4 border-green-500 pl-3 py-1 my-1 bg-green-50/50 italic text-gray-700">
                  <span dangerouslySetInnerHTML={{ __html: processedLine }} />
                </div>
              ) : (
                <span dangerouslySetInnerHTML={{ __html: processedLine || '<br/>' }} />
              )}
            </div>
          );
        })}
      </>
    );
  };

  // Format phone number input (auto-format as user types)
  const formatPhoneNumber = (value: string) => {
    // Remove all non-numeric characters
    const numbers = value.replace(/\D/g, '');
    
    // Format based on length
    if (numbers.length <= 3) {
      return numbers;
    } else if (numbers.length <= 6) {
      return `${numbers.slice(0, 3)}-${numbers.slice(3)}`;
    } else if (numbers.length <= 10) {
      return `${numbers.slice(0, 3)}-${numbers.slice(3, 6)}-${numbers.slice(6)}`;
    } else {
      // Limit to 10 digits
      return `${numbers.slice(0, 3)}-${numbers.slice(3, 6)}-${numbers.slice(6, 10)}`;
    }
  };

  const handlePhoneInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatPhoneNumber(e.target.value);
    setManualPhone(formatted);
  };

  // Helper function to show result dialog
  const showDialog = (type: "success" | "error", message: string, success: number = 0, failed: number = 0) => {
    setResultType(type);
    setResultMessage(message);
    setSuccessCount(success);
    setFailCount(failed);
    setShowResultDialog(true);
  };

  const addManualContact = () => {
    // Remove formatting to get raw numbers
    const rawPhone = manualPhone.replace(/\D/g, '');
    
    if (!rawPhone) {
      showDialog("error", "Please enter a phone number");
      return;
    }

    // Combine country code with phone number (remove the + from country code for storage)
    const fullPhone = countryCode.replace('+', '') + rawPhone;

    const newContact: Contact = {
      phone: fullPhone,
      name: manualName.trim() || undefined,
    };

    setContacts([...contacts, newContact]);
    setManualPhone("");
    setManualName("");
  };

  const handleSendMessages = async ()=> {
    // Validate based on message mode
    if (!useIndividualMessages && !message.trim()) {
      showDialog("error", "Please enter a message");
      return;
    }
    
    if (contacts.length === 0) {
      showDialog("error", "Please add contacts to send messages to");
      return;
    }

    if (!selectedSession) {
      showDialog("error", "Please select a WhatsApp session");
      return;
    }

    if (!token) {
      showDialog("error", "Please login to send messages");
      return;
    }

    setSending(true);
    
    try {
      // If scheduling is enabled, use the bulk send API
      if (enableScheduling) {
        // Generate random delays for each message
        const messagesWithDelays = contacts.map((contact, index) => {
          const phone = contact.phone.includes("@") ? contact.phone : `${contact.phone}@s.whatsapp.net`;
          
          // Use individual message if available, otherwise use the common message
          let messageToSend = useIndividualMessages && contact.message 
            ? contact.message 
            : message;
          
          // Replace {name} placeholder with contact's name
          if (messageToSend.includes("{name}")) {
            const contactName = contact.name || contact.phone;
            messageToSend = messageToSend.replace(/\{name\}/gi, contactName);
          }
          
          return {
            recipient: phone,
            message: messageToSend,
            randomDelay: Math.floor(Math.random() * (randomDelayMax - randomDelayMin + 1)) + randomDelayMin
          };
        });

        // Calculate cumulative delays
        // First message: delay = 0 (sent immediately)
        // Second message: delay = 0 + random_delay_1
        // Third message: delay = 0 + random_delay_1 + random_delay_2
        // And so on...
        let cumulativeDelay = 0;
        const messagesWithCumulativeDelays = messagesWithDelays.map((msg, index) => {
          let messageDelay = 0;
          if (index === 0) {
            // First message sent immediately
            messageDelay = 0;
          } else {
            // Add this message's random delay to cumulative total
            cumulativeDelay += msg.randomDelay;
            messageDelay = cumulativeDelay;
          }
          
          return {
            recipient: msg.recipient,
            message: msg.message,
            delay_seconds: messageDelay
          };
        });

        console.log('Sending to backend:', {
          session_name: selectedSession,
          messages: messagesWithCumulativeDelays,
          messageCount: messagesWithCumulativeDelays.length
        });

        // Call the bulk send API with individual delays
        const response = await fetch(`${import.meta.env.VITE_API_URL}/whatsapp/send-bulk`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify({
            session_name: selectedSession,
            messages: messagesWithCumulativeDelays
          })
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
          console.error('Backend error:', errorData);
          throw new Error(errorData.error || 'Failed to schedule messages');
        }

        const result = await response.json();
        
        showDialog(
          "success",
          `${contacts.length} messages scheduled successfully! Batch ID: ${result.batch_id}`,
          contacts.length,
          0
        );
        
        // Clear after successful scheduling
        setMessage("");
        clearAll();
      } else {
        // Original immediate send logic
        let successCount = 0;
        let failCount = 0;

        // Send messages to each contact
        for (const contact of contacts) {
          try {
            const phone = contact.phone.includes("@") ? contact.phone : `${contact.phone}@s.whatsapp.net`;
            
            // Use individual message if available, otherwise use the common message
            let messageToSend = useIndividualMessages && contact.message 
              ? contact.message 
              : message;
            
            // Replace {name} placeholder with contact's name
            if (messageToSend.includes("{name}")) {
              const contactName = contact.name || contact.phone;
              messageToSend = messageToSend.replace(/\{name\}/gi, contactName);
            }
            
            if (!messageToSend.trim()) {
              console.warn(`Skipping ${contact.phone} - no message`);
              failCount++;
              continue;
            }
            
            await whatsappApi.sendMessage(
              token,
              selectedSession,
              phone,
              messageToSend,
              false
            );
            successCount++;
            
            // Add small delay to avoid rate limiting
            await new Promise(resolve => setTimeout(resolve, 500));
          } catch (error) {
            console.error(`Failed to send to ${contact.phone}:`, error);
            failCount++;
          }
        }
        
        // Show result dialog
        if (successCount > 0 || failCount > 0) {
          const totalMessages = successCount + failCount;
          showDialog(
            failCount === 0 ? "success" : (successCount > 0 ? "success" : "error"),
            failCount === 0 
              ? `All ${successCount} messages sent successfully!`
              : (successCount > 0 
                  ? `Messages sent with some failures`
                  : `Failed to send all messages`),
            successCount,
            failCount
          );
        }
        
        // Clear after successful send
        if (successCount > 0) {
          setMessage("");
          clearAll();
        }
      }
    } catch (error) {
      console.error("Error sending messages:", error);
      showDialog("error", "Error sending messages. Please try again.");
    } finally {
      setSending(false);
    }
  };

  const downloadTemplate = () => {
    const csvContent = "phone,name,message\n1234567890,John Doe,Hello John! This is your custom message.\n9876543210,Jane Smith,Hi Jane! Your personalized message here.";
    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "contacts_template.csv";
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <DashboardSidebar />
        
        <div className="flex-1 flex flex-col">
          <DashboardHeader />
          
          <main className="flex-1 p-6 bg-gray-50">
            <div className="max-w-7xl mx-auto space-y-6">
              {/* Header */}
              <div className="mb-8">
                <h1 className="text-3xl font-bold tracking-tight mb-2">Send Messages</h1>
                <p className="text-muted-foreground">
                  Import contacts and send bulk WhatsApp messages
                </p>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Left Column - Import & Contacts */}
                <div className="space-y-6">
                  {/* Import Card */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Upload className="h-5 w-5" />
                        Add Contacts
                      </CardTitle>
                      <CardDescription>
                        Import from file or add manually
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Tabs defaultValue="file" className="w-full">
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="file">Upload File</TabsTrigger>
                    <TabsTrigger value="manual">Manual Entry</TabsTrigger>
                  </TabsList>
                  
                  {/* File Upload Tab */}
                  <TabsContent value="file" className="space-y-4">
                    <div>
                      <Label htmlFor="file-upload" className="cursor-pointer">
                        <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-gray-400 transition-colors">
                          <Upload className="h-12 w-12 mx-auto text-gray-400 mb-3" />
                          <p className="text-sm font-medium text-gray-700 mb-1">
                            Click to upload or drag and drop
                          </p>
                          <p className="text-xs text-gray-500">
                            CSV or XLSX files (MAX. 10MB)
                          </p>
                          {selectedFile && (
                            <div className="mt-3 flex items-center justify-center gap-2 text-sm text-green-600">
                              <FileText className="h-4 w-4" />
                              {selectedFile.name}
                            </div>
                          )}
                        </div>
                        <Input
                          id="file-upload"
                          type="file"
                          accept=".csv,.xlsx,.xls"
                          onChange={handleFileUpload}
                          className="hidden"
                        />
                      </Label>
                    </div>

                    {fileError && (
                      <Alert variant="destructive">
                        <AlertDescription>{fileError}</AlertDescription>
                      </Alert>
                    )}

                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={downloadTemplate}
                        className="flex-1"
                      >
                        <Download className="h-4 w-4 mr-2" />
                        Download Template
                      </Button>
                    </div>

                    <div className="pt-2">
                      <p className="text-sm font-medium text-gray-700 mb-2">
                        File Format (CSV/XLSX):
                      </p>
                      <div className="bg-gray-900 rounded-lg p-3 mb-2">
                        <pre className="text-xs text-gray-100">
                          <code>{`phone,name,message
1234567890,John Doe,Hello John!
9876543210,Jane Smith,Hi Jane!`}</code>
                        </pre>
                      </div>
                      <div className="text-xs text-gray-600 space-y-1">
                        <p>• <strong>phone</strong>: Required. Phone number (no + or spaces)</p>
                        <p>• <strong>name</strong>: Optional. Contact's name</p>
                        <p>• <strong>message</strong>: Optional. Individual message</p>
                        <p className="text-blue-600 mt-2">
                          💡 If message column is empty, the common message below will be used.
                        </p>
                      </div>
                    </div>
                  </TabsContent>
                  
                  {/* Manual Entry Tab */}
                  <TabsContent value="manual" className="space-y-4">
                    <div className="space-y-3">
                      <div>
                        <Label htmlFor="manual-phone">Phone Number *</Label>
                        <div className="flex gap-2 mt-2">
                          <Select value={countryCode} onValueChange={setCountryCode}>
                            <SelectTrigger className="w-[160px]">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {countryCodes.map((c) => (
                                <SelectItem key={c.code} value={c.code}>
                                  <span className="flex items-center gap-2">
                                    <span className="text-lg">{c.flag}</span>
                                    <span>{c.code}</span>
                                  </span>
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <Input
                            id="manual-phone"
                            type="text"
                            placeholder="123-456-7890"
                            value={manualPhone}
                            onChange={handlePhoneInput}
                            className="flex-1"
                          />
                        </div>
                        <p className="text-xs text-gray-500 mt-1">
                          Select country code and enter 10-digit phone number (auto-formatted)
                        </p>
                      </div>
                      
                      <div>
                        <Label htmlFor="manual-name">Name (Optional)</Label>
                        <Input
                          id="manual-name"
                          type="text"
                          placeholder="e.g., John Doe"
                          value={manualName}
                          onChange={(e) => setManualName(e.target.value)}
                          className="mt-2"
                        />
                      </div>
                      
                      <Button
                        onClick={addManualContact}
                        className="w-full"
                        variant="outline"
                      >
                        <Plus className="h-4 w-4 mr-2" />
                        Add Contact
                      </Button>
                    </div>
                    
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                      <p className="text-sm text-blue-800">
                        💡 <strong>Tip:</strong> You can add multiple contacts one by one. They'll appear in the contacts list below.
                      </p>
                    </div>
                  </TabsContent>
                </Tabs>

                {contacts.length > 0 && (
                  <div className="flex gap-2 pt-2 border-t">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={clearAll}
                      className="flex-1"
                    >
                      <X className="h-4 w-4 mr-2" />
                      Clear All ({contacts.length})
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Contacts List */}
            {contacts.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle>Imported Contacts ({contacts.length})</CardTitle>
                  <CardDescription>
                    Review and manage your contact list
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="max-h-96 overflow-y-auto">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-12">#</TableHead>
                          <TableHead>Phone</TableHead>
                          <TableHead>Name</TableHead>
                          {useIndividualMessages && <TableHead>Message</TableHead>}
                          <TableHead className="w-12"></TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {contacts.map((contact, index) => (
                          <TableRow key={index}>
                            <TableCell className="font-mono text-xs">
                              {index + 1}
                            </TableCell>
                            <TableCell className="font-mono">
                              {contact.phone}
                            </TableCell>
                            <TableCell>
                              {contact.name || <span className="text-gray-400">-</span>}
                            </TableCell>
                            {useIndividualMessages && (
                              <TableCell className="max-w-xs truncate">
                                {contact.message || <span className="text-gray-400">-</span>}
                              </TableCell>
                            )}
                            <TableCell>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => removeContact(index)}
                              >
                                <X className="h-4 w-4 text-red-500" />
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>

          {/* Right Column - Message & Send */}
          <div className="space-y-6">
            {/* WhatsApp Session Selection */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span>WhatsApp Session</span>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => window.location.reload()}
                  >
                    Refresh
                  </Button>
                </CardTitle>
                <CardDescription>
                  Select the WhatsApp account to send from
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {/* Debug Info */}
                {process.env.NODE_ENV === 'development' && (
                  <div className="bg-gray-100 rounded p-3 text-xs font-mono space-y-1">
                    <div className="font-bold text-gray-700">Debug Info:</div>
                    <div>Token exists: {token ? '✅ Yes' : '❌ No'}</div>
                    <div>Token length: {token?.length || 0} chars</div>
                    {token && token.length < 50 && (
                      <div className="text-red-600 font-bold">
                        ⚠️ WARNING: Token is too short! ({token.length} chars)
                      </div>
                    )}
                    <div>Loading: {loadingSessions.toString()}</div>
                    <div>Error: {sessionsError ? 'Yes' : 'No'}</div>
                    <div>Sessions Count: {whatsappSessions.length}</div>
                    <div>Selected: {selectedSession || 'none'}</div>
                    {token && token.length < 100 && (
                      <div className="mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded">
                        <div className="text-red-600 font-bold">Full token: {token}</div>
                        <div className="text-xs mt-1 text-gray-600">
                          This token seems invalid. Try logging out and logging back in.
                        </div>
                      </div>
                    )}
                  </div>
                )}
                
                {loadingSessions ? (
                  <div className="text-sm text-gray-500">Loading sessions...</div>
                ) : sessionsError ? (
                  <Alert variant="destructive">
                    <AlertDescription>
                      Error loading WhatsApp sessions. Please make sure you're logged in and have connected a WhatsApp account.
                    </AlertDescription>
                  </Alert>
                ) : whatsappSessions.length === 0 ? (
                  <Alert>
                    <AlertDescription>
                      No WhatsApp sessions found. Please connect a WhatsApp account in the <a href="/api-management" className="underline font-medium">API Management</a> page first.
                    </AlertDescription>
                  </Alert>
                ) : (
                  <>
                    <div className="space-y-2">
                      <Select 
                        value={selectedSession} 
                        onValueChange={(value) => {
                          console.log("Session changed to:", value);
                          setSelectedSession(value);
                        }}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Select WhatsApp session" />
                        </SelectTrigger>
                        <SelectContent>
                          {whatsappSessions.map((session) => (
                            <SelectItem key={session.id} value={session.session_id}>
                              {session.phone_number} {session.name ? `- ${session.name}` : ""} ({session.status})
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      
                      <div className="text-xs text-gray-600">
                        {whatsappSessions.length} session{whatsappSessions.length !== 1 ? 's' : ''} available
                      </div>
                    </div>
                    
                    {selectedSession && (
                      <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                        <div className="flex items-center gap-2">
                          <div className="h-2 w-2 bg-green-500 rounded-full"></div>
                          <div className="text-sm font-medium text-green-900">
                            Active Session: {whatsappSessions.find(s => s.session_id === selectedSession)?.phone_number}
                          </div>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </CardContent>
            </Card>

            {/* Message Composer */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Send className="h-5 w-5" />
                  Compose Message
                </CardTitle>
                <CardDescription>
                  {useIndividualMessages 
                    ? "Individual messages detected in file (This field will be ignored)" 
                    : "Write your message to send to all contacts"}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {useIndividualMessages && (
                  <Alert>
                    <AlertDescription>
                      📝 Your imported file contains individual messages for each contact. 
                      Each person will receive their personalized message from the file.
                    </AlertDescription>
                  </Alert>
                )}
                
                <div>
                  <Label htmlFor="message">
                    {useIndividualMessages ? "Fallback Message (Optional)" : "Message"}
                  </Label>
                  <Textarea
                    id="message"
                    placeholder={useIndividualMessages 
                      ? "This message will be sent to contacts without individual messages..." 
                      : "Type your message here...\n\nFormatting tips:\n*bold* _italic_ ~strikethrough~ `code`\n> Quote\n\nUse {name} for contact's name"}
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    className="min-h-[300px] mt-2 resize-none font-mono"
                    disabled={useIndividualMessages}
                  />
                  <div className="flex justify-between items-center mt-2">
                    <p className="text-xs text-gray-500">
                      {message.length} characters
                    </p>
                    <div className="flex gap-2 text-xs text-gray-500">
                      <span title="Bold">*text*</span>
                      <span title="Italic">_text_</span>
                      <span title="Strikethrough">~text~</span>
                      <span title="Monospace">`text`</span>
                      <span title="Quote">&gt; text</span>
                      <span title="Contact name" className="font-semibold text-blue-600">&#123;name&#125;</span>
                    </div>
                  </div>
                </div>

                <div className="border-t pt-4">
                  {!useIndividualMessages && message && (
                    <div className="bg-gradient-to-br from-green-50 to-blue-50 rounded-lg p-4 mb-4 shadow-sm">
                      <h4 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                        <span className="text-green-600">📱</span>
                        WhatsApp Preview {contacts.length > 0 && `(for ${contacts[0].name || contacts[0].phone})`}
                      </h4>
                      <div className="bg-white rounded-lg p-4 shadow-sm border border-gray-200">
                        <div className="text-[15px] leading-relaxed text-gray-800 font-sans">
                          {renderFormattedText(
                            contacts.length > 0 
                              ? message.replace(/\{name\}/gi, contacts[0].name || contacts[0].phone)
                              : message.replace(/\{name\}/gi, "[Contact Name]")
                          )}
                        </div>
                      </div>
                      {message.includes("{name}") && (
                        <p className="text-xs text-green-700 mt-3 flex items-center gap-1">
                          <span>✨</span>
                          {contacts.length > 0 ? `"{name}" will be replaced with each contact's name` : 'Add contacts to see personalized preview'}
                        </p>
                      )}
                    </div>
                  )}

                  {/* Scheduling Controls */}
                  <div className="border-t pt-4 mt-4">
                    <div className="flex items-center gap-2 mb-4">
                      <Checkbox
                        id="enable-scheduling"
                        checked={enableScheduling}
                        onCheckedChange={(checked) => setEnableScheduling(checked as boolean)}
                      />
                      <Label htmlFor="enable-scheduling" className="text-sm font-medium cursor-pointer">
                        Enable Message Scheduling
                      </Label>
                    </div>
                    
                    {enableScheduling && (
                      <div className="space-y-4 pl-6 border-l-2 border-green-200">
                        <div>
                          <div className="flex justify-between items-center mb-2">
                            <Label className="text-sm text-gray-700">
                              Random Delay Range
                            </Label>
                            <span className="text-sm font-medium text-green-600">
                              {randomDelayMin}s - {randomDelayMax}s
                            </span>
                          </div>
                          <Slider
                            min={5}
                            max={120}
                            step={5}
                            value={[randomDelayMin, randomDelayMax]}
                            onValueChange={(values) => {
                              setRandomDelayMin(values[0]);
                              setRandomDelayMax(values[1]);
                            }}
                            className="w-full"
                          />
                          <p className="text-xs text-gray-500 mt-2">
                            Messages will be sent with cumulative delays. First message immediately, then each subsequent message after a random delay.
                          </p>
                        </div>
                        
                        {contacts.length > 0 && (
                          <div className="bg-green-50 border border-green-200 rounded-md p-3">
                            <p className="text-xs font-medium text-green-800 mb-2">📊 Scheduling Preview (Cumulative):</p>
                            <div className="space-y-1">
                              <p className="text-xs text-green-700">
                                • Message 1: Sent immediately (0s delay)
                              </p>
                              <p className="text-xs text-green-700">
                                • Message 2: Sent after {randomDelayMin}-{randomDelayMax}s from Message 1
                              </p>
                              <p className="text-xs text-green-700">
                                • Message 3: Sent after {randomDelayMin}-{randomDelayMax}s from Message 2
                              </p>
                              <p className="text-xs text-green-700">
                                • And so on... (delays accumulate)
                              </p>
                            </div>
                            <div className="border-t border-green-300 mt-2 pt-2">
                              <p className="text-xs font-medium text-green-800">
                                ⏱️ Total estimated time: ~{Math.ceil(((randomDelayMin + randomDelayMax) / 2 * (contacts.length - 1)) / 60)} minutes for {contacts.length} messages
                              </p>
                              <p className="text-xs text-green-600 mt-1">
                                Each message gets a unique random delay. Backend will schedule them at exact times.
                              </p>
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </div>

                  <Button
                    onClick={handleSendMessages}
                    disabled={(useIndividualMessages ? false : !message.trim()) || contacts.length === 0 || !selectedSession || sending}
                    className="w-full"
                    size="lg"
                  >
                    <Send className="h-5 w-5 mr-2" />
                    {sending ? "Sending..." : enableScheduling ? `Schedule ${contacts.length} Message${contacts.length !== 1 ? "s" : ""}` : `Send to ${contacts.length} Contact${contacts.length !== 1 ? "s" : ""}`}
                  </Button>

                  {contacts.length === 0 && (
                    <p className="text-xs text-center text-gray-500 mt-2">
                      Import contacts to enable sending
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Message History Section */}
        <Card className="mt-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <RotateCcw className="h-5 w-5" />
                Message History
              </CardTitle>
              <div className="flex items-center gap-2">
                {messageStatsData && (
                  <div className="flex items-center gap-3 mr-4 text-sm">
                    <span className="flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
                      Pending: {messageStatsData.pending || 0}
                    </span>
                    <span className="flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-green-500"></span>
                      Sent: {messageStatsData.sent || 0}
                    </span>
                    <span className="flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-red-500"></span>
                      Failed: {messageStatsData.failed || 0}
                    </span>
                  </div>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setShowMessageHistory(!showMessageHistory);
                    if (!showMessageHistory) {
                      refetchLogs();
                    }
                  }}
                >
                  {showMessageHistory ? (
                    <>
                      <ChevronUp className="h-4 w-4 mr-2" />
                      Hide History
                    </>
                  ) : (
                    <>
                      <ChevronDown className="h-4 w-4 mr-2" />
                      Show History
                    </>
                  )}
                </Button>
              </div>
            </div>
          </CardHeader>
          
          {showMessageHistory && (
            <CardContent>
              {messageLogsData?.logs && messageLogsData.logs.length > 0 ? (
                <div className="space-y-3">
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead className="border-b">
                        <tr className="text-left">
                          <th className="pb-3 font-medium text-gray-600">Time / Scheduled</th>
                          <th className="pb-3 font-medium text-gray-600">Recipient</th>
                          <th className="pb-3 font-medium text-gray-600">Status</th>
                          <th className="pb-3 font-medium text-gray-600">Message</th>
                          <th className="pb-3 font-medium text-gray-600">Type</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y">
                        {messageLogsData.logs.map((log: MessageLog) => (
                          <tr key={log.id} className="hover:bg-gray-50">
                            <td className="py-3">
                              {log.status === 'pending' && log.scheduled_at ? (
                                <div className="text-sm">
                                  <div className="text-gray-600 font-medium">
                                    ⏰ {new Date(log.scheduled_at).toLocaleString()}
                                  </div>
                                  <div className="text-xs text-gray-400">
                                    Scheduled
                                  </div>
                                </div>
                              ) : (
                                <div className="text-sm text-gray-600">
                                  {new Date(log.sent_at || log.created_at).toLocaleString()}
                                </div>
                              )}
                            </td>
                            <td className="py-3 font-medium">
                              {log.recipient_phone?.replace('@s.whatsapp.net', '') || 'N/A'}
                            </td>
                            <td className="py-3">
                              {log.status === 'pending' && (
                                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                                  <Clock className="h-3 w-3" />
                                  Pending
                                </span>
                              )}
                              {log.status === 'sent' && (
                                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                  <CheckCircle2 className="h-3 w-3" />
                                  Sent
                                </span>
                              )}
                              {log.status === 'failed' && (
                                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                                  <XCircle className="h-3 w-3" />
                                  Failed
                                </span>
                              )}
                              {log.status === 'sending' && (
                                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                  <RefreshCw className="h-3 w-3 animate-spin" />
                                  Sending
                                </span>
                              )}
                            </td>
                            <td className="py-3 max-w-md truncate text-gray-600">
                              {log.message || 'N/A'}
                            </td>
                            <td className="py-3 text-gray-500 text-xs">
                              {log.message_type || 'text'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  
                  <div className="flex items-center justify-between pt-4 border-t">
                    <p className="text-sm text-gray-500">
                      Showing {messageLogsData.logs.length} of {messageLogsData.total || messageLogsData.logs.length} messages
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => refetchLogs()}
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Refresh
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <RotateCcw className="h-12 w-12 mx-auto mb-3 opacity-30" />
                  <p className="font-medium">No messages sent yet</p>
                  <p className="text-sm mt-1">Your message history will appear here</p>
                </div>
              )}
            </CardContent>
          )}
        </Card>
              </div>
            </main>
          </div>
        </div>

      {/* Result Alert Dialog */}
      <AlertDialog open={showResultDialog} onOpenChange={setShowResultDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              {resultType === "success" ? (
                <>
                  <CheckCircle2 className="h-5 w-5 text-green-600" />
                  <span className="text-green-600">Success</span>
                </>
              ) : (
                <>
                  <XCircle className="h-5 w-5 text-red-600" />
                  <span className="text-red-600">Error</span>
                </>
              )}
            </AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-3">
                <p className="text-base text-gray-700">{resultMessage}</p>
                
                {(successCount > 0 || failCount > 0) && (
                  <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-gray-600">Total Messages:</span>
                      <span className="text-sm font-semibold text-gray-900">{successCount + failCount}</span>
                    </div>
                    {successCount > 0 && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-green-700 flex items-center gap-1">
                          <CheckCircle2 className="h-4 w-4" />
                          Sent Successfully:
                        </span>
                        <span className="text-sm font-semibold text-green-700">{successCount}</span>
                      </div>
                    )}
                    {failCount > 0 && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-red-700 flex items-center gap-1">
                          <XCircle className="h-4 w-4" />
                          Failed:
                        </span>
                        <span className="text-sm font-semibold text-red-700">{failCount}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction 
              className={resultType === "success" ? "bg-green-600 hover:bg-green-700" : "bg-red-600 hover:bg-red-700"}
            >
              OK
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </SidebarProvider>
  );
}
