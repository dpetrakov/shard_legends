
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
        isSelected && crystal && "selected",
        isHint && crystal && "animate-pulse ring-2 ring-green-400"
      )}
      onClick={() => crystal && onCrystalClick(position)}
      aria-label={crystal ? `Crystal type ${crystal.type.name} at row ${position.row}, column ${position.col}` : `Empty cell at row ${position.row}, column ${position.col}`}
    >
      <AnimatePresence mode="wait">
        {crystal && crystal.type.iconType === 'lucide' && LucideCrystalComponent && (
          <motion.div
            key={`${crystal.id}-lucide`}
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
              scale: 0.3,
              y: 60,
              transition: { duration: 0.2, ease: "anticipate" }
            }}
            className={cn(
              "absolute inset-0 flex items-center justify-center cursor-pointer",
              crystalBgClass,
              "rounded-md shadow-lg"
            )}
          >
            <LucideCrystalComponent
              className={cn("w-3/4 h-3/4", crystal.type.colorClass)}
            />
          </motion.div>
        )}
        {crystal && crystal.type.iconType === 'image' && crystal.type.imageSrc && (
          <motion.div
            key={`${crystal.id}-image`}
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
              scale: 0.3,
              y: 60,
              transition: { duration: 0.2, ease: "anticipate" }
            }}
            className={cn(
              "absolute inset-0 flex items-center justify-center cursor-pointer",
              crystalBgClass,
              "rounded-md shadow-lg"
            )}
          >
            <div className="relative w-5/6 h-5/6">
              <Image
                src={crystal.type.imageSrc}
                alt={crystal.type.name}
                layout="fill"
                objectFit="contain"
                priority
              />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default CrystalCell;
