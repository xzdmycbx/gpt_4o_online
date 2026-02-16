import React, { createContext, useContext, ReactNode } from 'react';
import { useSettings as useSettingsHook } from '../hooks/useSettings';

interface SettingsContextType {
  settings: ReturnType<typeof useSettingsHook>['settings'];
  updateSettings: ReturnType<typeof useSettingsHook>['updateSettings'];
  resetSettings: ReturnType<typeof useSettingsHook>['resetSettings'];
  syncSettings: ReturnType<typeof useSettingsHook>['syncSettings'];
}

const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export const SettingsProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const settingsHook = useSettingsHook();

  return (
    <SettingsContext.Provider value={settingsHook}>
      {children}
    </SettingsContext.Provider>
  );
};

export const useSettings = () => {
  const context = useContext(SettingsContext);
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider');
  }
  return context;
};
