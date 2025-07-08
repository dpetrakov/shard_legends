
"use client";

import React, { useState, useEffect, useCallback } from 'react';
import Image from 'next/image';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { cn } from "@/lib/utils";
import { FlaskConical, Anvil, Sparkles, Wrench, ArrowRight, Plus } from 'lucide-react';
import { useInventory } from '@/contexts/InventoryContext';
import { useMining } from '@/contexts/MiningContext';
import type { BlueprintType, CraftedToolType, InventoryItemType, ResourceType, ReagentType, ProcessedItemType } from '@/types/inventory';
import { AllProcessedItemTypes, AllCraftedToolTypes, AllBlueprintTypes } from '@/types/inventory';
import { useProduction } from '@/contexts/ProductionContext';
import type { ProductionRecipe, ProductionTask } from '@/types/production';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";


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
    const { tasks, getRecipeById, claimCompleted } = useProduction();
    const { getItemImage } = useInventory();
    
    const MAX_FACTORY_SLOTS = 2; // This should probably come from context/API in the future
    const totalSlotsUsed = tasks.length;
    const completedTasksCount = tasks.filter(t => t.status === 'completed').length;

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
            {completedTasksCount > 0 && (
                 <Button onClick={claimCompleted} className="w-full">
                    Забрать все готовые ({completedTasksCount})
                </Button>
            )}
            {tasks.map(task => {
                const recipe = getRecipeById(task.recipe_id);
                if (!recipe) return null; // Should not happen if data is consistent

                const isFinished = task.status === 'completed';
                const endTimeMs = new Date(task.ends_at).getTime();
                const totalDuration = recipe.duration_seconds * task.quantity * 1000;
                const startTimeMs = endTimeMs - totalDuration;
                
                let progress = 0;
                if(isFinished) {
                    progress = 100;
                } else if (task.status === 'in_progress' && totalDuration > 0) {
                    const timeElapsed = Date.now() - startTimeMs;
                    progress = Math.min(100, (timeElapsed / totalDuration) * 100);
                }
                
                const outputItem = recipe.output_items?.[0];
                if (!outputItem) return null;

                const outputImage = getItemImage(outputItem.item_slug) ?? `https://placehold.co/64x64.png`;

                return (
                    <Card key={task.id} className="bg-card/70 p-3">
                        <div className="flex items-center gap-3">
                            <div className="relative flex-shrink-0">
                                <Image 
                                    src={outputImage} 
                                    alt={recipe.name} 
                                    width={48} 
                                    height={48} 
                                    className="bg-background/50 rounded-md p-1"
                                />
                                <span className="absolute -top-1 -right-1 bg-primary text-primary-foreground text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center border-2 border-card">
                                    {task.quantity}
                                </span>
                            </div>

                            <div className="flex-grow space-y-1">
                                <p className="font-semibold text-sm leading-tight">{recipe.name}</p>
                                <Progress value={progress} className="h-2" />
                                <div className="flex justify-between items-center text-xs text-muted-foreground">
                                    {isFinished || task.status === 'queued' ? <span>{task.status === 'queued' ? 'В очереди' : 'Готово!'}</span> : <CountdownTimer endTime={endTimeMs} />}
                                    <span>{isFinished ? '100%' : `${Math.floor(progress)}%`}</span>
                                </div>
                            </div>
                        </div>
                    </Card>
                );
            })}
        </div>
    );
}


