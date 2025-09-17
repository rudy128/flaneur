export interface User {
  user_id: number;
  username: string;
  email: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  session_id: string;
  user: User;
  status_code: number;
}

export interface WSMessage {
  type: 'register' | 'login' | 'logout' | 'auth' | 'auth_success' | 'auth_error' | 'error';
  data?: any;
  session_id?: string;
  status_code?: number;
  message?: string;
}

export interface WSAuthResponse {
  type: string;
  data?: {
    session_id?: string;
    user?: User;
    requestId?: string;
  };
  status_code: number;
  message: string;
}