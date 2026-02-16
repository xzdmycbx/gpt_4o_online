import axios, { AxiosInstance, AxiosError } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';
const PUBLIC_PATHS = new Set(['/login', '/register', '/forgot-password', '/reset-password', '/oauth2/callback']);
let loginRedirectInProgress = false;

const shouldRedirectToLogin = () => {
  if (typeof window === 'undefined') {
    return false;
  }
  const normalizedPath = window.location.pathname.replace(/\/+$/, '') || '/';
  return !PUBLIC_PATHS.has(normalizedPath);
};

const redirectToLoginOnce = () => {
  if (!shouldRedirectToLogin() || loginRedirectInProgress) {
    return;
  }

  loginRedirectInProgress = true;
  window.location.assign('/login');
};

class APIClient {
  private client: AxiosInstance;
  private token: string | null = null;
  private csrfToken: string | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true, // Important: send cookies with requests
    });

    // Load token from localStorage (for backwards compatibility)
    this.token = localStorage.getItem('auth_token');

    // Request interceptor to add auth token and CSRF token
    this.client.interceptors.request.use(
      (config) => {
        // Try Authorization header first (for regular login)
        if (this.token) {
          config.headers.Authorization = `Bearer ${this.token}`;
        }

        // Add CSRF token for state-changing requests
        if (['post', 'put', 'delete', 'patch'].includes(config.method?.toLowerCase() || '')) {
          const csrfToken = this.getCSRFToken();
          if (csrfToken) {
            config.headers['X-CSRF-Token'] = csrfToken;
          }
        }

        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for error handling and CSRF token extraction
    this.client.interceptors.response.use(
      (response) => {
        // Extract CSRF token from response headers if present
        const csrfToken = response.headers['x-csrf-token'];
        if (csrfToken) {
          this.csrfToken = csrfToken;
          sessionStorage.setItem('csrf_token', csrfToken);
        }
        return response;
      },
      async (error: AxiosError) => {
        if (error.response?.status === 401 && this.token) {
          // If we have a token in localStorage and got 401, it might be expired
          // Clear it and retry with cookie authentication
          this.clearToken();

          // Retry the request once without the Authorization header (will use cookie)
          const originalRequest = error.config;
          if (originalRequest && !originalRequest.headers?.['X-Retry-Count']) {
            originalRequest.headers = originalRequest.headers || {};
            originalRequest.headers['X-Retry-Count'] = '1';
            delete originalRequest.headers.Authorization;

            try {
              return await this.client.request(originalRequest);
            } catch (retryError) {
              // If retry also fails, redirect to login
              redirectToLoginOnce();
              return Promise.reject(retryError);
            }
          }
        }

        if (error.response?.status === 401) {
          // No token or retry failed, redirect to login
          redirectToLoginOnce();
        }

        return Promise.reject(error);
      }
    );

    // Initialize CSRF token from session storage
    this.csrfToken = sessionStorage.getItem('csrf_token');
  }

  getCSRFToken(): string | null {
    // First try to get from memory
    if (this.csrfToken) {
      return this.csrfToken;
    }

    // Then try session storage
    const stored = sessionStorage.getItem('csrf_token');
    if (stored) {
      this.csrfToken = stored;
      return stored;
    }

    // Finally try to extract from cookie (as fallback)
    const cookieValue = document.cookie
      .split('; ')
      .find(row => row.startsWith('csrf_token='))
      ?.split('=')[1];

    if (cookieValue) {
      this.csrfToken = cookieValue;
      sessionStorage.setItem('csrf_token', cookieValue);
      return cookieValue;
    }

    return null;
  }

  async initCSRFToken() {
    // Fetch CSRF token from server on app initialization
    try {
      const response = await this.client.get('/csrf-token');
      const token = response.data.csrf_token;
      if (token) {
        this.csrfToken = token;
        sessionStorage.setItem('csrf_token', token);
      }
    } catch (error) {
      console.warn('Failed to fetch CSRF token:', error);
    }
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem('auth_token', token);
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem('auth_token');
  }

  getToken() {
    return this.token;
  }

  // Generic HTTP methods (for components that use apiClient.get/post/put/delete directly)
  async get(url: string, config?: any) {
    const response = await this.client.get(url, config);
    return response;
  }

  async post(url: string, data?: any, config?: any) {
    const response = await this.client.post(url, data, config);
    return response;
  }

  async put(url: string, data?: any, config?: any) {
    const response = await this.client.put(url, data, config);
    return response;
  }

  async delete(url: string, config?: any) {
    const response = await this.client.delete(url, config);
    return response;
  }

  // Auth endpoints
  auth = {
    login: async (username: string, password: string) => {
      const response = await this.client.post('/auth/login', { username, password });
      return response.data;
    },

    register: async (username: string, password: string, email?: string) => {
      const response = await this.client.post('/auth/register', { username, password, email });
      return response.data;
    },

    getTwitterAuthUrl: async () => {
      const response = await this.client.get('/auth/oauth2/twitter');
      return response.data;
    },

    handleOAuthCallback: async (code: string, state: string) => {
      const response = await this.client.get(`/auth/oauth2/callback?code=${code}&state=${state}`);
      return response.data;
    },

    forgotPassword: async (email: string) => {
      const response = await this.client.post('/auth/forgot-password', { email });
      return response.data;
    },

    resetPassword: async (token: string, newPassword: string) => {
      const response = await this.client.post('/auth/reset-password', { token, new_password: newPassword });
      return response.data;
    },

    getCurrentUser: async () => {
      const response = await this.client.get('/me');
      return response.data;
    },

    logout: async () => {
      const response = await this.client.post('/logout');
      return response.data;
    },
  };

  // Conversation endpoints
  conversations = {
    list: async (limit = 20, offset = 0) => {
      const response = await this.client.get('/conversations', { params: { limit, offset } });
      return response.data;
    },

    create: async (title?: string, modelId?: string) => {
      const response = await this.client.post('/conversations', { title, model_id: modelId });
      return response.data;
    },

    get: async (id: string) => {
      const response = await this.client.get(`/conversations/${id}`);
      return response.data;
    },

    update: async (id: string, title: string) => {
      const response = await this.client.put(`/conversations/${id}`, { title });
      return response.data;
    },

    delete: async (id: string) => {
      const response = await this.client.delete(`/conversations/${id}`);
      return response.data;
    },

    getMessages: async (id: string, limit = 50, offset = 0) => {
      const response = await this.client.get(`/conversations/${id}/messages`, { params: { limit, offset } });
      return response.data;
    },

    sendMessage: async (id: string, content: string, modelId?: string) => {
      const response = await this.client.post(`/conversations/${id}/messages`, { content, model_id: modelId });
      return response.data;
    },
  };

  // Memory endpoints
  memories = {
    list: async (limit = 50, offset = 0) => {
      const response = await this.client.get('/memories', { params: { limit, offset } });
      return response.data;
    },

    create: async (content: string, category: string, importance: number) => {
      const response = await this.client.post('/memories', { content, category, importance });
      return response.data;
    },

    update: async (id: string, content?: string, category?: string, importance?: number) => {
      const response = await this.client.put(`/memories/${id}`, { content, category, importance });
      return response.data;
    },

    delete: async (id: string) => {
      const response = await this.client.delete(`/memories/${id}`);
      return response.data;
    },
  };

  // Settings endpoints
  settings = {
    get: async () => {
      const response = await this.client.get('/user/settings');
      return response.data;
    },

    update: async (settings: any) => {
      const response = await this.client.put('/user/settings', settings);
      return response.data;
    },

    sync: async (localSettings: any) => {
      const response = await this.client.post('/user/settings/sync', localSettings);
      return response.data;
    },
  };

  // Admin endpoints
  admin = {
    listUsers: async (limit = 20, offset = 0) => {
      const response = await this.client.get('/admin/users', { params: { limit, offset } });
      return response.data;
    },

    banUser: async (id: string, reason: string) => {
      const response = await this.client.put(`/admin/users/${id}/ban`, { reason });
      return response.data;
    },

    unbanUser: async (id: string) => {
      const response = await this.client.put(`/admin/users/${id}/unban`);
      return response.data;
    },

    listModels: async (activeOnly = false) => {
      const response = await this.client.get('/admin/models', { params: { active_only: activeOnly } });
      return response.data;
    },

    createModel: async (model: any) => {
      const response = await this.client.post('/admin/models', model);
      return response.data;
    },

    updateModel: async (id: string, updates: any) => {
      const response = await this.client.put(`/admin/models/${id}`, updates);
      return response.data;
    },

    deleteModel: async (id: string) => {
      const response = await this.client.delete(`/admin/models/${id}`);
      return response.data;
    },

    getTokenLeaderboard: async (limit = 10) => {
      const response = await this.client.get('/admin/statistics/tokens', { params: { limit } });
      return response.data;
    },

    getSystemOverview: async () => {
      const response = await this.client.get('/admin/statistics/overview');
      return response.data;
    },
  };
}

export const api = new APIClient();
export default api;
