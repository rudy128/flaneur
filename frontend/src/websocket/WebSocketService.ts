import type { WSMessage, WSAuthResponse, LoginRequest, RegisterRequest, User } from '@/types/auth';
import { AuthError, NetworkError } from '@/errors/AuthError';
import AuthService from '@/auth/AuthService';

class WebSocketService {
  private ws: WebSocket | null = null;
  private url = 'ws://localhost:8080/ws';
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private pendingRequests: Map<string, { resolve: (value?: any) => void; reject: (reason?: any) => void }> = new Map();

  connect(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => resolve();
      this.ws.onmessage = (event) => {
        const message: WSAuthResponse = JSON.parse(event.data);
        this.handleMessage(message);
      };
      this.ws.onclose = () => console.log('WebSocket connection closed');
      this.ws.onerror = () => reject(new NetworkError());
    });
  }

  async register(data: RegisterRequest): Promise<void> {
    await this.connect();
    return new Promise<void>((resolve, reject) => {
      const requestId = 'register_' + Date.now();
      this.pendingRequests.set(requestId, { resolve, reject });
      this.send({ type: 'register', data: { ...data, requestId } });
    });
  }

  async login(data: LoginRequest): Promise<{ session_id: string; user: User }> {
    await this.connect();
    return new Promise<{ session_id: string; user: User }>((resolve, reject) => {
      const requestId = 'login_' + Date.now();
      this.pendingRequests.set(requestId, { resolve, reject });
      this.send({ type: 'login', data: { ...data, requestId } });
    });
  }

  async logout(): Promise<void> {
    const sessionId = AuthService.getSessionId();
    if (!sessionId) return;
    
    return new Promise<void>((resolve, reject) => {
      const requestId = 'logout_' + Date.now();
      this.pendingRequests.set(requestId, { resolve, reject });
      this.send({ type: 'logout', session_id: sessionId, data: { requestId } });
    });
  }

  private send(message: WSMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  private handleMessage(message: WSAuthResponse): void {
    const requestId = message.data?.requestId;
    
    if (requestId) {
      const pending = this.pendingRequests.get(requestId);
      if (pending) {
        this.pendingRequests.delete(requestId);
        
        if (message.status_code >= 200 && message.status_code < 300) {
          pending.resolve(message.data);
        } else {
          pending.reject(new AuthError(message.message, message.status_code));
        }
        return;
      }
    }

    const handler = this.messageHandlers.get(message.type);
    if (handler) {
      handler(message);
    }
  }

  addMessageHandler(type: string, handler: (data: any) => void): void {
    this.messageHandlers.set(type, handler);
  }

  removeMessageHandler(type: string): void {
    this.messageHandlers.delete(type);
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
    this.pendingRequests.clear();
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export default new WebSocketService();