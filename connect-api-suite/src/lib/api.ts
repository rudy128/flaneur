import type { 
  LoginRequest, 
  LoginResponse, 
  SignupRequest, 
  MessageResponse, 
  ErrorResponse,
  ApiCallLog,
  ApiStats
} from "./schemas";

const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

/**
 * Custom error class for API errors
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    public data: ErrorResponse | unknown
  ) {
    super(`API Error: ${status} ${statusText}`);
    this.name = "ApiError";
  }
}

/**
 * Generic fetch wrapper with error handling
 */
async function fetchApi<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    let errorData: ErrorResponse | unknown;
    try {
      errorData = await response.json();
    } catch {
      errorData = { error: response.statusText };
    }
    throw new ApiError(response.status, response.statusText, errorData);
  }

  return response.json();
}

/**
 * Auth API
 */
export const authApi = {
  /**
   * Sign up a new user
   */
  signup: async (data: SignupRequest): Promise<MessageResponse> => {
    return fetchApi<MessageResponse>("/auth/signup", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  /**
   * Login user
   */
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    return fetchApi<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  /**
   * Get user profile
   */
  getProfile: async (token: string): Promise<{
    email: string;
    name: string;
    created_at: string;
    accounts_connected: number;
  }> => {
    return fetchApi("/profile", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  },

  /**
   * Change user password
   */
  changePassword: async (
    token: string,
    data: { current_password: string; new_password: string }
  ): Promise<MessageResponse> => {
    return fetchApi("/change-password", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });
  },
};

/**
 * Dashboard API
 */
export const dashboardApi = {
  /**
   * Get dashboard data
   */
  getDashboard: async (token: string): Promise<MessageResponse> => {
    return fetchApi<MessageResponse>("/dashboard", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  },
};

/**
 * Twitter API
 */
export const twitterApi = {
  /**
   * Get all Twitter accounts for the authenticated user
   */
  getAccounts: async (
    jwtToken: string
  ): Promise<{ accounts: Array<{ id: string; username: string; token: string }>; count: number }> => {
    return fetchApi("/twitter/", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Add Twitter account
   */
  addAccount: async (
    jwtToken: string,
    data: { username: string; password: string }
  ): Promise<{ id: string; username: string; token: string; user_id: string; message: string }> => {
    return fetchApi("/twitter/account", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
      body: JSON.stringify(data),
    });
  },

  /**
   * Get tweet data
   */
  getTweet: async (twitterToken: string, url: string): Promise<unknown> => {
    return fetchApi("/twitter/post", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${twitterToken}`,
      },
      body: JSON.stringify({ url }),
    });
  },

  /**
   * Get tweet likes
   */
  getTweetLikes: async (twitterToken: string, url: string): Promise<unknown> => {
    return fetchApi("/twitter/post/likes", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${twitterToken}`,
      },
      body: JSON.stringify({ url }),
    });
  },

  /**
   * Get tweet quotes
   */
  getTweetQuotes: async (twitterToken: string, url: string): Promise<unknown> => {
    return fetchApi("/twitter/post/quotes", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${twitterToken}`,
      },
      body: JSON.stringify({ url }),
    });
  },

  /**
   * Get tweet comments
   */
  getTweetComments: async (twitterToken: string, url: string): Promise<unknown> => {
    return fetchApi("/twitter/post/comments", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${twitterToken}`,
      },
      body: JSON.stringify({ url }),
    });
  },

  /**
   * Get tweet reposts
   */
  getTweetReposts: async (twitterToken: string, url: string): Promise<unknown> => {
    return fetchApi("/twitter/post/reposts", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${twitterToken}`,
      },
      body: JSON.stringify({ url }),
    });
  },

  /**
   * Regenerate Twitter token
   */
  regenerateToken: async (
    jwtToken: string,
    username: string
  ): Promise<{ token: string; message: string }> => {
    return fetchApi("/twitter/regenerate-token", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
      body: JSON.stringify({ username }),
    });
  },
};

/**
 * Logs API
 */
export const logsApi = {
  /**
   * Get API call logs
   */
  getLogs: async (jwtToken: string, limit: number = 50): Promise<ApiCallLog[]> => {
    return fetchApi<ApiCallLog[]>(`/logs?limit=${limit}`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Get API call statistics
   */
  getStats: async (jwtToken: string): Promise<ApiStats> => {
    return fetchApi<ApiStats>("/logs/stats", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },
};

/**
 * WhatsApp API
 */
export const whatsappApi = {
  /**
   * Generate QR code for WhatsApp login
   */
  generateQR: async (jwtToken: string): Promise<{
    session_id: string;
    qr_code: string;
    status: string;
    message: string;
  }> => {
    return fetchApi("/whatsapp/generate-qr", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Check WhatsApp session status
   */
  checkSessionStatus: async (jwtToken: string, sessionId: string): Promise<{
    session_id: string;
    status: string;
    phone_number?: string;
    name?: string;
    message: string;
  }> => {
    return fetchApi(`/whatsapp/session-status/${sessionId}`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Get connected WhatsApp accounts
   */
  getAccounts: async (jwtToken: string): Promise<{
    accounts: Array<{
      id: string;
      phone_number: string;
      name?: string;
      session_id: string;
      status: string;
      created_at: string;
    }>;
    count: number;
  }> => {
    return fetchApi("/whatsapp/", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Send WhatsApp message
   */
  sendMessage: async (
    jwtToken: string,
    sessionId: string,
    phone: string,
    message: string,
    reply: boolean = false
  ): Promise<{
    success: boolean;
    message: string;
    error?: string;
  }> => {
    return fetchApi("/whatsapp/send-message", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
      body: JSON.stringify({ session_id: sessionId, phone, message, reply }),
    });
  },

  /**
   * Delete WhatsApp account
   */
  deleteAccount: async (
    jwtToken: string,
    accountId: string
  ): Promise<{
    message: string;
    id: string;
  }> => {
    return fetchApi(`/whatsapp/account/${accountId}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Get message logs (history)
   */
  getMessageLogs: async (
    jwtToken: string,
    params?: {
      status?: string;
      batch_id?: string;
      limit?: number;
      offset?: number;
    }
  ): Promise<{
    logs: Array<{
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
      batch_id: string;
      sequence_number: number;
      delay_seconds: number;
      created_at: string;
      updated_at: string;
    }>;
    total: number;
    limit: number;
    offset: number;
  }> => {
    const queryParams = new URLSearchParams();
    if (params?.status) queryParams.append("status", params.status);
    if (params?.batch_id) queryParams.append("batch_id", params.batch_id);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    const url = `/whatsapp/message-logs${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    return fetchApi(url, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Get message log statistics
   */
  getMessageLogStats: async (
    jwtToken: string
  ): Promise<{
    total: number;
    pending: number;
    sent: number;
    failed: number;
    paused?: number;
  }> => {
    return fetchApi("/whatsapp/message-logs/stats", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    });
  },

  /**
   * Resume paused messages and optionally change session
   */
  resumePausedMessages: async (
    jwtToken: string,
    data: {
      old_session_id: string;
      new_session_id?: string;
      batch_id?: string;
    }
  ): Promise<{
    success: boolean;
    message: string;
    count: number;
  }> => {
    return fetchApi("/whatsapp/resume-paused", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${jwtToken}`,
      },
      body: JSON.stringify(data),
    });
  },
};

export { API_BASE_URL };
