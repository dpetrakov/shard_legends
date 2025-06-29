
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
import { FlaskConical, Anvil, Sparkles, Wrench, ArrowRight, Plus } from 'lucide-react';

import { REFINING_RECIPES } from '@/lib/refining-definitions';
import { TOOL_TYPES, TOOL_QUALITIES, CRAFTING_COST_PER_ITEM, CRAFTING_DURATION_SECONDS_PER_ITEM, getCraftedToolId } from '@/lib/crafting-definitions';
import type { RefiningRecipe } from '@/types/refining';
import type { BlueprintType, CraftedToolType, InventoryItemType, ResourceType, ReagentType, ProcessedItemType } from '@/types/inventory';
import { useInventory } from '@/contexts/InventoryContext';
import { useRefining } from '@/contexts/RefiningContext';
import { useCrafting } from '@/contexts/CraftingContext';
import { useMining } from '@/contexts/MiningContext';

const resourceImageMap: Record<ResourceType, string> = {
    stone: '/images/stone.png',
    wood: '/images/wood.png',
    ore: '/images/ore.png',
    diamond: '/images/almaz.png',
};

const reagentImageMap: Record<ReagentType, string> = {
    abrasive: '/images/ing-abraziv.png',
    disc: '/images/ing-disk.png',
    inductor: '/images/ing-ore.png',
    paste: '/images/ing-pasta.png',
};

const processedItemImageMap: Record<ProcessedItemType, string> = {
    wood_plank: '/images/block-wood.png',
    stone_block: '/images/block-stone.png',
    metal_ingot: '/images/block-ore.png',
    cut_diamond: '/images/block-almaz.png',
};

const blueprintImageMap: Record<BlueprintType, string> = {
    axe: '/images/blueprint-axie.png',
    pickaxe: '/images/blueprint-pickaxie.png',
    shovel: '/images/blueprint-shovel.png',
    sickle: '/images/blueprint-sickle.png',
};

const itemImageMap: Record<string, string> = {
    ...resourceImageMap,
    ...reagentImageMap,
    ...processedItemImageMap,
    ...blueprintImageMap,
};

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

function ActiveProcessesDisplay() {
    const { activeProcesses: refiningProcs, claimProcess: claimRefining } = useRefining();
    const { activeProcesses: craftingProcs, claimProcess: claimCrafting } = useCrafting();
    const { getItemName } = useInventory();

    const allProcs = [
        ...refiningProcs.map(p => ({
            ...p,
            type: 'refining' as const,
            name: getItemName(p.recipe.output.item),
            image: itemImageMap[p.recipe.output.item] as string,
            claimFn: () => claimRefining(p.id)
        })),
        ...craftingProcs.map(p => ({
            ...p,
            type: 'crafting' as const,
            name: p.outputItemName,
            image: craftedToolImageMap[p.outputItem as CraftedToolType],
            claimFn: () => claimCrafting(p.id)
        }))
    ];
    
    const MAX_FACTORY_SLOTS = 2;
    const totalSlotsUsed = allProcs.length;

    if (totalSlotsUsed === 0) {
        return <div className="text-center text-muted-foreground py-2">Нет активных процессов</div>;
    }

    return (
        <div className="mb-4 space-y-3 pt-4">
            <div className="flex justify-between items-center px-2">
                <h4 className="font-headline text-lg text-primary">Активные процессы</h4>
                <div className="text-sm text-muted-foreground">
                    Слотов занято: {totalSlotsUsed} / {MAX_FACTORY_SLOTS}
                </div>
            </div>
            {allProcs.map(proc => {
                const isFinished = Date.now() >= proc.endTime;
                let totalDuration: number;
                if (proc.type === 'refining') {
                    totalDuration = proc.recipe.durationSeconds * proc.quantity * 1000;
                } else { // crafting
                    totalDuration = CRAFTING_DURATION_SECONDS_PER_ITEM * proc.quantity * 1000;
                }
                const timeElapsed = Date.now() - (proc.endTime - totalDuration);
                const progress = isFinished ? 100 : Math.min(100, (timeElapsed / totalDuration) * 100);

                return (
                    <Card key={proc.id} className="bg-card/70 p-3">
                        <div className="flex items-center gap-3">
                            <div className="relative flex-shrink-0">
                                <Image 
                                    src={proc.image} 
                                    alt={proc.name} 
                                    width={48} 
                                    height={48} 
                                    className="bg-background/50 rounded-md p-1"
                                />
                                <span className="absolute -top-1 -right-1 bg-primary text-primary-foreground text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center border-2 border-card">
                                    {proc.quantity}
                                </span>
                            </div>

                            <div className="flex-grow space-y-1">
                                <p className="font-semibold text-sm leading-tight">{proc.name}</p>
                                <Progress value={progress} className="h-2" />
                                <div className="flex justify-between items-center text-xs text-muted-foreground">
                                    <CountdownTimer endTime={proc.endTime} />
                                    <span>{isFinished ? '100%' : `${Math.floor(progress)}%`}</span>
                                </div>
                            </div>
                            
                            <div className="flex-shrink-0">
                                <Button onClick={proc.claimFn} disabled={!isFinished} size="sm" className="px-4">
                                    Забрать
                                </Button>
                            </div>
                        </div>
                    </Card>
                );
            })}
        </div>
    );
}

