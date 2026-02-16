import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import api from '../api/client';

interface User {
  id: string;
  username: string;
  email?: string;
  display_name?: string;
  avatar_url?: string;
  role: string;
  oauth2_provider?: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string, email?: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
  isLoading: boolean;
  loading: boolean; // Alias for compatibility
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(api.getToken());
  const [isLoading, setIsLoading] = useState(true);

  const loadCurrentUser = useCallback(async () => {
    try {
      const response = await api.auth.getCurrentUser();
      setUser(response.user);
      // If user loaded successfully but no token in localStorage, it means cookie auth
      if (!token) {
        setToken('cookie'); // Placeholder to indicate authenticated state
      }
    } catch (error) {
      console.error('Failed to load current user:', error);
      // Clear any stale token from localStorage
      if (token && token !== 'cookie') {
        api.clearToken();
        setToken(null);
      }
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, [token]);

  useEffect(() => {
    // Always try to load user on mount - supports both localStorage token and HttpOnly cookie
    loadCurrentUser();
  }, [loadCurrentUser]);

  const login = async (username: string, password: string) => {
    const response = await api.auth.login(username, password);
    setToken(response.token);
    setUser(response.user);
    api.setToken(response.token);
  };

  const register = async (username: string, password: string, email?: string) => {
    const response = await api.auth.register(username, password, email);
    setToken(response.token);
    setUser(response.user);
    api.setToken(response.token);
  };

  const logout = async () => {
    try {
      // Call backend to clear HttpOnly cookie
      await api.auth.logout();
    } catch (error) {
      console.error('Logout error:', error);
      // Continue with local cleanup even if backend call fails
    }

    // Clear local state
    setUser(null);
    setToken(null);
    api.clearToken();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        setUser,
        setToken: (newToken: string | null) => {
          setToken(newToken);
          if (newToken) {
            api.setToken(newToken);
          } else {
            api.clearToken();
          }
        },
        login,
        register,
        logout,
        isAuthenticated: !!user,
        isLoading,
        loading: isLoading, // Alias for compatibility
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
