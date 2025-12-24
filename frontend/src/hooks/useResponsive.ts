import { useMemo } from 'react';
import { Grid } from 'antd';
import type { ScreenSize } from '@/theme/tokens';

const { useBreakpoint } = Grid;

/**
 * Hook for responsive design utilities
 * Returns the current screen size and responsive helpers
 */
export const useResponsive = () => {
  const screens = useBreakpoint();

  const screenSize = useMemo<ScreenSize>(() => {
    if (screens.lg) return 'desktop';
    if (screens.md) return 'tablet';
    return 'mobile';
  }, [screens]);

  const isMobile = screenSize === 'mobile';
  const isTablet = screenSize === 'tablet';
  const isDesktop = screenSize === 'desktop';
  const isMobileOrTablet = isMobile || isTablet;

  return {
    screenSize,
    isMobile,
    isTablet,
    isDesktop,
    isMobileOrTablet,
    // Ant Design breakpoints
    screens,
  };
};

/**
 * Hook to get responsive font size based on current screen size
 */
export const useResponsiveFontSize = (
  sizeKey: keyof typeof import('@/theme/tokens').designTokens.typography.responsiveFontSize
): string => {
  const { screenSize } = useResponsive();

  // Import and use the responsive font sizes
  const responsiveFontSize = {
    xs: { mobile: '12px', tablet: '13px', desktop: '13px' },
    sm: { mobile: '14px', tablet: '15px', desktop: '15px' },
    base: { mobile: '15px', tablet: '16px', desktop: '16px' },
    lg: { mobile: '17px', tablet: '18px', desktop: '19px' },
    xl: { mobile: '19px', tablet: '20px', desktop: '21px' },
    '2xl': { mobile: '22px', tablet: '24px', desktop: '26px' },
    '3xl': { mobile: '28px', tablet: '30px', desktop: '32px' },
    '4xl': { mobile: '34px', tablet: '36px', desktop: '40px' },
  };

  return responsiveFontSize[sizeKey][screenSize];
};

/**
 * Hook to get responsive spacing value
 */
export const useResponsiveSpacing = (
  baseSpacing: number,
  mobileMultiplier = 0.75,
  tabletMultiplier = 0.875
): number => {
  const { screenSize } = useResponsive();

  if (screenSize === 'mobile') return baseSpacing * mobileMultiplier;
  if (screenSize === 'tablet') return baseSpacing * tabletMultiplier;
  return baseSpacing;
};

/**
 * Hook to get responsive value based on screen size
 */
export const useResponsiveValue = <T,>(values: {
  mobile: T;
  tablet?: T;
  desktop: T;
}): T => {
  const { screenSize } = useResponsive();

  if (screenSize === 'mobile') return values.mobile;
  if (screenSize === 'tablet') return values.tablet ?? values.desktop;
  return values.desktop;
};
