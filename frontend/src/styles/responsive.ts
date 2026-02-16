// Breakpoints for responsive design
export const breakpoints = {
  mobile: '0px',      // 0 - 767px
  tablet: '768px',    // 768px - 1023px
  desktop: '1024px'   // 1024px and above
} as const;

// Media query helpers
export const media = {
  mobile: `@media (max-width: ${breakpoints.tablet})`,
  tablet: `@media (min-width: ${breakpoints.tablet}) and (max-width: ${breakpoints.desktop})`,
  desktop: `@media (min-width: ${breakpoints.desktop})`,
  tabletAndUp: `@media (min-width: ${breakpoints.tablet})`,
  mobileAndTablet: `@media (max-width: ${breakpoints.desktop})`
} as const;

// Check if device is mobile
export const isMobileDevice = (): boolean => {
  return window.innerWidth < parseInt(breakpoints.tablet);
};

// Check if device is tablet
export const isTabletDevice = (): boolean => {
  const width = window.innerWidth;
  return width >= parseInt(breakpoints.tablet) && width < parseInt(breakpoints.desktop);
};

// Check if device is desktop
export const isDesktopDevice = (): boolean => {
  return window.innerWidth >= parseInt(breakpoints.desktop);
};

// Get current device type
export type DeviceType = 'mobile' | 'tablet' | 'desktop';

export const getDeviceType = (): DeviceType => {
  const width = window.innerWidth;
  if (width < parseInt(breakpoints.tablet)) return 'mobile';
  if (width < parseInt(breakpoints.desktop)) return 'tablet';
  return 'desktop';
};
