import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import apiClient from '../api/client';

interface UserSettings {
  theme: 'dark' | 'light' | 'auto';
  fontSize: 'small' | 'medium' | 'large';
  language: string;
  notifications: {
    enabled: boolean;
    sound: boolean;
  };
  chatPreferences: {
    defaultModelId?: string;
    streamResponse: boolean;
    showTokenCount: boolean;
  };
  updatedAt?: string;
  deviceId?: string;
}

const DEFAULT_SETTINGS: UserSettings = {
  theme: 'dark',
  fontSize: 'medium',
  language: 'en',
  notifications: {
    enabled: true,
    sound: true,
  },
  chatPreferences: {
    streamResponse: true,
    showTokenCount: false,
  },
};

const STORAGE_KEY = 'ai_chat_settings';
const DEVICE_ID_KEY = 'ai_chat_device_id';

// Generate or retrieve device ID
const getDeviceId = (): string => {
  let deviceId = localStorage.getItem(DEVICE_ID_KEY);
  if (!deviceId) {
    deviceId = `device_${Date.now()}_${Math.random().toString(36).substring(2, 15)}`;
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
  }
  return deviceId;
};

export const useSettings = () => {
  const { isAuthenticated } = useAuth();
  const [settings, setSettings] = useState<UserSettings>(() => {
    // Load from localStorage
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) };
      } catch (error) {
        console.error('Failed to parse settings:', error);
      }
    }
    return DEFAULT_SETTINGS;
  });

  const [isSyncing, setIsSyncing] = useState(false);
  const [lastSyncTime, setLastSyncTime] = useState<Date | null>(null);

  // Sync with server
  const syncSettings = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      console.log('Not authenticated, skipping server sync');
      return;
    }

    setIsSyncing(true);

    try {
      const localSettings = {
        ...settings,
        updatedAt: new Date().toISOString(),
        deviceId: getDeviceId(),
      };

      const response = await apiClient.settings.sync(localSettings);

      if (response.action === 'pulled') {
        // Server had newer settings, update local state
        setSettings((prev: UserSettings) => ({
          ...prev,
          ...response.settings,
        }));
        console.log('Settings synced: pulled from server');
      } else {
        // Local settings were pushed to server
        console.log('Settings synced: pushed to server');
      }

      setLastSyncTime(new Date());
    } catch (error) {
      console.error('Failed to sync settings:', error);
      throw error;
    } finally {
      setIsSyncing(false);
    }
  }, [isAuthenticated, settings]);

  const loadServerSettings = useCallback(async () => {
    try {
      const response = await apiClient.settings.get();
      const serverSettings = response.settings;

      if (serverSettings) {
        // Compare updatedAt to decide whether to pull from server
        const localUpdatedAt = settings.updatedAt ? new Date(settings.updatedAt) : new Date(0);
        const serverUpdatedAt = serverSettings.updatedAt
          ? new Date(serverSettings.updatedAt)
          : new Date(0);

        if (serverUpdatedAt > localUpdatedAt) {
          // Server has newer settings, pull from server
          setSettings((prev: UserSettings) => ({
            ...prev,
            ...serverSettings,
          }));
          console.log('Settings pulled from server');
        }
      }
    } catch (error) {
      console.error('Failed to load server settings:', error);
    }
  }, [settings.updatedAt]);

  const updateSettings = useCallback((updates: Partial<UserSettings>) => {
    setSettings((prev: UserSettings) => ({
      ...prev,
      ...updates,
    }));
  }, []);

  const resetSettings = useCallback(() => {
    setSettings(DEFAULT_SETTINGS);
    localStorage.removeItem(STORAGE_KEY);
  }, []);

  // Save to localStorage whenever settings change
  useEffect(() => {
    const settingsToSave = {
      ...settings,
      updatedAt: new Date().toISOString(),
      deviceId: getDeviceId(),
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settingsToSave));

    // Auto-sync to server after 500ms debounce (if authenticated)
    if (isAuthenticated) {
      const timer = setTimeout(() => {
        syncSettings().catch((error: unknown) => {
          console.error('Auto-sync failed:', error);
        });
      }, 500);

      return () => clearTimeout(timer);
    }
  }, [settings, isAuthenticated, syncSettings]);

  // Load settings from server on mount (if authenticated)
  useEffect(() => {
    if (isAuthenticated) {
      loadServerSettings().catch((error: unknown) => {
        console.error('Failed to load settings on mount:', error);
      });
    }
  }, [isAuthenticated, loadServerSettings]);

  return {
    settings,
    updateSettings,
    resetSettings,
    syncSettings,
    isSyncing,
    lastSyncTime,
  };
};
