import type { User } from '@/types/auth';

class AuthService {

  setSession(sessionId: string, user: User): void {
    sessionStorage.setItem('session_id', sessionId);
    sessionStorage.setItem('user', JSON.stringify(user));
  }

  getSessionId(): string | null {
    return sessionStorage.getItem('session_id');
  }

  getUser(): User | null {
    const userStr = sessionStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }

  clearSession(): void {
    sessionStorage.removeItem('session_id');
    sessionStorage.removeItem('user');
  }

  isAuthenticated(): boolean {
    return !!this.getSessionId();
  }
}

export default new AuthService();