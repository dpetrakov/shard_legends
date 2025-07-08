
"use client";

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogTrigger } from '@/components/ui/dialog';
import { Wrench, Plus, BarChart, ArrowRight, ShieldCheck, Zap, Clover, ChevronLeft, Mountain, Trees, Hourglass, Gift } from 'lucide-react';
import { useMining } from '@/contexts/MiningContext';
import { useInventory } from '@/contexts/InventoryContext';
import { ALL_TOOL_TYPES, type ToolType, type MiningLocation } from '@/types/mining';
import type { CraftedToolType, InventoryItemType } from '@/types/inventory';
import { AllCraftedToolTypes } from '@/types/inventory';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';

// Component for selecting a tool from inventory
const ToolSelectorView = ({ toolType, onBack }: { toolType: ToolType; onBack: () => void }) => {
    const { inventory, getItemName, getItemImage } = useInventory();
    const { equipTool, getToolName } = useMining();

    const availableTools = AllCraftedToolTypes.filter(item =>
        item.endsWith(`_${toolType}`) && (inventory[item] || 0) > 0
    );

    const handleSelect = (item: CraftedToolType) => {
        if (equipTool(toolType, item)) {
            onBack();
        } else {
            alert("Не удалось экипировать инструмент.");
        }
    };

    return (
        <div>
            <DialogHeader className="mb-4">
                <div className="flex items-center gap-2">
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onBack}>
                        <ChevronLeft className="h-5 w-5" />
                    </Button>
                    <div className="flex-grow">
                        <DialogTitle>Выбор: {getToolName(toolType)}</DialogTitle>
                        <DialogDescription>Выберите инструмент из вашего инвентаря.</DialogDescription>
                    </div>
                </div>
            </DialogHeader>
            <ScrollArea className="h-64">
                <div className="space-y-2 pr-4">
                    {availableTools.length > 0 ? availableTools.map((item, idx) => (
                        <React.Fragment key={item}>
                            <button
                                onClick={() => handleSelect(item)}
                                className="flex items-center justify-between w-full p-2 rounded-md hover:bg-muted transition-colors text-left"
                            >
                                <div className="flex items-center gap-3">
                                    <Image src={getItemImage(item) ?? `https://placehold.co/40x40.png`} alt={getItemName(item)} width={40} height={40} />
                                    <div>
                                        <p className="font-semibold">{getItemName(item)}</p>
                                        <p className="text-xs text-muted-foreground">В наличии: {inventory[item]}</p>
                                    </div>
                                </div>
                                <ArrowRight className="h-5 w-5 text-primary" />
                            </button>
                            {idx < availableTools.length - 1 && <Separator />}
                        </React.Fragment>
                    )) : (
                        <p className="text-center text-muted-foreground pt-10">Нет доступных инструментов этого типа.</p>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
};

// Component for the main tool management view
const ToolManagerView = ({ onSelectClick }: { onSelectClick: (toolType: ToolType) => void }) => {
    const { equippedTools, getToolName, getToolIcon, unequipTool } = useMining();
    const { getItemName, getItemImage } = useInventory();

    return (
        <div>
            <DialogHeader>
                <DialogTitle>Управление инструментами</DialogTitle>
                <DialogDescription>Выберите и установите инструменты для добычи.</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
                {ALL_TOOL_TYPES.map(toolType => {
                    const tool = equippedTools[toolType];
                    const Icon = getToolIcon(toolType);
                    const name = getToolName(toolType);

                    return (
                        <Card key={toolType} className="p-4 bg-card/80">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                    <div className="bg-muted p-2 rounded-md">
                                        <Icon className="h-6 w-6 text-muted-foreground" />
                                    </div>
                                    <h3 className="font-semibold">{name}</h3>
                                </div>
                                {tool && (
                                    <Button variant="destructive" size="sm" onClick={() => unequipTool(toolType)}>
                                        Снять
                                    </Button>
                                )}
                            </div>
                            {tool ? (
                                <CardContent className="pt-4 px-0 pb-0">
                                    <div className="flex items-center gap-2 mb-2">
                                        <Image src={getItemImage(tool.item) ?? `https://placehold.co/32x32.png`} alt={name} width={32} height={32} />
                                        <p className="text-sm font-medium">{getItemName(tool.item)}</p>
                                    </div>
                                    <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm text-muted-foreground">
                                        <div className="flex items-center gap-2"><BarChart className="w-4 h-4 text-primary" /> Сила: {tool.stats.strength}</div>
                                        <div className="flex items-center gap-2"><Zap className="w-4 h-4 text-primary" /> Скорость: {tool.stats.speed}</div>
                                        <div className="flex items-center gap-2"><Clover className="w-4 h-4 text-primary" /> Удача: {tool.stats.luck}</div>
                                        <div className="flex items-center gap-2"><ShieldCheck className="w-4 h-4 text-primary" /> Прочность: {tool.stats.durability}</div>
                                    </div>
                                </CardContent>
                            ) : (
                                <Button variant="outline" className="w-full mt-4" onClick={() => onSelectClick(toolType)}>
                                    <Plus className="mr-2 h-4 w-4" /> Выбрать
                                </Button>
                            )}
                        </Card>
                    );
                })}
            </div>
        </div>
    );
};


// The main dialog component that manages views
const ToolManagerDialog = () => {
    const [open, setOpen] = useState(false);
    const [view, setView] = useState<'main' | 'selector'>('main');
    const [selectingFor, setSelectingFor] = useState<ToolType | null>(null);

    const handleSelectClick = (toolType: ToolType) => {
        setSelectingFor(toolType);
        setView('selector');
    };
    
    const handleBack = () => {
        setView('main');
        setSelectingFor(null);
    };

    const handleOpenChange = (isOpen: boolean) => {
        setOpen(isOpen);
        if (!isOpen) {
            // Reset view state when closing the dialog
            handleBack();
        }
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
                <Button variant="outline" size="icon" className="bg-card/50 backdrop-blur-sm border-primary/50">
                    <Wrench className="h-6 w-6 text-primary" />
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
                {view === 'main' && <ToolManagerView onSelectClick={handleSelectClick} />}
                {view === 'selector' && selectingFor && <ToolSelectorView toolType={selectingFor} onBack={handleBack} />}
            </DialogContent>
        </Dialog>
    );
};

// Countdown Timer Component
const CountdownTimer = ({ endTime }: { endTime: number }) => {
    const [timeLeft, setTimeLeft] = useState(endTime - Date.now());

    useEffect(() => {
        if (endTime <= Date.now()) {
            setTimeLeft(0);
            return;
        }
        const timer = setInterval(() => {
            const newTimeLeft = endTime - Date.now();
            if (newTimeLeft <= 0) {
                setTimeLeft(0);
                clearInterval(timer);
            } else {
                setTimeLeft(newTimeLeft);
            }
        }, 1000);
        return () => clearInterval(timer);
    }, [endTime]);

    if (timeLeft <= 0) {
        return <span>Готово!</span>;
    }

    const hours = Math.floor(timeLeft / (1000 * 60 * 60)).toString().padStart(2, '0');
    const minutes = Math.floor((timeLeft % (1000 * 60 * 60)) / (1000 * 60)).toString().padStart(2, '0');
    const seconds = Math.floor((timeLeft % (1000 * 60)) / 1000).toString().padStart(2, '0');

    return <span>{`${hours}:${minutes}:${seconds}`}</span>;
};


export default function MinePage() {
    const { activeProcess, startMining, claimMining, equippedTools, getToolName } = useMining();
    const { getItemName } = useInventory();
    
    const [isLocationModalOpen, setIsLocationModalOpen] = useState(false);
    const [selectedLocation, setSelectedLocation] = useState<MiningLocation | null>(null);
    const [view, setView] = useState<'select' | 'confirm'>('select');

    const [isRewardModalOpen, setIsRewardModalOpen] = useState(false);
    const [lastReward, setLastReward] = useState<{ location: MiningLocation; rewards: any } | null>(null);

    const handleOpenLocationModal = () => {
        setView('select');
        setSelectedLocation(null);
        setIsLocationModalOpen(true);
    };

    const handleSelectLocation = (location: MiningLocation) => {
        const requiredTool: ToolType = location === 'cave' ? 'pickaxe' : 'axe';
        if (!equippedTools[requiredTool]) {
            alert(`Для этой локации требуется экипированный ${getToolName(requiredTool).toLowerCase()}.`);
            return;
        }
        setSelectedLocation(location);
        setView('confirm');
    };

    const handleStartMining = () => {
        if (!selectedLocation) return;
        if (startMining(selectedLocation)) {
            setIsLocationModalOpen(false);
        }
    };

    const handleClaim = () => {
        const result = claimMining();
        if (result) {
            setLastReward(result);
            setIsRewardModalOpen(true);
        }
    };
    
    const isMiningFinished = activeProcess && Date.now() >= activeProcess.endTime;

    return (
        <div
            className="absolute inset-0 flex flex-col justify-between px-4 pt-20 pb-28 bg-cover bg-bottom bg-no-repeat"
            style={{ backgroundImage: "url('/images/mine-background.jpg')" }}
        >
            <div className="w-full flex justify-end gap-2">
                <Link href="/gifts">
                    <Button variant="outline" size="icon" className="bg-card/50 backdrop-blur-sm border-primary/50">
                        <Gift className="h-6 w-6 text-primary" />
                    </Button>
                </Link>
                <ToolManagerDialog />
            </div>

            <div className="w-full flex flex-col items-center">
                {activeProcess ? (
                    <Card className="bg-card/80 backdrop-blur-md shadow-xl text-center p-6 w-full max-w-sm">
                        <CardHeader className="p-0 mb-4">
                            <CardTitle className="text-2xl font-headline text-primary">Добыча в процессе</CardTitle>
                            <CardDescription>Локация: {activeProcess.location === 'cave' ? 'Пещера' : 'Лес'}</CardDescription>
                        </CardHeader>
                        <CardContent className="p-0">
                            {isMiningFinished ? (
                                <Button size="lg" onClick={handleClaim}>Забрать награду</Button>
                            ) : (
                                <div className="flex flex-col items-center gap-2">
                                    <Hourglass className="w-12 h-12 text-primary animate-spin" />
                                    <p className="text-2xl font-bold font-mono tabular-nums">
                                        <CountdownTimer endTime={activeProcess.endTime} />
                                    </p>
                                    <p className="text-sm text-muted-foreground">Осталось времени</p>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                ) : (
                    <button onClick={handleOpenLocationModal} className="flex flex-col items-center text-center cursor-pointer hover:scale-105 transition-transform group">
                        <div className="bg-gradient-to-t from-primary/80 to-primary/40 rounded-full shadow-xl shadow-primary/50 flex items-center justify-center p-4">
                            <Image
                                src="/images/mine-play.png"
                                alt="Начать добычу"
                                width={180}
                                height={180}
                                data-ai-hint="mining play"
                            />
                        </div>
                        <span className="text-2xl font-headline text-white text-shadow-lg mt-2 group-hover:text-primary transition-colors">
                            Начать добычу
                        </span>
                    </button>
                )}
            </div>
            
            <Dialog open={isLocationModalOpen} onOpenChange={setIsLocationModalOpen}>
                <DialogContent className="sm:max-w-md">
                    {view === 'select' && (
                        <>
                          <DialogHeader>
                                <DialogTitle>Выбор локации</DialogTitle>
                                <DialogDescription>Выберите, где вы хотите провести добычу.</DialogDescription>
                            </DialogHeader>
                            <div className="grid grid-cols-1 gap-4 py-4">
                                <Card onClick={() => handleSelectLocation('cave')} className="p-4 hover:bg-muted cursor-pointer transition-colors">
                                    <div className="flex items-center gap-4">
                                        <Mountain className="w-10 h-10 text-primary flex-shrink-0" />
                                        <div>
                                            <h3 className="font-bold text-lg">Пещера</h3>
                                            <p className="text-sm text-muted-foreground">Добыча камня. Шанс найти руду, золото и $SLCW. Требуется: Кирка</p>
                                        </div>
                                    </div>
                                </Card>
                                <Card onClick={() => handleSelectLocation('forest')} className="p-4 hover:bg-muted cursor-pointer transition-colors">
                                    <div className="flex items-center gap-4">
                                        <Trees className="w-10 h-10 text-primary flex-shrink-0" />
                                        <div>
                                            <h3 className="font-bold text-lg">Лес</h3>
                                            <p className="text-sm text-muted-foreground">Добыча дерева. Шанс найти золото, алмазы и $SLCW. Требуется: Топор</p>
                                        </div>
                                    </div>
                                </Card>
                            </div>
                        </>
                    )}
                    {view === 'confirm' && selectedLocation && (
                         <>
                            <DialogHeader>
                                <div className="flex items-center gap-2">
                                     <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setView('select')}>
                                        <ChevronLeft className="h-5 w-5" />
                                    </Button>
                                    <div>
                                        <DialogTitle>Подтверждение: {selectedLocation === 'cave' ? 'Пещера' : 'Лес'}</DialogTitle>
                                        <DialogDescription>Проверьте детали и начните добычу.</DialogDescription>
                                    </div>
                                </div>
                            </DialogHeader>
                            <Card className="p-4 bg-card/80">
                                <CardContent className="pt-4 px-0 pb-0 space-y-3">
                                    <p className="text-sm font-semibold">Активный инструмент: <span className="text-primary font-normal">{getItemName(equippedTools[selectedLocation === 'cave' ? 'pickaxe' : 'axe']!.item)}</span></p>
                                    <p className="text-sm font-semibold">Предварительные результаты:</p>
                                    <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                                       <p>Скорость: <span className="font-bold text-foreground">10/час</span></p>
                                       <p>Сила: <span className="font-bold text-foreground">10/тик</span></p>
                                       <p>Шанс награды: <span className="font-bold text-foreground">10%</span></p>
                                       <p>Прочность: <span className="font-bold text-destructive">-1</span></p>
                                    </div>
                                    <Button className="w-full mt-4" onClick={handleStartMining}>Запустить 1 час</Button>
                                </CardContent>
                            </Card>
                         </>
                    )}
                </DialogContent>
            </Dialog>

            <Dialog open={isRewardModalOpen} onOpenChange={setIsRewardModalOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Награда за добычу!</DialogTitle>
                        <DialogDescription>Вы успешно завершили добычу в локации "{lastReward?.location === 'cave' ? 'Пещера' : 'Лес'}"</DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <p className="font-semibold">Вы получили:</p>
                        <ul className="list-disc list-inside">
                            {lastReward && Object.entries(lastReward.rewards).map(([key, value]) => (
                                <li key={key}>{getItemName(key as InventoryItemType)}: {value as number}</li>
                            ))}
                        </ul>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}

    

    
