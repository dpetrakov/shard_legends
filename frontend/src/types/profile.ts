
export type ChestType = 'small' | 'medium' | 'large';

export interface ChestData {
  type: ChestType;
  name: string;
  imageHint: string;
}

export interface ChestCounts {
  small: number;
  medium: number;
  large: number;
}

export interface ChestContextType {
  chestCounts: ChestCounts;
  awardChest: (chestType: ChestType) => void;
  getChestName: (chestType: ChestType) => string;
}
