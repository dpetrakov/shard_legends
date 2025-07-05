
"use client";

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import Image from 'next/image';
import { ScrollArea } from "@/components/ui/scroll-area";
import { Award, Check } from 'lucide-react';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

interface Achievement {
  id: number;
  icon: string;
  iconHint: string;
  title: string;
  description: string;
  currentProgress: number;
  targetProgress: number;
  rewardIcon: string;
  rewardAmount: number;
  claimed: boolean;
}

const initialAchievementsData: Achievement[] = [
  { id: 1, icon: '/images/wood.png', iconHint: 'wood resource', title: 'Новичок-собиратель', description: 'Собрать 1,000 дерева', currentProgress: 1000, targetProgress: 1000, rewardIcon: '/images/gold.png', rewardAmount: 1000, claimed: true },
  { id: 2, icon: '/images/stone.png', iconHint: 'stone resource', title: 'Каменщик', description: 'Собрать 1,000 камня', currentProgress: 750, targetProgress: 1000, rewardIcon: '/images/gold.png', rewardAmount: 1000, claimed: false },
  { id: 3, icon: '/images/ore.png', iconHint: 'ore resource', title: 'Рудокоп', description: 'Собрать 500 руды', currentProgress: 500, targetProgress: 500, rewardIcon: '/images/gold.png', rewardAmount: 2000, claimed: false },
  { id: 4, icon: '/images/diamond.png', iconHint: 'diamond resource', title: 'Ювелир', description: 'Собрать 100 алмазов', currentProgress: 25, targetProgress: 100, rewardIcon: '/images/diamond.png', rewardAmount: 50, claimed: false },
  { id: 5, icon: 'https://placehold.co/64x64/ff4500/ffffff.png', iconHint: 'combo score', title: 'Мастер комбинаций', description: 'Сделать комбо 10x в игре', currentProgress: 10, targetProgress: 10, rewardIcon: '/images/diamond.png', rewardAmount: 100, claimed: false },
  { id: 6, icon: 'https://placehold.co/64x64/ff00ff/ffffff.png', iconHint: 'high combo score', title: 'Великий комбинатор', description: 'Сделать комбо 15x в игре', currentProgress: 8, targetProgress: 15, rewardIcon: '/images/diamond.png', rewardAmount: 250, claimed: false },
  { id: 7, icon: '/images/small-chess-res.png', iconHint: 'small chest', title: 'Охотник за сокровищами', description: 'Открыть 10 малых сундуков', currentProgress: 10, targetProgress: 10, rewardIcon: '/images/gold.png', rewardAmount: 5000, claimed: true },
  { id: 8, icon: '/images/medium-chess-res.png', iconHint: 'medium chest', title: 'Кладоискатель', description: 'Открыть 5 средних сундуков', currentProgress: 3, targetProgress: 5, rewardIcon: '/images/gold.png', rewardAmount: 10000, claimed: false },
  { id: 9, icon: '/images/big-chess-res.png', iconHint: 'large chest', title: 'Легендарный искатель', description: 'Открыть 1 большой сундук', currentProgress: 0, targetProgress: 1, rewardIcon: '/images/diamond.png', rewardAmount: 500, claimed: false },
  { id: 10, icon: '/images/menu-forge.png', iconHint: 'crafting hammer', title: 'Первые шаги в крафте', description: 'Создать 5 любых предметов в кузнице', currentProgress: 5, targetProgress: 5, rewardIcon: '/images/gold.png', rewardAmount: 2500, claimed: false },
  { id: 11, icon: '/images/block-stone.png', iconHint: 'crafted item', title: 'Ремесленник', description: 'Создать 25 любых предметов в кузнице', currentProgress: 15, targetProgress: 25, rewardIcon: '/images/gold.png', rewardAmount: 12500, claimed: false },
  { id: 12, icon: '/images/block-ore.png', iconHint: 'ingot', title: 'Мастер-кузнец', description: 'Создать 100 любых предметов в кузнице', currentProgress: 17, targetProgress: 100, rewardIcon: '/images/diamond.png', rewardAmount: 1000, claimed: false },
  { id: 13, icon: '/images/axie-wood.png', iconHint: 'wooden axe', title: 'Первый инструмент', description: 'Создать свой первый инструмент', currentProgress: 1, targetProgress: 1, rewardIcon: '/images/gold.png', rewardAmount: 5000, claimed: false },
  { id: 14, icon: '/images/blueprint-axie.png', iconHint: 'blueprint scroll', title: 'Полный набор', description: 'Собрать по одному чертежу каждого типа', currentProgress: 2, targetProgress: 4, rewardIcon: '/images/diamond.png', rewardAmount: 200, claimed: false },
  { id: 15, icon: '/images/menu-friend.png', iconHint: 'friends handshake', title: 'Душа компании', description: 'Пригласить 1 друга', currentProgress: 0, targetProgress: 1, rewardIcon: '/images/gold.png', rewardAmount: 25000, claimed: false },
];

