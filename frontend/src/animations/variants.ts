import { Variants } from 'framer-motion';

/**
 * Fade in animation variants
 */
export const fadeIn: Variants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1 },
};

/**
 * Slide up animation variants
 */
export const slideUp: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
  },
};

/**
 * Slide in from left
 */
export const slideInLeft: Variants = {
  hidden: { opacity: 0, x: -20 },
  visible: {
    opacity: 1,
    x: 0,
  },
};

/**
 * Scale in animation
 */
export const scaleIn: Variants = {
  hidden: { opacity: 0, scale: 0.9 },
  visible: {
    opacity: 1,
    scale: 1,
  },
};

/**
 * Stagger container for list items
 */
export const staggerContainer: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.1,
    },
  },
};

/**
 * Stagger item (child of staggerContainer)
 */
export const staggerItem: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
  },
};

/**
 * Page transition variants
 */
export const pageVariants: Variants = {
  initial: { opacity: 0, y: 10 },
  in: { opacity: 1, y: 0 },
  out: { opacity: 0, y: -10 },
};

/**
 * Hover animation for cards
 */
export const hoverCard = {
  whileHover: { y: -4 },
  whileTap: { scale: 0.98 },
};

/**
 * Pulse animation for emphasis
 */
export const pulse: Variants = {
  initial: { scale: 1 },
  animate: { scale: [1, 1.05, 1] },
};

/**
 * Shimmer/skeleton loading animation
 */
export const shimmer: Variants = {
  initial: { backgroundPosition: '-1000px 0' },
  animate: { backgroundPosition: '1000px 0' },
};
