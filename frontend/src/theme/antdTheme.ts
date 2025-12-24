import { ThemeConfig } from 'antd';
import { designTokens } from './tokens';

/**
 * Ant Design theme configuration using design tokens
 * Optimized for high-DPI displays with larger font sizes
 */
export const antdTheme: ThemeConfig = {
  token: {
    // Primary colors
    colorPrimary: designTokens.colors.primary,
    colorSuccess: designTokens.colors.success,
    colorWarning: designTokens.colors.warning,
    colorError: designTokens.colors.error,
    colorInfo: designTokens.colors.info,

    // Neutral colors
    colorBgBase: designTokens.colors.bg.base,
    colorBgContainer: designTokens.colors.bg.base,
    colorBgLayout: designTokens.colors.bg.secondary,
    colorBgSpotlight: designTokens.colors.gray[50],

    // Text colors
    colorText: designTokens.colors.text.primary,
    colorTextSecondary: designTokens.colors.text.secondary,
    colorTextTertiary: designTokens.colors.text.tertiary,
    colorTextQuaternary: designTokens.colors.gray[400],

    // Border colors
    colorBorder: designTokens.colors.gray[200],
    colorBorderSecondary: designTokens.colors.gray[100],

    // Typography (Ant Design expects numbers for fontSize, not strings)
    // Increased base font size for better readability on high-DPI displays
    fontFamily: designTokens.typography.fontFamily.base,
    fontSize: 15, // Increased from 14 to 15
    fontSizeHeading1: parseInt(designTokens.typography.fontSize['5xl']), // 48px
    fontSizeHeading2: parseInt(designTokens.typography.fontSize['4xl']), // 36px
    fontSizeHeading3: parseInt(designTokens.typography.fontSize['3xl']), // 30px
    fontSizeHeading4: parseInt(designTokens.typography.fontSize['2xl']), // 24px
    fontSizeHeading5: parseInt(designTokens.typography.fontSize.xl),   // 20px

    // Border radius
    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 4,
    borderRadiusXS: 2,

    // Spacing
    marginXS: designTokens.spacing.xs,
    marginSM: designTokens.spacing.sm,
    margin: designTokens.spacing.md,
    marginLG: designTokens.spacing.lg,
    marginXL: designTokens.spacing.xl,

    // Animation
    motionDurationFast: designTokens.animation.fast,
    motionDurationMid: designTokens.animation.base,
    motionDurationSlow: designTokens.animation.slow,
  },
  components: {
    Layout: {
      headerBg: designTokens.colors.bg.base,
      headerHeight: 64,
      siderBg: designTokens.colors.gray[900],
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: 'rgba(94, 106, 210, 0.1)',
      itemSelectedColor: designTokens.colors.primary,
      itemHoverBg: 'rgba(94, 106, 210, 0.05)',
      itemBorderRadius: 6,
    },
    Card: {
      borderRadiusLG: parseInt(designTokens.borderRadius.lg),
    },
    Button: {
      borderRadius: 6,
      controlHeight: 38, // Increased from 36 for better touch targets
      controlHeightLG: 46, // Increased from 44
      controlHeightSM: 30, // Increased from 28
      primaryShadow: designTokens.shadow.sm,
      paddingInline: 18, // Increased padding
    },
    Input: {
      borderRadius: 6,
      controlHeight: 38, // Increased from 36
      controlHeightLG: 46,
      controlHeightSM: 30,
      paddingInline: 14, // Increased padding
      activeBorderColor: designTokens.colors.primary,
      hoverBorderColor: designTokens.colors.primaryHover,
    },
    Modal: {
      borderRadiusLG: parseInt(designTokens.borderRadius.lg),
    },
    Tag: {
      borderRadiusSM: 4,
    },
    Select: {
      borderRadius: 6,
      controlHeight: 38, // Increased from 36
    },
    Form: {
      itemMarginBottom: designTokens.spacing.lg,
      verticalLabelPadding: '0 0 8px', // Increased padding
    },
    Tooltip: {
      colorBgElevated: '#fff',
      colorText: '#000',
      borderRadius: 6,
    },
    Typography: {
      margin: 0,
    },
    Table: {
      cellPaddingInline: 16, // Increased cell padding
      cellPaddingBlock: 12,
    },
    List: {
      itemPadding: '12px 16px', // Increased list item padding
    },
  },
};