const AchievementItem = ({ achievement, onClaim }: { achievement: Achievement; onClaim: (id: number) => void; }) => {
  const isCompleted = achievement.currentProgress >= achievement.targetProgress;
  
  return (
    <div className="flex items-center gap-3 p-3">
        <div className="flex-shrink-0 bg-background/50 rounded-lg p-1.5 border border-border/50">
            <Image src={achievement.icon} alt={achievement.title} width={48} height={48} data-ai-hint={achievement.iconHint} />
        </div>
        <div className="flex-grow space-y-1">
            <p className="font-semibold text-sm leading-tight text-foreground">{achievement.title}</p>
            <p className="text-xs text-muted-foreground">{achievement.description}</p>
        </div>
        <div className="flex flex-col items-center justify-center gap-1 w-24">
            <div className="flex items-center gap-1.5 bg-black/20 px-2 py-1 rounded-full">
                <Image src={achievement.rewardIcon} alt="reward" width={16} height={16} />
                <span className="text-xs font-bold text-yellow-300">{achievement.rewardAmount.toLocaleString()}</span>
            </div>
            <span className="text-xs text-muted-foreground">{achievement.currentProgress}/{achievement.targetProgress}</span>
        </div>
        <div className="flex-shrink-0 w-24">
            <Button 
                className="w-full" 
                disabled={!isCompleted || achievement.claimed}
                onClick={() => onClaim(achievement.id)}
            >
                {achievement.claimed ? (
                    <>
                        <Check className="mr-2 h-4 w-4" />
                        Получено
                    </>
                ) : isCompleted ? (
                    'Забрать'
                ) : (
                    `${Math.round((achievement.currentProgress / achievement.targetProgress) * 100)}%`
                )}
            </Button>
        </div>
    </div>
  );
}


export default function AchievementsPage() {
    const [achievements, setAchievements] = useState<Achievement[]>(initialAchievementsData);

    const handleClaim = (id: number) => {
        // Here you would typically call an API to claim the reward and update the state
        console.log(`Claiming achievement ${id}`);
        setAchievements(prev => prev.map(ach => 
            ach.id === id ? { ...ach, claimed: true } : ach
        ));
        // You would also add the rewards to the user's inventory here.
    };
  
    const totalAchievements = achievements.length;
    const completedAchievements = achievements.filter(a => a.currentProgress >= a.targetProgress).length;
    const overallProgress = totalAchievements > 0 ? (completedAchievements / totalAchievements) * 100 : 0;

    return (
        <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-4 text-foreground">
            <h1 className="text-3xl font-headline text-primary text-shadow-lg">Достижения</h1>

            <Card className="w-full max-w-2xl bg-card/80 backdrop-blur-md shadow-xl">
                <CardHeader>
                    <div className="flex items-center gap-4">
                        <div className="bg-primary/20 p-3 rounded-lg border border-primary/30">
                           <Award className="w-8 h-8 text-primary" />
                        </div>
                        <div>
                            <CardTitle>Прогресс достижений</CardTitle>
                            <CardDescription>Выполняйте задания, чтобы получить эксклюзивные награды.</CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-4">
                        <Progress value={overallProgress} className="h-3 flex-grow" />
                        <span className="font-bold text-primary">{Math.round(overallProgress)}%</span>
                    </div>
                </CardContent>
            </Card>

            <Card className="w-full max-w-2xl bg-card/80 backdrop-blur-md shadow-xl">
                <ScrollArea className="h-[60vh] w-full">
                    <CardContent className="p-0">
                        {achievements.map((ach, index) => (
                            <React.Fragment key={ach.id}>
                                <AchievementItem achievement={ach} onClaim={handleClaim} />
                                {index < achievements.length - 1 && <Separator className="bg-border/50" />}
                            </React.Fragment>
                        ))}
                    </CardContent>
                </ScrollArea>
            </Card>
        </div>
    );
}
