
"use client";

import type React from 'react';
import type { Crystal, Position } from '@/types/crystal-cascade';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';
import Image from 'next/image';

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
      <AnimatePresence mode="wait">
        {crystal && (
          <motion.div
            key={`${crystal.id}-${crystal.type.name}`}
            layout
            initial={{ opacity: 0, y: -60, scale: 0.5 }}
            animate={{
              opacity: crystal.isMatched ? 0.2 : 1,
              scale: crystal.isMatched ? 0.5 : 1,
              y: 0,
              transition: { duration: 0.3, ease: "easeInOut" }
            }}
            exit={{
              opacity: 0,
              scale: 0,
              transition: { duration: 0.3, ease: "easeIn" }
            }}
            className={cn(
              "absolute inset-0 flex items-center justify-center cursor-pointer",
              crystalBgClass,
              "rounded-md shadow-lg backface-hidden",
              isSelected && "z-10 ring-2 ring-accent"
            )}
          >
            <div className="relative w-full h-full backface-hidden">
              {crystal.type.iconType === 'lucide' && LucideCrystalComponent ? (
                <LucideCrystalComponent
                  className={cn("w-3/4 h-3/4 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2", crystal.type.colorClass)}
                />
              ) : crystal.type.iconType === 'image' && crystal.type.imageSrc ? (
                  <Image
                    src={crystal.type.imageSrc}
                    alt={crystal.type.name}
                    width={128}
                    height={128}
                    className="object-contain w-5/6 h-5/6 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 backface-hidden"
                  />
              ) : null}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default CrystalCell;