const ProductionTab = ({ category }: { category: 'refining' | 'crafting' }) => {
    const { inventory, getItemName, getItemImage } = useInventory();
    const { recipes, tasks, startProduction } = useProduction();
    
    const [selectedRecipe, setSelectedRecipe] = useState<ProductionRecipe | null>(null);
    const [quantity, setQuantity] = useState(1);
    
    const MAX_FACTORY_SLOTS = 2; // Should come from API/context
    const totalSlotsUsed = tasks.length;
    
    const getRecipeCategory = (recipe: ProductionRecipe): 'refining' | 'crafting' => {
        const outputItemSlug = recipe.output_items?.[0]?.item_slug;
        if (outputItemSlug && AllProcessedItemTypes.includes(outputItemSlug as any)) {
            return 'refining';
        }
        return 'crafting';
    };

    const relevantRecipes = recipes.filter(r => getRecipeCategory(r) === category);

    useEffect(() => {
        // Auto-select first recipe when tab is switched
        if (relevantRecipes.length > 0) {
            setSelectedRecipe(relevantRecipes[0]);
            setQuantity(1);
        } else {
            setSelectedRecipe(null);
        }
    }, [category, recipes]);


    const handleRecipeSelect = (recipe: ProductionRecipe) => {
        setSelectedRecipe(recipe);
        setQuantity(1);
    };

    const handleStartProcess = () => {
        if (!selectedRecipe) return;
        startProduction(selectedRecipe.id, quantity);
        setQuantity(1);
    };
    
    const canAfford = selectedRecipe && (selectedRecipe.input_items || []).every(
        (ing) => (inventory[ing.item_slug] || 0) >= ing.quantity * quantity
    );

    return (
        <Card className="bg-background/50 shadow-inner">
            <CardContent className="pt-6 space-y-6">
                 {selectedRecipe ? (
                    <Card className="bg-card/70 p-4">
                        <div className="flex items-center justify-center gap-1 sm:gap-2 mb-4 text-primary font-bold">
                            {(selectedRecipe.input_items || []).map((ing, index) => (
                                <React.Fragment key={ing.item_slug}>
                                    {index > 0 && <Plus className="h-5 w-5 text-muted-foreground" />}
                                    <div className="flex flex-col items-center w-16 text-center">
                                        <Image src={getItemImage(ing.item_slug) ?? `https://placehold.co/40x40.png`} alt={getItemName(ing.item_slug)} width={40} height={40} />
                                        <span className={cn("text-xs mt-1", (inventory[ing.item_slug] || 0) < ing.quantity * quantity ? 'text-destructive' : 'text-muted-foreground')}>
                                            {ing.quantity * quantity}/{inventory[ing.item_slug] || 0}
                                        </span>
                                    </div>
                                </React.Fragment>
                            ))}
                            <ArrowRight className="h-6 w-6 mx-1 sm:mx-2 text-primary" />
                             {(selectedRecipe.output_items || []).map((out) => (
                                <div key={out.item_slug} className="flex flex-col items-center w-16 text-center">
                                    <Image src={getItemImage(out.item_slug) ?? `https://placehold.co/40x40.png`} alt={getItemName(out.item_slug)} width={40} height={40} />
                                    <span className="text-xs mt-1 text-muted-foreground">
                                        {out.quantity * quantity}
                                    </span>
                                </div>
                             ))}
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
                                {category === 'refining' ? 'Переплавить' : 'Создать'}
                            </Button>
                        </div>
                        <p className="text-xs text-muted-foreground text-center mt-2">Время: {selectedRecipe.duration_seconds * quantity} сек.</p>
                    </Card>
                ) : <p className="text-center text-muted-foreground">Выберите предмет для производства</p>}

                <div>
                    <h4 className="font-headline text-lg text-primary mb-2 text-center">Выберите, что произвести</h4>
                    <div className="grid grid-cols-4 gap-3">
                        {relevantRecipes.map(recipe => {
                            const outputItem = recipe.output_items?.[0];
                            if (!outputItem) return null;
                            
                            const hasBlueprint = (() => {
                                if (category !== 'crafting') {
                                    return true; // Refining recipes don't need blueprints
                                }
                                // Check if any input is a blueprint
                                const requiredBlueprint = (recipe.input_items || []).find(ing => AllBlueprintTypes.includes(ing.item_slug as any));
                                if (requiredBlueprint) {
                                    // If a blueprint is required, check if player has it
                                    return (inventory[requiredBlueprint.item_slug] || 0) > 0;
                                }
                                // If no blueprint is required for this crafting recipe, it's always available to be selected
                                return true;
                            })();
                            
                            return (
                                <button
                                    key={recipe.id}
                                    onClick={() => handleRecipeSelect(recipe)}
                                    className={cn(
                                        "relative aspect-square flex items-center justify-center bg-card/50 rounded-lg p-2 transition-all border-2",
                                        selectedRecipe?.id === recipe.id ? "border-primary ring-2 ring-primary/50 scale-105" : "border-transparent hover:border-primary/50",
                                        hasBlueprint ? "" : "opacity-50 grayscale cursor-not-allowed"
                                    )}
                                    disabled={!hasBlueprint}
                                >
                                    <Image src={getItemImage(outputItem.item_slug) ?? `https://placehold.co/48x48.png`} alt={getItemName(outputItem.item_slug)} width={48} height={48} />
                                     {!hasBlueprint && (
                                        <div className="absolute inset-0 bg-black/60 flex items-center justify-center rounded-lg">
                                            <p className="text-white text-xs font-bold text-center">Нет чертежа</p>
                                        </div>
                                    )}
                                </button>
                            )
                        })}
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}

function RepairingTab() {
    const { equippedTools } = useMining();
    const { getItemName, getItemImage } = useInventory();

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
                                    src={getItemImage(tool.item) ?? `https://placehold.co/48x48.png`}
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
                                        <Image src={getItemImage('metal_ingot') ?? `https://placehold.co/16x16.png`} alt={getItemName('metal_ingot')} width={16} height={16} />
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
    const { getItemName, inventory, getItemImage } = useInventory();
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
    
    const getCraftedToolId = (tool: BlueprintType, qualityId: string): CraftedToolType => {
        return `${qualityId}_${tool}` as CraftedToolType;
    };

    const getSharpenedItemName = (level: number): string => {
        const baseToolId = getCraftedToolId(toolType, quality);
        const baseToolName = getItemName(baseToolId);
        if (level > 0) {
            return `+${level} ${baseToolName}`;
        }
        return baseToolName;
    };
    
    const selectedToolBaseId = getCraftedToolId(toolType, quality);
    const selectedToolBaseImage = getItemImage(selectedToolBaseId) ?? `https://placehold.co/48x48.png`;
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
                                    {['axe', 'pickaxe', 'shovel', 'sickle'].map(t => <SelectItem key={t} value={t}>{toolTypeDisplayNames[t as BlueprintType]}</SelectItem>)}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium">Качество</label>
                            <Select value={quality} onValueChange={(val) => {setSelectedLevel(null); setQuality(val)}}>
                                <SelectTrigger><SelectValue /></SelectTrigger>
                                <SelectContent>
                                    {[{id: 'wooden', name: 'Деревянный'},{id: 'stone', name: 'Каменный'},{id: 'metal', name: 'Металлический'},{id: 'diamond', name: 'Бриллиантовый'}].map(q => <SelectItem key={q.id} value={q.id}>{q.name}</SelectItem>)}
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
  const { fetchData } = useProduction();

  useEffect(() => {
    fetchData();
  }, [fetchData]);


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
                <ProductionTab category="refining" />
            </TabsContent>
            <TabsContent value="crafting">
                <ProductionTab category="crafting" />
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
