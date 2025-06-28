
import type { CrystalIcon, IconSetType as ImportedIconSetType } from '@/types/crystal-cascade';

export type IconSet = ImportedIconSetType; // 'classic' | 'sweets' | 'gothic' | 'animals' | 'in-match3'

export interface IconSetContextType {
  iconSet: IconSet;
  setIconSet: (iconSet: IconSet) => void;
  getActiveIconList: () => CrystalIcon[];
}
