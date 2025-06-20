
"use client";

import type React from 'react';
import type { Shard, Position } from '@/types/shard-legends';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';
import Image from 'next/image';

interface ShardCellProps {
  shard: Shard | null;
  position: Position;
  isSelected: boolean;
  isHint?: boolean;
  onShardClick: (position: Position) => void;
}

const ShardCell: React.FC<ShardCellProps> = ({
  shard,
  position,
  isSelected,
  isHint,
  onShardClick,
}) => {
  const isDarkerSquare = (position.row + position.col) % 2 === 0;

  const cellBgClass = isDarkerSquare ? "bg-background" : "bg-card";
  const shardBgClass = "bg-card/50 backdrop-blur-sm";

  const LucideShardComponent = shard?.type.iconType === 'lucide' ? shard.type.component : null;

  return (
    <div
      className={cn(
        "shard-cell relative",
        cellBgClass,
        isSelected && shard && "selected",
        isHint && shard && "animate-pulse ring-2 ring-green-400"
      )}
      onClick={() => shard && onShardClick(position)}
      aria-label={shard ? `Shard type ${shard.type.name} at row ${position.row}, column ${position.col}` : `Empty cell at row ${position.row}, column ${position.col}`}
    >
      <AnimatePresence mode="wait">
        {shard && shard.type.iconType === 'lucide' && LucideShardComponent && (
          <motion.div
            key={`${shard.id}-lucide`}
            layout
            initial={{ opacity: 0, y: -60, scale: 0.5 }}
            animate={{
              opacity: shard.isMatched ? 0.2 : 1,
              scale: shard.isMatched ? 0.5 : 1,
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
              shardBgClass,
              "rounded-md shadow-lg"
            )}
          >
            <LucideShardComponent
              className={cn("w-3/4 h-3/4", shard.type.colorClass)}
            />
          </motion.div>
        )}
        {shard && shard.type.iconType === 'image' && shard.type.imageSrc && (
          <motion.div
            key={`${shard.id}-image`}
            layout
            initial={{ opacity: 0, y: -60, scale: 0.5 }}
            animate={{
              opacity: shard.isMatched ? 0.2 : 1,
              scale: shard.isMatched ? 0.5 : 1,
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
              shardBgClass,
              "rounded-md shadow-lg"
            )}
          >
            <div className="relative w-5/6 h-5/6">
              <Image
                src={shard.type.imageSrc}
                alt={shard.type.name}
                fill
                style={{ objectFit: 'contain' }}
                priority
              />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default ShardCell;
