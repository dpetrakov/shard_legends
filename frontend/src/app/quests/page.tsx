
"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import Image from 'next/image';
import { Calendar, Award, Briefcase, Users } from 'lucide-react';

interface Quest {
  id: number;
  icon: string;
  iconHint: string;
  title: string;
  currentProgress: number;
  targetProgress: number;
  rewards: {
    gold: number;
    diamonds: number;
  };
  actionType: 'claim' | 'special';
  actionText?: string;
}

const dailyQuests: Quest[] = [
  { id: 1, icon: 'https://placehold.co/64x64/8a2be2/ffffff.png', iconHint: "matched crystals", title: 'Соберите 500 красных кристаллов', currentProgress: 335, targetProgress: 500, rewards: { gold: 5000, diamonds: 135 }, actionType: 'claim' },
  { id: 2, icon: 'https://placehold.co/64x64/7df9ff/000000.png', iconHint: "combo", title: 'Создайте 10 комбинаций "Комбо 4x"', currentProgress: 6, targetProgress: 10, rewards: { gold: 7500, diamonds: 150 }, actionType: 'claim' },
  { id: 3, icon: '/images/chess-blueprint.png', iconHint: "chest", title: 'Откройте 5 любых сундуков', currentProgress: 5, targetProgress: 5, rewards: { gold: 3000, diamonds: 75 }, actionType: 'claim' },
  { id: 4, icon: 'https://placehold.co/64x64/ffd700/000000.png', iconHint: "diamond", title: 'Ускорить процесс за бриллианты', currentProgress: 0, targetProgress: 1, rewards: { gold: 10000, diamonds: 0 }, actionType: 'special', actionText: 'Ускорить' },
  { id: 5, icon: '/images/menu-forge.png', iconHint: "crafting", title: 'Создайте 2 предмета в кузнице', currentProgress: 1, targetProgress: 2, rewards: { gold: 4000, diamonds: 100 }, actionType: 'claim' },
];

const QuestItem = ({ quest }: { quest: Quest }) => {
  const isCompleted = quest.currentProgress >= quest.targetProgress;
  
  return (
    <Card className="bg-card/70 border-border/50 p-3">
        <div className="flex items-center gap-3">
            <div className="flex-shrink-0 bg-background/50 rounded-lg p-2">
                <Image src={quest.icon} alt={quest.title} width={48} height={48} data-ai-hint={quest.iconHint} />
            </div>
            <div className="flex-grow space-y-2">
                <p className="font-semibold text-sm leading-tight text-foreground">{quest.title}</p>
                <div className="relative w-full h-5 bg-black/30 rounded-full overflow-hidden border border-border/30">
                    <Progress value={(quest.currentProgress / quest.targetProgress) * 100} className="h-full bg-green-500" />
                    <span className="absolute inset-0 flex items-center justify-center text-xs font-bold text-white text-shadow">{quest.currentProgress}/{quest.targetProgress}</span>
                </div>
                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-1">
                        <Image src="/images/gold.png" alt="Золото" width={16} height={16}/>
                        <span className="text-xs font-bold text-yellow-300">{quest.rewards.gold.toLocaleString()}</span>
                    </div>
                    <div className="flex items-center gap-1">
                        <Image src="/images/diamond.png" alt="Бриллианты" width={16} height={16}/>
                        <span className="text-xs font-bold text-cyan-300">{quest.rewards.diamonds.toLocaleString()}</span>
                    </div>
                </div>
            </div>
            <div className="flex-shrink-0">
                <Button 
                    className="w-20 h-14" 
                    disabled={!isCompleted}
                    variant={quest.actionType === 'special' ? 'secondary' : 'default'}
                >
                    {isCompleted ? (quest.actionType === 'claim' ? 'Забрать' : quest.actionText) : '...'}
                </Button>
            </div>
        </div>
    </Card>
  );
}


export default function QuestsPage() {
  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-4 text-foreground">
        <h1 className="text-3xl font-headline text-primary text-shadow-lg">Задания</h1>

        <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
            <CardContent className="p-2">
                <Tabs defaultValue="daily" className="w-full">
                    <TabsList className="grid w-full grid-cols-4">
                        <TabsTrigger value="daily"><Calendar className="w-4 h-4 mr-2" />Ежедневные</TabsTrigger>
                        <TabsTrigger value="seasonal"><Award className="w-4 h-4 mr-2" />Сезонные</TabsTrigger>
                        <TabsTrigger value="monthly"><Briefcase className="w-4 h-4 mr-2" />Месячные</TabsTrigger>
                        <TabsTrigger value="partner"><Users className="w-4 h-4 mr-2" />Партнерские</TabsTrigger>
                    </TabsList>
                    
                    <TabsContent value="daily" className="pt-4 px-2 space-y-3">
                        {dailyQuests.map(quest => (
                            <QuestItem key={quest.id} quest={quest} />
                        ))}
                        <div className="text-center text-muted-foreground text-sm pt-4">
                            Задания обновятся через: 9ч 58м
                        </div>
                    </TabsContent>
                    
                    <TabsContent value="seasonal">
                        <p className="text-center text-muted-foreground p-8">Сезонные задания появятся скоро.</p>
                    </TabsContent>

                    <TabsContent value="monthly">
                        <p className="text-center text-muted-foreground p-8">Месячные задания появятся скоро.</p>
                    </TabsContent>

                     <TabsContent value="partner">
                        <p className="text-center text-muted-foreground p-8">Партнерские задания появятся скоро.</p>
                    </TabsContent>
                </Tabs>
            </CardContent>
        </Card>
    </div>
  );
}
