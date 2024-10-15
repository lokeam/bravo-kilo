import { motion } from 'framer-motion';
import { ReactNode } from 'react';

type PageTransitionProps = {
  children: ReactNode;
};

const pageVariants = {
  initial: {
    opacity: 0,
  },
  in: {
    opacity: 1,
  },
  out: {
    opacity: 0,
  },
};

const pageTransition = {
  type: 'tween',
  ease: 'easeIn',
  duration: '0.15'
}

const PageTransition = ({ children }: PageTransitionProps) => {
  return (
    <motion.div
      initial="initial"
      animate="in"
      exit="out"
      variants={pageVariants}
      transition={pageTransition}
    >
      { children }
    </motion.div>
  )
};

export default PageTransition;
