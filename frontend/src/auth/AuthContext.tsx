import React, { useState, useEffect, type ReactNode } from 'react';
import type { User} from '@/types/auth';
import AuthService from '@/auth/AuthService';
import WebSocketService from '@/websocket/WebSocketService';
import { AuthContext } from '@/auth/context';

const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const sessionUser = AuthService.getUser();
    if (sessionUser && AuthService.isAuthenticated()) {
      setUser(sessionUser);
      setIsAuthenticated(true);
    }
    setLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    const response = await WebSocketService.login({ email, password });
    AuthService.setSession(response.session_id, response.user);
    setUser(response.user);
    setIsAuthenticated(true);
  };

  const register = async (name: string, email: string, password: string) => {
    await WebSocketService.register({ name, email, password });
  };

  const logout = async () => {
    try {
      await WebSocketService.logout();
    } catch (error) {
      console.error('Logout request failed:', error);
    }
    AuthService.clearSession();
    setUser(null);
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ user, login, register, logout, isAuthenticated, loading }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;