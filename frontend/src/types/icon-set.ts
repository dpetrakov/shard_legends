
import type { ShardIcon, IconSetType as ImportedIconSetType } from '@/types/shard-legends';

export type IconSet = ImportedIconSetType; // 'classic' | 'sweets' | 'gothic' | 'animals' | 'in-match3'

export interface IconSetContextType {
  iconSet: IconSet;
  setIconSet: (iconSet: IconSet) => void;
  getActiveIconList: () => ShardIcon[];
}
