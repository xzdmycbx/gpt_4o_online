import { useState, useEffect } from 'react';

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

export const useSettings = () => {
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

  // Save to localStorage whenever settings change
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  }, [settings]);

  const updateSettings = (updates: Partial<UserSettings>) => {
    setSettings((prev) => ({
      ...prev,
      ...updates,
    }));
  };

  const resetSettings = () => {
    setSettings(DEFAULT_SETTINGS);
    localStorage.removeItem(STORAGE_KEY);
  };

  // Sync with server (to be implemented with API)
  const syncSettings = async () => {
    // TODO: Implement server sync
    console.log('Syncing settings with server...');
  };

  return {
    settings,
    updateSettings,
    resetSettings,
    syncSettings,
  };
};