function RefiningTab() {
    const { inventory, getItemName } = useInventory();
    const { startProcess, processSlots } = useRefining();
    const { processSlots: craftingSlots } = useCrafting();
    const [selectedRecipe, setSelectedRecipe] = useState<RefiningRecipe | null>(REFINING_RECIPES[0] || null);
    const [quantity, setQuantity] = useState<number>(1);
    
    const totalSlotsUsed = processSlots.current + craftingSlots.current;
    const MAX_FACTORY_SLOTS = 2;

    const handleRecipeSelect = (recipe: RefiningRecipe) => {
        setSelectedRecipe(recipe);
        setQuantity(1);
    };
    
    const handleStartProcess = () => {
        if (!selectedRecipe) return;
        startProcess(selectedRecipe, quantity);
        setQuantity(1);
    };

    const canAfford = selectedRecipe && selectedRecipe.input.every(
        (ing) => (inventory[ing.item] || 0) >= ing.quantity * quantity
    );

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardContent className="pt-6 space-y-6">
                 {selectedRecipe ? (
                    <Card className="bg-card/70 p-4">
                        <div className="flex items-center justify-center gap-1 sm:gap-2 mb-4 text-primary font-bold">
                            {selectedRecipe.input.map((ing, index) => (
                                <React.Fragment key={ing.item}>
                                    {index > 0 && <Plus className="h-5 w-5 text-muted-foreground" />}
                                    <div className="flex flex-col items-center w-16 text-center">
                                        <Image src={itemImageMap[ing.item] as string} alt={getItemName(ing.item)} width={40} height={40} />
                                        <span className={cn("text-xs mt-1", (inventory[ing.item] || 0) < ing.quantity * quantity ? 'text-destructive' : 'text-muted-foreground')}>
                                            {ing.quantity * quantity}/{inventory[ing.item] || 0}
                                        </span>
                                    </div>
                                </React.Fragment>
                            ))}
                            <ArrowRight className="h-6 w-6 mx-1 sm:mx-2 text-primary" />
                            <div className="flex flex-col items-center w-16 text-center">
                                <Image src={itemImageMap[selectedRecipe.output.item] as string} alt={getItemName(selectedRecipe.output.item)} width={40} height={40} />
                                <span className="text-xs mt-1 text-muted-foreground">
                                    {selectedRecipe.output.quantity * quantity}
                                </span>
                            </div>
                        </div>
                        <div className="flex gap-2">
                            <Input
                                type="number"
                                min="1"
                                value={quantity}
                                onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value, 10) || 1))}
                                className="w-24"
                            />
                            <Button
                                onClick={handleStartProcess}
                                disabled={!canAfford || totalSlotsUsed >= MAX_FACTORY_SLOTS}
                                className="flex-grow"
                            >
                                Переплавить
                            </Button>
                        </div>
                        <p className="text-xs text-muted-foreground text-center mt-2">Время: {selectedRecipe.durationSeconds * quantity} сек.</p>
                    </Card>
                ) : <p className="text-center text-muted-foreground">Выберите предмет для переработки</p>}

                <div>
                    <h4 className="font-headline text-lg text-primary mb-2 text-center">Выберите, что переработать</h4>
                    <div className="grid grid-cols-4 gap-3">
                        {REFINING_RECIPES.map(recipe => (
                            <button
                                key={recipe.output.item}
                                onClick={() => handleRecipeSelect(recipe)}
                                className={cn(
                                    "aspect-square flex items-center justify-center bg-card/50 rounded-lg p-2 transition-all border-2",
                                    selectedRecipe?.output.item === recipe.output.item ? "border-primary ring-2 ring-primary/50 scale-105" : "border-transparent hover:border-primary/50"
                                )}
                            >
                                <Image src={itemImageMap[recipe.output.item] as string} alt={getItemName(recipe.output.item)} width={48} height={48} />
                            </button>
                        ))}
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CraftingTab() {
    const { inventory, getItemName } = useInventory();
    const { startProcess, processSlots } = useCrafting();
    const { processSlots: refiningSlots } = useRefining();

    const [selectedCraft, setSelectedCraft] = useState<{ tool: BlueprintType; quality: string; } | null>(null);
    const [quantity, setQuantity] = useState<number>(1);
    
    const totalSlotsUsed = processSlots.current + refiningSlots.current;
    const MAX_FACTORY_SLOTS = 2;

    const allCraftableItems: { tool: BlueprintType; quality: string; }[] = TOOL_TYPES.flatMap(tool => 
        TOOL_QUALITIES.map(quality => ({ tool, quality: quality.id }))
    );

    // Auto-select first available if none selected and the user has the blueprint for it
    useEffect(() => {
        if (!selectedCraft) {
            const firstOwnedCraftable = allCraftableItems.find(item => (inventory[item.tool] || 0) > 0);
            if (firstOwnedCraftable) {
                setSelectedCraft(firstOwnedCraftable);
            }
        }
    }, [inventory, selectedCraft, allCraftableItems]);

    const handleSelectCraft = (item: { tool: BlueprintType; quality: string; }) => {
        setSelectedCraft(item);
        setQuantity(1);
    }

    const selectedTool = selectedCraft?.tool;
    const selectedQuality = selectedCraft?.quality;
    const qualityInfo = TOOL_QUALITIES.find(q => q.id === selectedQuality);
    const materialNeeded = qualityInfo ? qualityInfo.material : null;
    const materialCost = materialNeeded ? CRAFTING_COST_PER_ITEM * quantity : 0;
    const blueprintCost = quantity;

    const hasBlueprintForSelected = selectedTool ? (inventory[selectedTool] || 0) > 0 : false;
    
    const canAfford = hasBlueprintForSelected && selectedTool && qualityInfo && materialNeeded &&
                      (inventory[selectedTool] || 0) >= blueprintCost &&
                      (inventory[materialNeeded] || 0) >= materialCost;

    const handleStartCrafting = () => {
        if (!selectedTool || !selectedQuality) return;
        startProcess({ tool: selectedTool, quality: selectedQuality }, quantity);
        setQuantity(1); // Reset quantity after starting
    };

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardContent className="pt-6 space-y-6">
                <Card className="bg-card/70 p-4">
                    {selectedTool && selectedQuality && qualityInfo && materialNeeded ? (
                        <div className="space-y-4">
                            <div className="flex justify-center items-center gap-2">
                                <div className="flex flex-col items-center w-16 text-center">
                                    <Image src={blueprintImageMap[selectedTool]} alt={getItemName(selectedTool)} width={40} height={40} />
                                    <span className={cn("text-xs mt-1", (inventory[selectedTool] || 0) < blueprintCost ? 'text-destructive' : 'text-muted-foreground')}>
                                        {blueprintCost}/{inventory[selectedTool] || 0}
                                    </span>
                                </div>
                                <Plus className="h-5 w-5 text-muted-foreground" />
                                <div className="flex flex-col items-center w-16 text-center">
                                    <Image src={processedItemImageMap[materialNeeded]} alt={getItemName(materialNeeded)} width={40} height={40} />
                                     <span className={cn("text-xs mt-1", (inventory[materialNeeded] || 0) < materialCost ? 'text-destructive' : 'text-muted-foreground')}>
                                        {materialCost}/{inventory[materialNeeded] || 0}
                                    </span>
                                </div>
                                <ArrowRight className="h-6 w-6 mx-1 sm:mx-2 text-primary" />
                                <div className="flex flex-col items-center w-16 text-center">
                                    <Image src={craftedToolImageMap[getCraftedToolId(selectedTool, selectedQuality)]} alt={getItemName(getCraftedToolId(selectedTool, selectedQuality))} width={40} height={40} />
                                    <span className="text-xs mt-1 text-muted-foreground">
                                        {quantity}
                                    </span>
                                </div>
                            </div>

                            <div className="flex gap-2">
                                <Input
                                    type="number"
                                    min="1"
                                    value={quantity}
                                    onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value, 10) || 1))}
                                    className="w-24"
                                />
                                <Button
                                    onClick={handleStartCrafting}
                                    disabled={!canAfford || totalSlotsUsed >= MAX_FACTORY_SLOTS}
                                    className="flex-grow"
                                >
                                    Создать
                                </Button>
                            </div>
                            <p className="text-xs text-muted-foreground text-center mt-2">Время: {CRAFTING_DURATION_SECONDS_PER_ITEM * quantity} сек.</p>
                        </div>
                    ) : (
                        <p className="text-center text-muted-foreground">Выберите предмет для создания.</p>
                    )}
                </Card>

                <div>
                    <h4 className="font-headline text-lg text-primary mb-2 text-center">Выберите, что создать</h4>
                    <div className="grid grid-cols-4 gap-3">
                        {allCraftableItems.map(item => {
                            const toolId = getCraftedToolId(item.tool, item.quality);
                            const isSelected = selectedCraft?.tool === item.tool && selectedCraft?.quality === item.quality;
                            const hasBlueprint = (inventory[item.tool] || 0) > 0;
                            return (
                                <button
                                    key={toolId}
                                    onClick={() => hasBlueprint && handleSelectCraft(item)}
                                    disabled={!hasBlueprint}
                                    className={cn(
                                        "aspect-square flex items-center justify-center bg-card/50 rounded-lg p-2 transition-all border-2",
                                        isSelected ? "border-primary ring-2 ring-primary/50 scale-105" : "border-transparent",
                                        hasBlueprint ? "hover:border-primary/50 cursor-pointer" : "opacity-50 grayscale cursor-not-allowed"
                                    )}
                                >
                                    <Image src={craftedToolImageMap[toolId]} alt={getItemName(toolId)} width={48} height={48} />
                                    {!hasBlueprint && (
                                        <div className="absolute inset-0 bg-black/60 flex items-center justify-center rounded-lg">
                                            <p className="text-white text-xs font-bold text-center">Нет чертежа</p>
                                        </div>
                                    )}
                                </button>
                            );
                        })}
                    </div>
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

