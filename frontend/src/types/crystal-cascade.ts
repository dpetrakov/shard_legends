
import type { LucideProps } from 'lucide-react';
import type React from 'react';

export type CrystalColor = 'red' | 'blue' | 'yellow' | 'pink' | 'green' | 'purple';

export interface CrystalIcon {
  iconType: 'lucide' | 'image';
  name: string;
  component?: React.ForwardRefExoticComponent<Omit<LucideProps, "ref"> & React.RefAttributes<SVGSVGElement>>; // Optional for image type
  imageSrc?: string; // Optional for lucide type
  colorClass: string; // For Lucide icons color, or fallback/border for images
}

export interface Crystal {
  id: number;
  type: CrystalIcon;
  row: number;
  col: number;
  isMatched?: boolean;
}

export type GameBoard = (Crystal | null)[][];

export interface Position {
  row: number;
  col: number;
}

export type IconSetType = 'classic' | 'sweets' | 'gothic' | 'animals' | 'in-match3';

export interface FloatingScoreItem {
  id: string;
  key: string;
  value: number;
  position: Position;
}
