import { useAnimation } from 'framer-motion';

/**
 * Hook for fade-in animation on mount
 */
export const useFadeIn = () => {
  const controls = useAnimation();

  const fadeIn = () => {
    controls.start('visible');
  };

  return { controls, fadeIn };
};

/**
 * Hook for staggered list animations
 */
export const useStaggeredList = () => {
  const controls = useAnimation();

  const animateIn = () => {
    controls.start('visible');
  };

  return { controls, animateIn };
};

/**
 * Hook for hover card animation
 */
export const useHoverCard = () => {
  return {
    whileHover: { y: -4 },
    whileTap: { scale: 0.98 },
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { type: 'spring', stiffness: 300, damping: 24 },
  };
};

/**
 * Hook for scale animation on hover
 */
export const useScaleOnHover = (scale = 1.02) => {
  return {
    whileHover: { scale },
    whileTap: { scale: 0.98 },
    transition: { type: 'spring', stiffness: 400, damping: 17 },
  };
};

/**
 * Hook for confetti celebration effect
 */
export const useConfetti = () => {
  const fireConfetti = (options?: {
    particleCount?: number;
    spread?: number;
    origin?: { x: number; y: number };
  }) => {
    const {
      particleCount = 100,
      spread = 70,
      origin = { y: 0.6 },
    } = options || {};

    // Import canvas-confetti dynamically to avoid SSR issues
    import('canvas-confetti').then((confetti) => {
      confetti.default({ particleCount, spread, origin });
    });
  };

  return { fireConfetti };
};
