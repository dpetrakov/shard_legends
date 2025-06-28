
export type ChestType =
  | 'resource_small'
  | 'resource_medium'
  | 'resource_large'
  | 'reagent_small'
  | 'reagent_medium'
  | 'reagent_large'
  | 'booster_small'
  | 'booster_medium'
  | 'booster_large'
  | 'blueprint';

export type ChestCounts = { [key in ChestType]?: number };

export interface ChestContextType {
  chestCounts: ChestCounts;
  awardChest: (chestType: ChestType) => void;
  spendChests: (chestType: ChestType, amount: number) => void;
  getChestName: (chestType: ChestType) => string;
}