function SharpeningTab() {
    const { getItemName, inventory } = useInventory();
    const [toolType, setToolType] = useState<BlueprintType>('pickaxe');
    const [quality, setQuality] = useState<string>('wooden');
    const [selectedLevel, setSelectedLevel] = useState<number | null>(null);
    const [quantity, setQuantity] = useState(1);

    const SHARPENING_DURATION_SECONDS = 120; // Placeholder
    const SHARPENING_GOLD_COST_MULTIPLIER = 1000; // Placeholder

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

    const getSharpenedItemName = (level: number): string => {
        const baseToolId = getCraftedToolId(toolType, quality);
        const baseToolName = getItemName(baseToolId);
        if (level > 0) {
            return `+${level} ${baseToolName}`;
        }
        return baseToolName;
    };
    
    const selectedToolBaseImage = craftedToolImageMap[getCraftedToolId(toolType, quality)];
    const requiredItemName = selectedLevel ? getSharpenedItemName(selectedLevel - 1) : '';
    // Simplified inventory check: only checks for base (level 0) items.
    const requiredItemCountInInv = selectedLevel === 1 ? (inventory[getCraftedToolId(toolType, quality)] || 0) : 0;
    const requiredItemsForOne = 2;
    const totalRequired = requiredItemsForOne * quantity;
    const canAfford = requiredItemCountInInv >= totalRequired; // Only for level 1 sharpening for now

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardContent className="pt-6 space-y-6">
                <Card className="bg-card/70 p-4">
                    {selectedLevel ? (
                        <div className="space-y-4">
                            <div className="flex justify-center items-center gap-1 sm:gap-2 mb-2 text-primary font-bold">
                                <div className="flex flex-col items-center w-16 text-center">
                                    <Image src={selectedToolBaseImage} alt={requiredItemName} width={40} height={40} />
                                    <span className="text-xs mt-1 text-muted-foreground">{getSharpenedItemName(selectedLevel - 1)}</span>
                                </div>
                                <Plus className="h-5 w-5 text-muted-foreground" />
                                <div className="flex flex-col items-center w-16 text-center">
                                     <Image src={selectedToolBaseImage} alt={requiredItemName} width={40} height={40} />
                                     <span className="text-xs mt-1 text-muted-foreground">{getSharpenedItemName(selectedLevel - 1)}</span>
                                </div>
                                <ArrowRight className="h-6 w-6 mx-1 sm:mx-2 text-primary" />
                                <div className="flex flex-col items-center w-16 text-center">
                                    <Image src={selectedToolBaseImage} alt={getSharpenedItemName(selectedLevel)} width={40} height={40} />
                                     <span className="text-xs mt-1 text-muted-foreground">{getSharpenedItemName(selectedLevel)}</span>
                                </div>
                            </div>
                            
                            <div className="text-sm space-y-2 text-center">
                                <p className="font-semibold">Требуется для x{quantity}:</p>
                                <div className="flex justify-center items-center gap-4 text-muted-foreground">
                                    <span className={!canAfford && selectedLevel === 1 ? 'text-destructive' : ''}>
                                      {requiredItemName}: {totalRequired} (у вас: {requiredItemCountInInv})
                                    </span>
                                    <span>
                                      Золото: {(selectedLevel * SHARPENING_GOLD_COST_MULTIPLIER * quantity).toLocaleString()}
                                    </span>
                                </div>
                            </div>

                            <div className="flex gap-2">
                                <Input
                                    type="number"
                                    min="1"
                                    value={quantity}
                                    onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value, 10) || 1))}
                                    className="w-24"
                                />
                                <Button
                                    disabled 
                                    className="flex-grow"
                                >
                                    Заточить
                                </Button>
                            </div>
                            <p className="text-xs text-muted-foreground text-center mt-2">Время: {SHARPENING_DURATION_SECONDS * quantity} сек.</p>
                        </div>
                    ) : (
                        <p className="text-center text-muted-foreground">Выберите уровень заточки.</p>
                    )}
                </Card>

                <div>
                    <h4 className="font-headline text-lg text-primary mb-4 text-center">Выберите, что и до какого уровня заточить</h4>
                    <div className="grid grid-cols-2 gap-4 mb-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium">Категория</label>
                            <Select value={toolType} onValueChange={(val: BlueprintType) => {setSelectedLevel(null); setToolType(val)}}>
                                <SelectTrigger><SelectValue /></SelectTrigger>
                                <SelectContent>
                                    {TOOL_TYPES.map(t => <SelectItem key={t} value={t}>{toolTypeDisplayNames[t]}</SelectItem>)}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium">Качество</label>
                            <Select value={quality} onValueChange={(val) => {setSelectedLevel(null); setQuality(val)}}>
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
                                {selectedToolBaseImage && (
                                    <Image src={selectedToolBaseImage} alt={getSharpenedItemName(level)} width={48} height={48} className="flex-shrink-0" />
                                )}
                                <span className="text-sm font-bold mt-1">+{level}</span>
                            </button>
                        ))}
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

