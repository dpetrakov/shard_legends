
import { Gem, Diamond, Star, Heart, Square, Candy, Lollipop, Cookie, CakeSlice, IceCream2, Skull, Moon, Ghost, Bone, VenetianMask, Bird, Cat, Dog, Fish, Rabbit } from 'lucide-react';
import type { ShardIcon, IconSetType } from '@/types/shard-legends';

export const CLASSIC_SHARD_ICONS: ShardIcon[] = [
  { iconType: 'lucide', component: Gem, name: 'Gem', colorClass: 'text-[hsl(var(--shard-red))]' },
  { iconType: 'lucide', component: Diamond, name: 'Diamond', colorClass: 'text-[hsl(var(--shard-sky-blue))]' },
  { iconType: 'lucide', component: Star, name: 'Star', colorClass: 'text-[hsl(var(--shard-yellow))]' },
  { iconType: 'lucide', component: Heart, name: 'Heart', colorClass: 'text-[hsl(var(--shard-pink))]' },
  { iconType: 'lucide', component: Square, name: 'Square', colorClass: 'text-[hsl(var(--shard-green))]' },
];

export const SWEET_SHARD_ICONS: ShardIcon[] = [
  { iconType: 'lucide', component: Lollipop, name: 'Lollipop', colorClass: 'text-[hsl(var(--candy-red))]' },
  { iconType: 'lucide', component: Candy, name: 'WrappedCandy', colorClass: 'text-[hsl(var(--candy-pink))]' },
  { iconType: 'lucide', component: Cookie, name: 'Cookie', colorClass: 'text-[hsl(var(--candy-brown))]' },
  { iconType: 'lucide', component: CakeSlice, name: 'CakeSlice', colorClass: 'text-[hsl(var(--candy-yellow))]' },
  { iconType: 'lucide', component: IceCream2, name: 'IceCream', colorClass: 'text-[hsl(var(--candy-lightblue))]' },
];

export const GOTHIC_SHARD_ICONS: ShardIcon[] = [
  { iconType: 'lucide', component: Skull, name: 'Skull', colorClass: 'text-[hsl(var(--gothic-crimson))]' },
  { iconType: 'lucide', component: Moon, name: 'Moon', colorClass: 'text-[hsl(var(--gothic-midnight-blue))]' },
  { iconType: 'lucide', component: Ghost, name: 'Ghost', colorClass: 'text-[hsl(var(--gothic-ectoplasm))]' },
  { iconType: 'lucide', component: Bone, name: 'Bone', colorClass: 'text-[hsl(var(--gothic-deep-purple))]' },
  { iconType: 'lucide', component: VenetianMask, name: 'VenetianMask', colorClass: 'text-[hsl(var(--gothic-darkwood))]' },
];

export const ANIMAL_SHARD_ICONS: ShardIcon[] = [
  { iconType: 'lucide', component: Bird, name: 'Bird', colorClass: 'text-[hsl(var(--animal-green))]' },
  { iconType: 'lucide', component: Cat, name: 'Cat', colorClass: 'text-[hsl(var(--animal-orange))]' },
  { iconType: 'lucide', component: Dog, name: 'Dog', colorClass: 'text-[hsl(var(--animal-red))]' },
  { iconType: 'lucide', component: Fish, name: 'Fish', colorClass: 'text-[hsl(var(--animal-blue))]' },
  { iconType: 'lucide', component: Rabbit, name: 'Rabbit', colorClass: 'text-[hsl(var(--animal-pink))]' },
];

export const IN_MATCH3_ICONS: ShardIcon[] = [
  { iconType: 'image', name: 'RedPNG', imageSrc: '/images/red.png', colorClass: '' },
  { iconType: 'image', name: 'BluePNG', imageSrc: '/images/blue.png', colorClass: '' },
  { iconType: 'image', name: 'GreenPNG', imageSrc: '/images/green.png', colorClass: '' },
  { iconType: 'image', name: 'YellowPNG', imageSrc: '/images/yellow.png', colorClass: '' },
  { iconType: 'image', name: 'VioletPNG', imageSrc: '/images/violet.png', colorClass: '' },
];

export const ICON_SETS: Record<IconSetType, ShardIcon[]> = {
  classic: CLASSIC_SHARD_ICONS,
  sweets: SWEET_SHARD_ICONS,
  gothic: GOTHIC_SHARD_ICONS,
  animals: ANIMAL_SHARD_ICONS,
  'in-match3': IN_MATCH3_ICONS,
};

export const BOARD_ROWS = 6;
export const BOARD_COLS = 6;
