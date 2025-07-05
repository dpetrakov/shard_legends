"use client";

import type React from 'react';
import type { Crystal, Position } from '@/types/crystal-cascade';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';

interface CrystalCellProps {
  crystal: Crystal | null;
  position: Position;
  isSelected: boolean;
  isHint?: boolean;
  onCrystalClick: (position: Position) => void;
}

const CrystalCell: React.FC<CrystalCellProps> = ({
  crystal,
  position,
  isSelected,
  isHint,
  onCrystalClick,
}) => {
  const isDarkerSquare = (position.row + position.col) % 2 === 0;

  const cellBgClass = isDarkerSquare ? "bg-background" : "bg-card";
  const crystalBgClass = "bg-card/50 backdrop-blur-sm";

  const LucideCrystalComponent = crystal?.type.iconType === 'lucide' ? crystal.type.component : null;

  return (
    <div
      className={cn(
        "crystal-cell relative",
        cellBgClass,
        isHint && crystal && "animate-pulse ring-2 ring-green-400"
      )}
      onClick={() => crystal && onCrystalClick(position)}
      aria-label={crystal ? `Crystal type ${crystal.type.name} at row ${position.row}, column ${position.col}` : `Empty cell at row ${position.row}, column ${position.col}`}
    >
      <AnimatePresence>
        {crystal && (
          <motion.div
            key={crystal.id}
            layout
            initial={{ y: -60, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{
              opacity: 0,
              scale: 0.5,
              transition: { duration: 0.3, ease: "easeIn" }
            }}
            transition={{ type: 'tween', ease: 'easeInOut', duration: 0.4 }}
            className={cn(
              "absolute inset-0 flex items-center justify-center cursor-pointer",
              crystalBgClass,
              "rounded-md shadow-lg backface-hidden",
              isSelected && "z-10 ring-2 ring-accent"
            )}
            style={{ transform: 'translateZ(0)' }}
          >
            {crystal.type.iconType === 'lucide' && LucideCrystalComponent ? (
              <LucideCrystalComponent
                className={cn("w-3/4 h-3/4", crystal.type.colorClass)}
              />
            ) : crystal.type.iconType === 'image' && crystal.type.imageSrc ? (
                <img
                  src={crystal.type.imageSrc}
                  alt={crystal.type.name}
                  className="object-contain w-5/6 h-5/6 backface-hidden"
                  style={{ transform: 'translateZ(0)' }}
                />
            ) : null}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default CrystalCell;
