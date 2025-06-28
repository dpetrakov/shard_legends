
"use client";

import Image from 'next/image';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Crown, Lock, CheckCircle2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import React from 'react';

// Dummy data for rewards
const BATTLE_PASS_LEVELS = [
  {
    level: 1,
    premium: { type: 'chest', id: 'reagent_small', amount: 1, icon: '/images/small-chess-ing.png' },
    free: { type: 'chest', id: 'resource_small', amount: 1, icon: '/images/small-chess-res.png' },
  },
  {
    level: 2,
    premium: { type: 'resource', id: 'wood', amount: 500, icon: '/images/wood.png' },
    free: { type: 'reagent', id: 'abrasive', amount: 10, icon: '/images/ing-abraziv.png' },
  },
  {
    level: 3,
    premium: { type: 'chest', id: 'resource_medium', amount: 1, icon: '/images/medium-chess-res.png' },
    free: { type: 'resource', id: 'stone', amount: 250, icon: '/images/stone.png' },
  },
  {
    level: 4,
    premium: { type: 'blueprint', id: 'axe', amount: 1, icon: '/images/blueprint-axie.png' },
    free: { type: 'reagent', id: 'disc', amount: 10, icon: '/images/ing-disk.png' },
  },
  {
    level: 5,
    premium: { type: 'chest', id: 'reagent_medium', amount: 1, icon: '/images/medium-chess-ing.png' },
    free: { type: 'chest', id: 'resource_small', amount: 2, icon: '/images/small-chess-res.png' },
  },
];

const CURRENT_LEVEL = 3;
const IS_PREMIUM = false;
const CLAIMED_FREE_LEVELS = [1, 2];
const CLAIMED_PREMIUM_LEVELS: number[] = [];
const CARDS_FOR_NEXT_REWARD = 3;
const TOTAL_CARDS_FOR_REWARD = 5;

interface Reward {
  type: string;
  id: string;
  amount: number;
  icon: string;
}

const RewardCard = ({ reward, state }: { reward: Reward | null, state: 'locked' | 'claimable' | 'claimed' | 'unlocked' | 'unavailable' }) => {
  if (!reward) {
    return <div className="aspect-square w-full invisible" />;
  }

  const isClaimed = state === 'claimed';
  const isLocked = state === 'locked';

  return (
    <Card className={cn(
      "relative w-full aspect-[4/5] sm:aspect-square transition-all duration-300 transform-gpu",
      "bg-card/60 border-border/50",
      isClaimed && "bg-black/30 border-green-500/50",
      state === 'claimable' && "border-primary shadow-lg shadow-primary/50 scale-105",
      isLocked && "grayscale opacity-70"
    )}>
      <CardContent className="p-2 flex flex-col items-center justify-center h-full">
        <div className="relative">
          <Image src={reward.icon} alt={reward.id} width={64} height={64} className="flex-grow object-contain max-w-[80%] max-h-[80%] mx-auto" />
          <span className="absolute bottom-0 right-0 bg-background/80 text-foreground font-bold text-xs px-1.5 py-0.5 rounded-full shadow-md">
            x{reward.amount}
          </span>
        </div>
      </CardContent>
      {isLocked && !IS_PREMIUM && reward && (
        <div className="absolute inset-0 bg-black/50 rounded-lg flex items-center justify-center">
          <Lock className="w-8 h-8 text-white/70" />
        </div>
      )}
      {isClaimed && (
        <div className="absolute inset-0 bg-black/50 rounded-lg flex items-center justify-center">
          <CheckCircle2 className="w-8 h-8 text-green-400" />
        </div>
      )}
    </Card>
  );
};

