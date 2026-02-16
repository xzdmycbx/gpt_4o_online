// Global theme colors
export const colors = {
  // Dark theme (default)
  dark: {
    background: '#0e1621',
    surface: '#1e2832',
    surfaceHighlight: '#2b3642',
    primary: '#2b5278',
    primaryHover: '#3a6a95',
    text: '#e8eaed',
    textSecondary: '#9aa0a6',
    textMuted: '#5f6368',
    border: '#3c4043',
    error: '#f28b82',
    success: '#81c995',
    warning: '#fdd663',
    // Message bubbles
    userBubble: '#2b5278',
    aiBubble: '#1e2832',
  },

  // Light theme
  light: {
    background: '#ffffff',
    surface: '#f8f9fa',
    surfaceHighlight: '#e8eaed',
    primary: '#1a73e8',
    primaryHover: '#1557b0',
    text: '#202124',
    textSecondary: '#5f6368',
    textMuted: '#9aa0a6',
    border: '#dadce0',
    error: '#d93025',
    success: '#1e8e3e',
    warning: '#f9ab00',
    // Message bubbles
    userBubble: '#1a73e8',
    aiBubble: '#f8f9fa',
  },
} as const;

// Typography
export const typography = {
  fontFamily: {
    base: `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif`,
    mono: `'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace`,
  },
  fontSize: {
    small: {
      xs: '0.7rem',
      sm: '0.8rem',
      base: '0.875rem',
      lg: '1rem',
      xl: '1.125rem',
      '2xl': '1.25rem',
    },
    medium: {
      xs: '0.75rem',
      sm: '0.875rem',
      base: '1rem',
      lg: '1.125rem',
      xl: '1.25rem',
      '2xl': '1.5rem',
    },
    large: {
      xs: '0.875rem',
      sm: '1rem',
      base: '1.125rem',
      lg: '1.25rem',
      xl: '1.5rem',
      '2xl': '1.75rem',
    },
  },
  fontWeight: {
    normal: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
  },
} as const;

// Spacing
export const spacing = {
  xs: '0.25rem',   // 4px
  sm: '0.5rem',    // 8px
  md: '1rem',      // 16px
  lg: '1.5rem',    // 24px
  xl: '2rem',      // 32px
  '2xl': '3rem',   // 48px
  '3xl': '4rem',   // 64px
} as const;

// Border radius
export const borderRadius = {
  sm: '4px',
  md: '8px',
  lg: '12px',
  xl: '16px',
  full: '9999px',
} as const;

// Shadows
export const shadows = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
} as const;

// Z-index layers
export const zIndex = {
  base: 0,
  dropdown: 1000,
  sticky: 1020,
  fixed: 1030,
  modalBackdrop: 1040,
  modal: 1050,
  popover: 1060,
  tooltip: 1070,
} as const;

// Transitions
export const transitions = {
  fast: '150ms ease-in-out',
  normal: '250ms ease-in-out',
  slow: '350ms ease-in-out',
} as const;
