import { motion } from 'framer-motion';
import { pageVariants } from './variants';

interface PageTransitionProps {
  children: React.ReactNode;
}

/**
 * Page wrapper with fade-in animation
 */
export const PageTransition: React.FC<PageTransitionProps> = ({ children }) => {
  return (
    <motion.div
      initial="initial"
      animate="in"
      exit="out"
      variants={pageVariants}
      transition={{ type: 'tween', ease: 'anticipate', duration: 0.3 }}
    >
      {children}
    </motion.div>
  );
};
