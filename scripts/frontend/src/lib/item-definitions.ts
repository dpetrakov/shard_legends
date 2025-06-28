
import type { InventoryItemType } from '@/types/inventory';

export const itemNames: Record<InventoryItemType, string> = {
  // Resources
  stone: 'Камень',
  wood: 'Дерево',
  ore: 'Руда',
  diamond: 'Алмаз',
  // Reagents
  abrasive: 'Абразив',
  disc: 'Диск',
  inductor: 'Индуктор',
  paste: 'Паста',
  // Blueprints
  shovel: 'Чертеж лопаты',
  sickle: 'Чертеж серпа',
  axe: 'Чертеж топора',
  pickaxe: 'Чертеж кирки',
  // Boosters (placeholders)
  speed_1h: 'Ускоритель (1ч)',
  speed_3h: 'Ускоритель (3ч)',
  speed_12h: 'Ускоритель (12ч)',
  // Processed Items
  wood_plank: 'Деревянный брусок',
  stone_block: 'Каменный блок',
  metal_ingot: 'Металлический слиток',
  cut_diamond: 'Бриллиант',
  // Crafted Tools
  wooden_shovel: 'Деревянная лопата',
  stone_shovel: 'Каменная лопата',
  metal_shovel: 'Металлическая лопата',
  diamond_shovel: 'Бриллиантовая лопата',
  wooden_sickle: 'Деревянный серп',
  stone_sickle: 'Каменный серп',
  metal_sickle: 'Металлический серп',
  diamond_sickle: 'Бриллиантовый серп',
  wooden_axe: 'Деревянный топор',
  stone_axe: 'Каменный топор',
  metal_axe: 'Металлический топор',
  diamond_axe: 'Бриллиантовый топор',
  wooden_pickaxe: 'Деревянная кирка',
  stone_pickaxe: 'Каменная кирка',
  metal_pickaxe: 'Металлическая кирка',
  diamond_pickaxe: 'Бриллиантовая кирка',
};
