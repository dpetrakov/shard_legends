
"use client";

import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { FloatingScoreItem } from '@/types/crystal-cascade';

interface FloatingScoreEffectProps {
  item: FloatingScoreItem;
  spawnElement: HTMLDivElement | null; 
  scoreElement: HTMLDivElement | null; 
  onAnimationComplete: (id: string) => void;
}

const FloatingScoreEffect: React.FC<FloatingScoreEffectProps> = ({
  item,
  spawnElement,
  scoreElement,
  onAnimationComplete,
}) => {
  const [isVisible, setIsVisible] = useState(true);
  const [initialPos, setInitialPos] = useState<{ x: number; y: number } | null>(null);
  const [targetPos, setTargetPos] = useState<{ x: number; y: number } | null>(null);

  useEffect(() => {
    if (spawnElement) {
      const spawnRect = spawnElement.getBoundingClientRect();
      // Spawn at the center of the spawnElement
      const startX = spawnRect.left + spawnRect.width / 2 + (Math.random() - 0.5) * spawnRect.width * 0.2; // Slight horizontal jitter
      const startY = spawnRect.top + spawnRect.height / 2; 
      setInitialPos({ x: startX, y: startY });
    }

    if (scoreElement) {
      const scoreRect = scoreElement.getBoundingClientRect();
      // Target the center of the score display element
      const endX = scoreRect.left + scoreRect.width / 2;
      const endY = scoreRect.top + scoreRect.height / 2;
      setTargetPos({ x: endX, y: endY });
    }
  }, [spawnElement, scoreElement]);

  if (!isVisible || !initialPos || !targetPos) {
    return null;
  }

  return (
    <motion.div
      key={item.id}
      initial={{
        x: initialPos.x,
        y: initialPos.y,
        opacity: 1,
        scale: 1.2, 
      }}
      animate={{
        x: targetPos.x,
        y: targetPos.y,
        opacity: 0,
        scale: 0.5, 
        transition: {
          duration: 0.8, 
          ease: "easeInOut",
        },
      }}
      exit={{ opacity: 0, scale: 0 }} 
      onAnimationComplete={() => {
        setIsVisible(false);
        onAnimationComplete(item.id);
      }}
      className="fixed text-accent font-headline text-xl z-[1000] pointer-events-none"
      style={{
        transformOrigin: 'center',
        textShadow: '1px 1px 2px rgba(0,0,0,0.3)', 
      }}
    >
      +{item.value}
    </motion.div>
  );
};


const FloatingScoreManager: React.FC<{
  floatingScores: FloatingScoreItem[];
  spawnElement: HTMLDivElement | null;
  scoreElement: HTMLDivElement | null;
  onAnimationComplete: (id: string) => void;
}> = ({ floatingScores, spawnElement, scoreElement, onAnimationComplete }) => {
  return (
    <AnimatePresence>
      {floatingScores.map((item) => (
        <FloatingScoreEffect
          key={item.key}
          item={item}
          spawnElement={spawnElement}
          scoreElement={scoreElement}
          onAnimationComplete={onAnimationComplete}
        />
      ))}
    </AnimatePresence>
  );
};

export default FloatingScoreManager;

    
