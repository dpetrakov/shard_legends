
"use client";

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";

import { REFINING_RECIPES } from '@/lib/refining-definitions';
import { TOOL_TYPES, TOOL_QUALITIES, CRAFTING_COST_PER_ITEM, CRAFTING_DURATION_SECONDS_PER_ITEM, getCraftedToolId } from '@/lib/crafting-definitions';
import type { RefiningRecipe } from '@/types/refining';
import type { BlueprintType, CraftedToolType } from '@/types/inventory';
import { useInventory } from '@/contexts/InventoryContext';
import { useRefining } from '@/contexts/RefiningContext';
import { useCrafting } from '@/contexts/CraftingContext';
import { useMining } from '@/contexts/MiningContext';

const craftedToolImageMap: Record<CraftedToolType, string> = {
    wooden_axe: '/images/axie-wood.png',
    stone_axe: '/images/axie-stone.png',
    metal_axe: '/images/axie-metal.png',
    diamond_axe: '/images/axie-almaz.png',
    wooden_pickaxe: '/images/pickaxie-wood.png',
    stone_pickaxe: '/images/pickaxie-stone.png',
    metal_pickaxe: '/images/pickaxie-metal.png',
    diamond_pickaxe: '/images/pickaxie-almaz.png',
    wooden_shovel: '/images/shovel-wood.png',
    stone_shovel: '/images/shovel-stone.png',
    metal_shovel: '/images/shovel-metal.png',
    diamond_shovel: '/images/shovel-almaz.png',
    wooden_sickle: '/images/sickle-wood.png',
    stone_sickle: '/images/sickle-stone.png',
    metal_sickle: '/images/sickle-metal.png',
    diamond_sickle: '/images/sickle-almaz.png',
};


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

function RefiningTab() {
    const { inventory, getItemName } = useInventory();
    const { activeProcesses, startProcess, claimProcess, processSlots } = useRefining();
    const [quantities, setQuantities] = useState<Record<string, number>>({});

    const handleQuantityChange = (recipeOutput: string, value: string) => {
        const numValue = parseInt(value, 10);
        setQuantities(prev => ({
            ...prev,
            [recipeOutput]: isNaN(numValue) || numValue < 1 ? 1 : numValue,
        }));
    };

    const handleStartProcess = (recipe: RefiningRecipe) => {
        const quantity = quantities[recipe.output.item] || 1;
        startProcess(recipe, quantity);
    };

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardHeader>
                <div className="flex justify-between items-center">
                    <CardTitle className="text-xl font-headline text-accent">Переработка</CardTitle>
                    <div className="text-sm text-muted-foreground">
                        Слотов: {processSlots.max - processSlots.current}/{processSlots.max}
                    </div>
                </div>
            </CardHeader>
            <CardContent className="pt-4 space-y-6">
                {activeProcesses.length > 0 && (
                    <div className="space-y-4">
                        <h4 className="font-headline text-lg text-primary">Активные процессы</h4>
                        {activeProcesses.map(proc => {
                             const isFinished = Date.now() >= proc.endTime;
                             const totalDuration = proc.recipe.durationSeconds * proc.quantity * 1000;
                             const timeElapsed = Date.now() - (proc.endTime - totalDuration);
                             const progress = isFinished ? 100 : Math.min(100, (timeElapsed / totalDuration) * 100);

                            return (
                            <Card key={proc.id} className="bg-card/70">
                                <CardHeader className="p-4">
                                    <CardTitle className="text-base">{getItemName(proc.recipe.output.item)} x{proc.quantity}</CardTitle>
                                </CardHeader>
                                <CardContent className="p-4 pt-0">
                                    <Progress value={progress} className="h-2 mb-2" />
                                    <div className="flex justify-between items-center text-xs text-muted-foreground">
                                        <CountdownTimer endTime={proc.endTime} />
                                        <span>{isFinished ? '100%' : `${Math.floor(progress)}%`}</span>
                                    </div>
                                </CardContent>
                                <CardFooter className="p-4 pt-0">
                                     <Button onClick={() => claimProcess(proc.id)} disabled={!isFinished} className="w-full">
                                        Забрать
                                    </Button>
                                </CardFooter>
                            </Card>
                        )})}
                    </div>
                )}
                <div className="space-y-4">
                    <h4 className="font-headline text-lg text-primary">Рецепты</h4>
                    {REFINING_RECIPES.map(recipe => {
                        const quantity = quantities[recipe.output.item] || 1;
                        const canAfford = recipe.input.every(
                            (ing) => (inventory[ing.item] || 0) >= ing.quantity * quantity
                        );
                        return (
                            <Card key={recipe.output.item} className="bg-card/70">
                                <CardHeader className="p-4">
                                    <CardTitle className="text-base">
                                        {getItemName(recipe.output.item)}
                                    </CardTitle>
                                    <CardDescription className="text-xs">
                                        Время: {recipe.durationSeconds * quantity / 60} мин.
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="p-4 pt-0 text-sm space-y-2">
                                    <p className="font-semibold">Требуется:</p>
                                    <ul className="list-disc list-inside text-muted-foreground">
                                        {recipe.input.map(ing => (
                                            <li key={ing.item} className={ (inventory[ing.item] || 0) < ing.quantity * quantity ? 'text-destructive' : ''}>
                                                {getItemName(ing.item)}: {ing.quantity * quantity} (у вас: {inventory[ing.item] || 0})
                                            </li>
                                        ))}
                                    </ul>
                                </CardContent>
                                <CardFooter className="p-4 pt-0 flex gap-2">
                                    <Input
                                        type="number"
                                        min="1"
                                        value={quantity}
                                        onChange={(e) => handleQuantityChange(recipe.output.item, e.target.value)}
                                        className="w-20"
                                    />
                                    <Button
                                        onClick={() => handleStartProcess(recipe)}
                                        disabled={!canAfford || processSlots.current >= processSlots.max}
                                        className="flex-grow"
                                    >
                                        Переплавить
                                    </Button>
                                </CardFooter>
                            </Card>
                        );
                    })}
                </div>
            </CardContent>
        </Card>
    );
}