const TABS = [
  { value: 'refining', label: 'Переработка', icon: <FlaskConical className="w-6 h-6" /> },
  { value: 'crafting', label: 'Крафт', icon: <Anvil className="w-6 h-6" /> },
  { value: 'sharpening', label: 'Заточка', icon: <Sparkles className="w-6 h-6" /> },
  { value: 'repairing', label: 'Починка', icon: <Wrench className="w-6 h-6" /> },
] as const;

export default function FactoryPage() {
  const [activeTab, setActiveTab] = useState<(typeof TABS)[number]['value']>('refining');

  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
        <CardContent className="p-2 sm:p-4">
          <Tabs defaultValue="refining" onValueChange={(value) => setActiveTab(value as any)} className="w-full">
            <TabsList className="flex w-full items-center justify-around rounded-lg p-1 mb-2">
               {TABS.map((tab) => {
                  const isActive = activeTab === tab.value;
                  return (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className={cn(
                        "flex flex-col items-center justify-center transition-all duration-300 ease-in-out rounded-lg p-0",
                        "data-[state=active]:bg-primary/20 data-[state=active]:shadow-lg",
                        isActive
                          ? "h-16 w-24 gap-1 text-primary"
                          : "h-14 w-14 text-muted-foreground hover:bg-muted/50"
                      )}
                    >
                      {tab.icon}
                      {isActive && <span className="text-xs font-medium mt-1">{tab.label}</span>}
                    </TabsTrigger>
                  )
                })}
            </TabsList>
            
            <ActiveProcessesDisplay />

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
