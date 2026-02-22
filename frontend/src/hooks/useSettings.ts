import { useState, useEffect, useCallback, useRef } from 'react';
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
  // Use a ref to track updatedAt so loadServerSettings doesn't re-create on every settings change
  const updatedAtRef = useRef<string | undefined>(settings.updatedAt);

  // Keep ref in sync with latest settings.updatedAt without causing re-renders
  useEffect(() => {
    updatedAtRef.current = settings.updatedAt;
  }, [settings.updatedAt]);

  // settingsRef lets syncSettings always read the latest settings without being in its deps
  const settingsRef = useRef<UserSettings>(settings);
  useEffect(() => {
    settingsRef.current = settings;
  }, [settings]);

  // Sync with server
  const syncSettings = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }

    setIsSyncing(true);

    try {
      const now = new Date().toISOString();
      const deviceId = getDeviceId();
      const localSettings = {
        ...settingsRef.current,
        // Send both camelCase (local) and snake_case (Go JSON binding)
        updatedAt: now,
        updated_at: now,
        deviceId,
        device_id: deviceId,
      };

      const response = await apiClient.settings.sync(localSettings);

      if (response.action === 'pulled' && response.settings) {
        const s = response.settings;
        // Map snake_case backend fields to camelCase frontend schema
        setSettings((prev: UserSettings) => ({
          ...prev,
          theme: s.theme ?? prev.theme,
          fontSize: s.font_size ?? s.fontSize ?? prev.fontSize,
          language: s.language ?? prev.language,
        }));
        console.log('Settings synced: pulled from server');
      } else {
        console.log('Settings synced: pushed to server');
      }

      setLastSyncTime(new Date());
    } catch (error) {
      console.error('Failed to sync settings:', error);
      throw error;
    } finally {
      setIsSyncing(false);
    }
  }, [isAuthenticated]);
  // NOTE: syncSettings no longer depends on `settings` — it reads via settingsRef

  // loadServerSettings uses a ref for updatedAt so it doesn't change on every save
  const loadServerSettings = useCallback(async () => {
    try {
      // GET /user/settings returns the settings object directly (not wrapped)
      const serverSettings = await apiClient.settings.get();

      if (serverSettings) {
        const localUpdatedAt = updatedAtRef.current ? new Date(updatedAtRef.current) : new Date(0);
        // Backend uses snake_case: updated_at
        const serverTimestamp = serverSettings.updated_at || serverSettings.updatedAt;
        const serverUpdatedAt = serverTimestamp ? new Date(serverTimestamp) : new Date(0);

        if (serverUpdatedAt > localUpdatedAt) {
          // Map snake_case server fields back to camelCase frontend schema
          setSettings((prev: UserSettings) => ({
            ...prev,
            theme: serverSettings.theme ?? prev.theme,
            fontSize: serverSettings.font_size ?? serverSettings.fontSize ?? prev.fontSize,
            language: serverSettings.language ?? prev.language,
          }));
          console.log('Settings pulled from server');
        }
      }
    } catch (error) {
      console.error('Failed to load server settings:', error);
    }
  }, []); // stable — reads updatedAt via ref, not as closure

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

  // Save to localStorage whenever settings change, then debounce-sync to server
  useEffect(() => {
    const settingsToSave = {
      ...settings,
      updatedAt: new Date().toISOString(),
      deviceId: getDeviceId(),
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settingsToSave));

    if (isAuthenticated) {
      const timer = setTimeout(() => {
        syncSettings().catch((error: unknown) => {
          console.error('Auto-sync failed:', error);
        });
      }, 500);

      return () => clearTimeout(timer);
    }
  }, [settings, isAuthenticated, syncSettings]);

  // Load settings from server ONCE on mount (when authenticated)
  // loadServerSettings is now stable (no changing deps), so this only runs
  // when isAuthenticated changes, not on every settings change.
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