function CraftingTab() {
    const { inventory, getItemName } = useInventory();
    const { activeProcesses, startProcess, claimProcess, processSlots } = useCrafting();

    const [selectedTool, setSelectedTool] = useState<BlueprintType | undefined>();
    const [selectedQuality, setSelectedQuality] = useState<string | undefined>();
    const [quantity, setQuantity] = useState<number>(1);

    const ownedBlueprints = TOOL_TYPES.filter(bp => (inventory[bp] || 0) > 0);

    useEffect(() => {
        if (ownedBlueprints.length > 0 && !selectedTool) {
            setSelectedTool(ownedBlueprints[0]);
        }
        if (TOOL_QUALITIES.length > 0 && !selectedQuality) {
            setSelectedQuality(TOOL_QUALITIES[0].id);
        }
    }, [ownedBlueprints, selectedTool, selectedQuality]);
    
    const qualityInfo = TOOL_QUALITIES.find(q => q.id === selectedQuality);
    const materialNeeded = qualityInfo ? qualityInfo.material : null;
    const materialCost = materialNeeded ? CRAFTING_COST_PER_ITEM * quantity : 0;
    const blueprintCost = quantity;

    const canAfford = selectedTool && qualityInfo && materialNeeded &&
                      (inventory[selectedTool] || 0) >= blueprintCost &&
                      (inventory[materialNeeded] || 0) >= materialCost;

    const handleStartCrafting = () => {
        if (!selectedTool || !selectedQuality) return;
        startProcess({ tool: selectedTool, quality: selectedQuality }, quantity);
    };

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardHeader>
                <div className="flex justify-between items-center">
                    <CardTitle className="text-xl font-headline text-accent">Крафт</CardTitle>
                    <div className="text-sm text-muted-foreground">
                        Слотов: {processSlots.max - processSlots.current}/{processSlots.max}
                    </div>
                </div>
            </CardHeader>
            <CardContent className="pt-4 space-y-6">
                {activeProcesses.length > 0 && (
                    <div className="space-y-4">
                        <h4 className="font-headline text-lg text-primary">Активные процессы</h4>
                        {activeProcesses.map(proc => {
                            const isFinished = Date.now() >= proc.endTime;
                            const totalDuration = CRAFTING_DURATION_SECONDS_PER_ITEM * proc.quantity * 1000;
                            const timeElapsed = Date.now() - (proc.endTime - totalDuration);
                            const progress = isFinished ? 100 : Math.min(100, (timeElapsed / totalDuration) * 100);

                            return (
                            <Card key={proc.id} className="bg-card/70">
                                <CardHeader className="p-4">
                                    <CardTitle className="text-base">{proc.outputItemName} x{proc.quantity}</CardTitle>
                                </CardHeader>
                                <CardContent className="p-4 pt-0">
                                    <Progress value={progress} className="h-2 mb-2" />
                                    <div className="flex justify-between items-center text-xs text-muted-foreground">
                                        <CountdownTimer endTime={proc.endTime} />
                                        <span>{isFinished ? '100%' : `${Math.floor(progress)}%`}</span>
                                    </div>
                                </CardContent>
                                <CardFooter className="p-4 pt-0">
                                     <Button onClick={() => claimProcess(proc.id)} disabled={!isFinished} className="w-full">
                                        Забрать
                                    </Button>
                                </CardFooter>
                            </Card>
                        )})}
                    </div>
                )}
                <div className="space-y-4">
                    <h4 className="font-headline text-lg text-primary">Рецепты</h4>
                    {ownedBlueprints.length > 0 ? (
                        <Card className="bg-card/70 p-4 space-y-4">
                             <div className="space-y-2">
                                <label className="text-sm font-medium">Инструмент (нужен чертеж)</label>
                                <Select value={selectedTool} onValueChange={(val: BlueprintType) => setSelectedTool(val)}>
                                    <SelectTrigger><SelectValue placeholder="Выберите инструмент..." /></SelectTrigger>
                                    <SelectContent>
                                        {ownedBlueprints.map(bp => (
                                            <SelectItem key={bp} value={bp}>{getItemName(bp)}</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">Качество</label>
                                <Select value={selectedQuality} onValueChange={(val: string) => setSelectedQuality(val)}>
                                    <SelectTrigger><SelectValue placeholder="Выберите качество..." /></SelectTrigger>
                                    <SelectContent>
                                        {TOOL_QUALITIES.map(q => (
                                            <SelectItem key={q.id} value={q.id}>{q.name}</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            
                            {selectedTool && qualityInfo && materialNeeded && (
                                <CardContent className="p-0 text-sm space-y-2">
                                    <p className="font-semibold">Требуется на {quantity} шт.:</p>
                                    <ul className="list-disc list-inside text-muted-foreground">
                                        <li className={(inventory[selectedTool] || 0) < blueprintCost ? 'text-destructive' : ''}>
                                            {getItemName(selectedTool)}: {blueprintCost} (у вас: {inventory[selectedTool] || 0})
                                        </li>
                                        <li className={(inventory[materialNeeded] || 0) < materialCost ? 'text-destructive' : ''}>
                                            {getItemName(materialNeeded)}: {materialCost} (у вас: {inventory[materialNeeded] || 0})
                                        </li>
                                    </ul>
                                    <p>Время: {CRAFTING_DURATION_SECONDS_PER_ITEM * quantity / 60} мин.</p>
                                </CardContent>
                            )}

                            <CardFooter className="p-0 pt-2 flex gap-2">
                                <Input
                                    type="number"
                                    min="1"
                                    value={quantity}
                                    onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value, 10) || 1))}
                                    className="w-20"
                                />
                                <Button
                                    onClick={handleStartCrafting}
                                    disabled={!canAfford || processSlots.current >= processSlots.max}
                                    className="flex-grow"
                                >
                                    Создать
                                </Button>
                            </CardFooter>
                        </Card>
                    ) : (
                        <p className="text-center text-muted-foreground py-4">У вас нет чертежей для создания инструментов.</p>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}

function RepairingTab() {
    const { equippedTools } = useMining();
    const { getItemName } = useInventory();

    const equippedToolsArray = Object.values(equippedTools).filter(Boolean);

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardHeader>
                <CardTitle className="text-xl font-headline text-accent">Починка</CardTitle>
                <CardDescription>Восстановите прочность ваших инструментов.</CardDescription>
            </CardHeader>
            <CardContent className="pt-4 space-y-4">
                {equippedToolsArray.length > 0 ? (
                    equippedToolsArray.map(tool => (
                        <Card key={tool.item} className="bg-card/70 p-4">
                            <div className="flex items-center gap-4">
                                <Image
                                    src={craftedToolImageMap[tool.item]}
                                    alt={getItemName(tool.item)}
                                    width={48}
                                    height={48}
                                />
                                <div className="flex-grow">
                                    <p className="font-semibold">{getItemName(tool.item)}</p>
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <span>Прочность:</span>
                                        <Progress value={tool.stats.durability} className="w-24 h-2" />
                                        <span>{tool.stats.durability}/100</span>
                                    </div>
                                </div>
                            </div>
                            <div className="mt-4 space-y-2">
                                <p className="text-sm font-semibold">Стоимость починки:</p>
                                <div className="flex items-center gap-4 text-sm">
                                    <div className="flex items-center gap-1.5">
                                        <Image src="/images/gold.png" alt="Золото" width={16} height={16} />
                                        <span>1,000</span>
                                    </div>
                                    <div className="flex items-center gap-1.5">
                                        <Image src="/images/block-ore.png" alt={getItemName('metal_ingot')} width={16} height={16} />
                                        <span>5</span>
                                    </div>
                                </div>
                            </div>
                            <CardFooter className="p-0 pt-4">
                                <Button className="w-full" disabled={tool.stats.durability >= 100}>
                                    Починить
                                </Button>
                            </CardFooter>
                        </Card>
                    ))
                ) : (
                    <p className="text-center text-muted-foreground py-10">Нет экипированных инструментов для починки.</p>
                )}
            </CardContent>
        </Card>
    );
}

const SharpeningTab = () => {
    const { getItemName, inventory } = useInventory();
    const [toolType, setToolType] = useState<BlueprintType>('pickaxe');
    const [quality, setQuality] = useState<string>('wooden');
    const [selectedLevel, setSelectedLevel] = useState<number | null>(null);
    const [quantity, setQuantity] = useState(1);

    const toolTypeDisplayNames: Record<BlueprintType, string> = {
        axe: 'Топор',
        pickaxe: 'Кирка',
        shovel: 'Лопата',
        sickle: 'Серп',
    };

    const levelColorClasses: Record<number, string> = {
        1: 'bg-green-300/20 border-green-400/50 text-green-200',
        2: 'bg-green-600/30 border-green-500/70 text-green-300',
        3: 'bg-blue-300/20 border-blue-400/50 text-blue-200',
        4: 'bg-blue-600/30 border-blue-500/70 text-blue-300',
        5: 'bg-purple-300/20 border-purple-400/50 text-purple-200',
        6: 'bg-purple-600/30 border-purple-500/70 text-purple-300',
        7: 'bg-red-300/20 border-red-400/50 text-red-200',
        8: 'bg-red-600/30 border-red-500/70 text-red-300',
        9: 'bg-yellow-300/20 border-yellow-400/50 text-yellow-200',
        10: 'bg-yellow-500/30 border-yellow-500/70 text-yellow-300 font-bold',
    };

    const getSharpenedItemName = (level: number) => {
        const baseToolId = getCraftedToolId(toolType, quality);
        const baseToolName = getItemName(baseToolId);
        if (level > 0) {
            return `+${level} ${baseToolName}`;
        }
        return baseToolName;
    };

    const selectedToolImage = craftedToolImageMap[getCraftedToolId(toolType, quality)];

    const getRequiredItemInfo = () => {
        if (!selectedLevel) return null;

        if (selectedLevel === 1) {
            const baseToolId = getCraftedToolId(toolType, quality);
            return {
                name: getItemName(baseToolId),
                countInInventory: inventory[baseToolId] || 0
            };
        } else {
            return {
                name: getSharpenedItemName(selectedLevel - 1),
                // Inventory system doesn't support sharpened items yet.
                countInInventory: 0
            };
        }
    };
    
    const requiredItemInfo = getRequiredItemInfo();

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardHeader>
                <CardTitle className="text-xl font-headline text-accent">Заточка</CardTitle>
                <CardDescription>Улучшайте инструменты, чтобы повысить их характеристики.</CardDescription>
            </CardHeader>
            <CardContent className="pt-4 space-y-6">
                <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Категория</label>
                        <Select value={toolType} onValueChange={(val: BlueprintType) => setToolType(val)}>
                            <SelectTrigger><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {TOOL_TYPES.map(t => <SelectItem key={t} value={t}>{toolTypeDisplayNames[t]}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Качество</label>
                        <Select value={quality} onValueChange={setQuality}>
                            <SelectTrigger><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {TOOL_QUALITIES.map(q => <SelectItem key={q.id} value={q.id}>{q.name}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div className="grid grid-cols-3 sm:grid-cols-5 gap-3">
                    {Array.from({ length: 10 }, (_, i) => i + 1).map(level => (
                        <button
                            key={level}
                            className={cn(
                                "h-auto p-2 flex flex-col items-center justify-center transition-all rounded-lg border aspect-square",
                                levelColorClasses[level],
                                selectedLevel === level ? 'ring-2 ring-primary' : 'hover:ring-1 hover:ring-primary/50'
                            )}
                            onClick={() => setSelectedLevel(level)}
                        >
                            {selectedToolImage && (
                                <Image src={selectedToolImage} alt={getSharpenedItemName(level)} width={48} height={48} className="flex-shrink-0" />
                            )}
                            <span className="text-sm font-bold mt-1">+{level}</span>
                        </button>
                    ))}
                </div>

                {selectedLevel && requiredItemInfo && (
                    <Card className="bg-card/70 p-4">
                        <CardTitle className="text-base mb-2">
                            Создать: {getSharpenedItemName(selectedLevel)}
                        </CardTitle>
                        <CardContent className="p-0 text-sm space-y-2">
                            <p className="font-semibold">Требуется:</p>
                            <ul className="list-disc list-inside text-muted-foreground">
                                <li className={requiredItemInfo.countInInventory < 2 * quantity ? 'text-destructive' : ''}>
                                    {requiredItemInfo.name}: {2 * quantity} (у вас: {requiredItemInfo.countInInventory})
                                </li>
                                <li>Золото: {(selectedLevel * 1000 * quantity).toLocaleString()} (у вас: 0)</li>
                            </ul>
                        </CardContent>
                        <CardFooter className="p-0 pt-4 flex gap-2">
                            <Input
                                type="number"
                                min="1"
                                value={quantity}
                                onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value, 10) || 1))}
                                className="w-20"
                            />
                            <Button
                                disabled // Заточка будет реализована на следующем шаге
                                className="flex-grow"
                            >
                                Заточить
                            </Button>
                        </CardFooter>
                    </Card>
                )}
            </CardContent>
        </Card>
    );
};

export default function FactoryPage() {
  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-3xl font-headline text-center text-primary">Кузница</CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <Tabs defaultValue="refining" className="w-full">
            <TabsList className="grid w-full grid-cols-4 mb-4 text-xs sm:text-sm">
              <TabsTrigger value="refining">Переработка</TabsTrigger>
              <TabsTrigger value="crafting">Крафт</TabsTrigger>
              <TabsTrigger value="sharpening">Заточка</TabsTrigger>
              <TabsTrigger value="repairing">Починка</TabsTrigger>
            </TabsList>
            <TabsContent value="refining">
                <RefiningTab />
            </TabsContent>
            <TabsContent value="crafting">
                <CraftingTab />
            </TabsContent>
            <TabsContent value="sharpening">
                <SharpeningTab />
            </TabsContent>
            <TabsContent value="repairing">
              <RepairingTab />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