export default function BattlePassPage() {
    return (
        <div className="flex flex-col items-center justify-start min-h-full bg-background text-foreground space-y-4 pt-4 pb-28">
            {/* Header */}
            <div className="w-full max-w-md relative text-center text-white">
                <Image
                    src="https://placehold.co/600x200.png"
                    alt="Охотник за наградами"
                    width={600}
                    height={200}
                    className="w-full h-auto object-cover"
                    data-ai-hint="fantasy landscape"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-background via-background/50 to-transparent" />
                <div className="absolute bottom-0 left-0 right-0 p-4 flex flex-col items-center">
                    <p className="text-sm text-primary font-semibold">Еженедельный сезон</p>
                    <h1 className="text-3xl font-headline text-shadow-lg">Охотник за наградами</h1>
                    <Button size="lg" className="mt-3 bg-yellow-400 hover:bg-yellow-500 text-black font-bold shadow-xl">
                        <Crown className="mr-2 fill-current" /> Купить пропуск
                    </Button>
                </div>
            </div>

            {/* Progress Bar */}
             <div className="w-full max-w-md px-4">
                <Card className="p-3 bg-card/80 backdrop-blur-md">
                    <div className="flex items-center justify-between text-xs sm:text-sm">
                        <div className="flex items-center gap-2">
                             <div className="p-1.5 bg-muted rounded-md">
                                <Image src="/images/battlepass.png" alt="pass icon" width={24} height={24}/>
                            </div>
                            <span className="font-bold">Уровень: {CURRENT_LEVEL} / {BATTLE_PASS_LEVELS.length}</span>
                        </div>
                        <div className="flex flex-col items-end">
                            <span className="font-semibold text-muted-foreground">До завершения:</span>
                            <span className="font-bold text-primary">6д 18ч</span>
                        </div>
                    </div>
                    <div className="mt-3">
                        <div className="text-center text-xs text-muted-foreground mb-2">
                            Прогресс до следующей награды: {CARDS_FOR_NEXT_REWARD}/{TOTAL_CARDS_FOR_REWARD} карт
                        </div>
                        <div className="flex justify-center gap-2">
                            {Array.from({ length: TOTAL_CARDS_FOR_REWARD }).map((_, index) => (
                                <div key={index} className={cn(
                                    "w-10 h-14 rounded transition-all duration-300",
                                     index < CARDS_FOR_NEXT_REWARD ? "opacity-100 ring-1 ring-primary/80" : "opacity-40"
                                )}>
                                     <Image 
                                        src="/images/card/card-back.jpg" 
                                        alt="Card back" 
                                        width={40} 
                                        height={56} 
                                        className="w-full h-full object-cover rounded"
                                    />
                                </div>
                            ))}
                        </div>
                    </div>
                </Card>
            </div>

            {/* Reward Tracks */}
            <div className="w-full max-w-md px-4 space-y-4">
                 <div className="grid grid-cols-2 gap-4 text-center">
                    <h2 className="text-xl font-headline text-primary">Премиум награды</h2>
                    <h2 className="text-xl font-headline text-primary">Бесплатные награды</h2>
                </div>

                <div className="relative space-y-3">
                     {/* Central Line */}
                     <div className="absolute top-0 bottom-0 left-1/2 w-1 bg-primary/20 -translate-x-1/2" />
                    
                    {BATTLE_PASS_LEVELS.map(({ level, premium, free }) => {
                        const isUnlocked = level <= CURRENT_LEVEL;
                        
                        const freeState = free
                            ? CLAIMED_FREE_LEVELS.includes(level) ? 'claimed' : isUnlocked ? 'claimable' : 'locked'
                            : 'unavailable';
                            
                        const premiumState = premium
                            ? CLAIMED_PREMIUM_LEVELS.includes(level) ? 'claimed' : isUnlocked ? (IS_PREMIUM ? 'claimable' : 'locked') : 'locked'
                            : 'unavailable';

                        return (
                            <React.Fragment key={level}>
                                <div className="relative grid grid-cols-[1fr_auto_1fr] items-center gap-3 sm:gap-4 z-10">
                                    {/* Premium Reward */}
                                    <div className="flex justify-end">
                                        <div className="w-24 sm:w-28">
                                            <RewardCard reward={premium} state={premiumState} />
                                        </div>
                                    </div>

                                    {/* Level Marker */}
                                    <div className={cn(
                                        "w-10 h-10 sm:w-12 sm:h-12 flex items-center justify-center font-bold text-lg rotate-45 rounded-md",
                                        isUnlocked ? "bg-primary text-primary-foreground shadow-lg" : "bg-muted text-muted-foreground"
                                    )}>
                                        <span className="-rotate-45">{level}</span>
                                    </div>
                                    
                                    {/* Free Reward */}
                                    <div className="flex justify-start">
                                        <div className="w-24 sm:w-28">
                                            <RewardCard reward={free} state={freeState} />
                                        </div>
                                    </div>
                                </div>
                                <div className="relative grid grid-cols-2 gap-3 sm:gap-4 z-10 -mt-2">
                                     {premiumState === 'claimable' ? (
                                        <Button size="sm">Забрать</Button>
                                     ) : <div/>}

                                      {freeState === 'claimable' ? (
                                        <Button size="sm">Забрать</Button>
                                     ) : <div/>}
                                </div>
                                {level === CURRENT_LEVEL && (
                                     <div className="text-center -mt-2">
                                        <Button variant="secondary" size="sm">Ускорить прогресс</Button>
                                    </div>
                                )}
                            </React.Fragment>
                        )
                    })}
                </div>
            </div>
        </div>
    );
}
