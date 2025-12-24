/**
 * Design Tokens for Assistant Qisumi
 * Modern, Linear-inspired design system with responsive support
 */

export const designTokens = {
  // Color Palette - Linear-inspired
  colors: {
    // Primary colors
    primary: '#5E6AD2',
    primaryHover: '#4E5AC4',
    primaryActive: '#3E4AB6',

    // Semantic colors
    success: '#27AE60',
    warning: '#F2994A',
    error: '#EB5757',
    info: '#2D9CDB',

    // Neutral grays
    gray: {
      50: '#F7F7F8',
      100: '#E8E8EB',
      200: '#D3D3D7',
      300: '#B4B4BC',
      400: '#8F8F99',
      500: '#6B6B76',
      600: '#52525B',
      700: '#3F3F46',
      800: '#27272A',
      900: '#18181B',
    },

    // Background colors
    bg: {
      base: '#FFFFFF',
      secondary: '#F7F7F8',
      tertiary: '#E8E8EB',
      overlay: 'rgba(0, 0, 0, 0.5)',
    },

    // Text colors
    text: {
      primary: '#18181B',
      secondary: '#6B6B76',
      tertiary: '#A1A1AA',
      inverse: '#FFFFFF',
    },
  },

  // Spacing scale (8px base unit)
  spacing: {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
    xxxl: 64,
  },

  // Typography scale with responsive support
  typography: {
    fontFamily: {
      base: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      mono: '"SF Mono", "Monaco", "Cascadia Code", "Roboto Mono", Consolas, monospace',
    },
    // Base font sizes (for desktop/large screens)
    fontSize: {
      xs: '13px',
      sm: '15px',
      base: '16px',
      lg: '18px',
      xl: '20px',
      '2xl': '24px',
      '3xl': '30px',
      '4xl': '36px',
      '5xl': '48px',
    },
    // Responsive font sizes (scaled by screen size)
    responsiveFontSize: {
      xs: { mobile: '12px', tablet: '13px', desktop: '13px' },
      sm: { mobile: '14px', tablet: '15px', desktop: '15px' },
      base: { mobile: '15px', tablet: '16px', desktop: '16px' },
      lg: { mobile: '17px', tablet: '18px', desktop: '19px' },
      xl: { mobile: '19px', tablet: '20px', desktop: '21px' },
      '2xl': { mobile: '22px', tablet: '24px', desktop: '26px' },
      '3xl': { mobile: '28px', tablet: '30px', desktop: '32px' },
      '4xl': { mobile: '34px', tablet: '36px', desktop: '40px' },
    },
    fontWeight: {
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
    lineHeight: {
      tight: 1.25,
      normal: 1.5,
      relaxed: 1.75,
    },
  },

  // Border radius
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },

  // Shadows
  shadow: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
    inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
  },

  // Animation durations
  animation: {
    fast: '150ms',
    base: '200ms',
    slow: '300ms',
  },

  // Z-index scale
  zIndex: {
    dropdown: 1000,
    sticky: 1020,
    fixed: 1030,
    modalBackdrop: 1040,
    modal: 1050,
    popover: 1060,
    tooltip: 1070,
  },

  // Responsive breakpoints
  breakpoints: {
    mobile: '576px',
    tablet: '768px',
    desktop: '992px',
    large: '1200px',
    wide: '1600px',
  },
} as const;

export type DesignTokens = typeof designTokens;

// Helper type for screen size
export type ScreenSize = 'mobile' | 'tablet' | 'desktop';
