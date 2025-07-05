"use client";

import { useState } from 'react';
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Archive, Gem, FlaskConical, Wrench } from 'lucide-react';
import { cn } from '@/lib/utils';

import { useChests } from "@/contexts/ChestContext";
import { useInventory } from "@/contexts/InventoryContext";
import Image from 'next/image';
import { allChestTypes, chestDetails } from "@/lib/chest-definitions";
import { openChest } from '@/lib/loot-tables';
import type { ChestType } from '@/types/profile';
import type { LootResult, InventoryItemType, BlueprintType, ResourceType, ProcessedItemType, ReagentType, CraftedToolType } from '@/types/inventory';
import { AllResourceTypes, AllReagentTypes, AllBlueprintTypes, AllProcessedItemTypes, AllCraftedToolTypes } from '@/types/inventory';

const TABS = [
  { value: 'chests', label: 'Сундуки', icon: <Archive className="w-6 h-6" /> },
  { value: 'resources', label: 'Ресурсы', icon: <Gem className="w-6 h-6" /> },
  { value: 'reagents', label: 'Реагенты', icon: <FlaskConical className="w-6 h-6" /> },
  { value: 'tools', label: 'Инструменты', icon: <Wrench className="w-6 h-6" /> },
] as const;


export default function InventoryPage() {
  const { chestCounts, getChestName, spendChests } = useChests();
  const { inventory, addItems, getItemName } = useInventory();

  const [activeTab, setActiveTab] = useState<(typeof TABS)[number]['value']>('chests');
  const [selectedChest, setSelectedChest] = useState<ChestType | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [openAmount, setOpenAmount] = useState(0);

  const ownedChests = allChestTypes.filter(chestType => (chestCounts[chestType] || 0) > 0);

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

  const chestImageMap: Partial<Record<ChestType, string>> = {
    'resource_small': '/images/small-chess-res.png',
    'resource_medium': '/images/medium-chess-res.png',
    'resource_large': '/images/big-chess-res.png',
    'reagent_small': '/images/small-chess-ing.png',
    'reagent_medium': '/images/medium-chess-ing.png',
    'reagent_large': '/images/big-chess-ing.png',
    'blueprint': '/images/chess-blueprint.png',
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

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1).replace(/\.0$/, '') + 'K';
    }
    return num.toString();
  };

  const handleChestClick = (chestType: ChestType) => {
    setSelectedChest(chestType);
    setIsModalOpen(true);
  };

  const handleOpenClick = (amount: number) => {
    if (!selectedChest) return;
    const available = chestCounts[selectedChest] || 0;
    let amountToOpen = amount;
    if (amount === -1) { // -1 signifies "Open All"
      amountToOpen = available;
    } else {
      amountToOpen = Math.min(amount, available);
    }
    
    if (amountToOpen > 0) {
        setOpenAmount(amountToOpen);
        setIsAlertOpen(true);
    }
  };

  const handleConfirmOpen = () => {
    if (!selectedChest || openAmount <= 0) return;

    const totalLoot: LootResult = {};
    for (let i = 0; i < openAmount; i++) {
        const singleLoot = openChest(selectedChest);
        for (const key in singleLoot) {
            const itemKey = key as InventoryItemType;
            const amount = singleLoot[itemKey] || 0;
            totalLoot[itemKey] = (totalLoot[itemKey] || 0) + amount;
        }
    }

    spendChests(selectedChest, openAmount);
    addItems(totalLoot);

    const lootLines = Object.keys(totalLoot).map(key => {
        const itemKey = key as InventoryItemType;
        const amount = totalLoot[itemKey];
        return `${getItemName(itemKey)}: ${amount?.toLocaleString() ?? 0} шт.`;
    });
    const lootMessage = lootLines.length > 0 ? lootLines.join('\n') : "Сундук оказался пуст.";
    alert(`Вы получили:\n\n${lootMessage}`);

    setIsAlertOpen(false);
    setIsModalOpen(false);
    setSelectedChest(null);
  };

  const currentChestDetails = selectedChest ? chestDetails[selectedChest] : null;
  const currentChestCount = selectedChest ? chestCounts[selectedChest] || 0 : 0;

  const ownedResources = AllResourceTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedReagents = AllReagentTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedBlueprints = AllBlueprintTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedProcessedItems = AllProcessedItemTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedCraftedTools = AllCraftedToolTypes.filter(item => (inventory[item] || 0) > 0);

  const ItemGrid = ({ items, imageMap, title }: { items: InventoryItemType[], imageMap: Record<string, string>, title?: string }) => (
    items.length > 0 ? (
      <div>
        {title && <h4 className="text-lg font-semibold text-primary mb-4">{title}</h4>}
        <div className="grid grid-cols-4 sm:grid-cols-5 gap-4">
          {items.map(item => (
            <div key={item} className="relative aspect-square flex items-center justify-center">
              <Image
                src={imageMap[item as keyof typeof imageMap]}
                alt={getItemName(item)}
                width={64}
                height={64}
                className="w-full h-full object-contain"
              />
              <span className="absolute bottom-0 right-0 bg-background/80 backdrop-blur-sm text-foreground text-xs font-bold px-1.5 py-0.5 rounded-full shadow-md">
                {formatNumber(inventory[item] || 0)}
              </span>
            </div>
          ))}
        </div>
      </div>
    ) : null
  );

  return (
    <>
      <div className="flex flex-col items-center justify-start min-h-full p-4 text-foreground">
        <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
          <CardContent className="p-2 sm:p-4">
            <Tabs defaultValue="chests" onValueChange={(value) => setActiveTab(value as any)} className="w-full">
              <TabsList className="flex w-full items-center justify-around rounded-lg p-1 mb-4">
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

              <TabsContent value="chests">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <div className="grid grid-cols-4 sm:grid-cols-5 gap-4">
                    {ownedChests.map(chestType => (
                      <div
                        key={chestType}
                        className="relative aspect-square flex items-center justify-center cursor-pointer transition-transform hover:scale-105"
                        onClick={() => handleChestClick(chestType)}
                        role="button"
                        tabIndex={0}
                        aria-label={`Открыть сундук: ${getChestName(chestType)}`}
                      >
                        <Image
                          src={chestImageMap[chestType] ?? `https://placehold.co/64x64.png`}
                          alt={getChestName(chestType)}
                          width={64}
                          height={64}
                          className="w-full h-full object-contain"
                          data-ai-hint={chestImageMap[chestType] ? undefined : chestDetails[chestType]?.hint || 'treasure chest'}
                        />
                        <span className="absolute bottom-0 right-0 bg-background/80 backdrop-blur-sm text-foreground text-xs font-bold px-1.5 py-0.5 rounded-full shadow-md">
                          {formatNumber(chestCounts[chestType] || 0)}
                        </span>
                      </div>
                    ))}
                    {ownedChests.length === 0 && (
                      <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет сундуков.</p>
                    )}
                  </div>
                </ScrollArea>
              </TabsContent>
              
              <TabsContent value="resources">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <div className="space-y-6">
                    <ItemGrid items={ownedResources} imageMap={resourceImageMap} title="Сырые ресурсы" />
                    <ItemGrid items={ownedProcessedItems} imageMap={processedItemImageMap} title="Изделия" />
                    <ItemGrid items={ownedBlueprints} imageMap={blueprintImageMap} title="Чертежи" />
                    {ownedResources.length === 0 && ownedProcessedItems.length === 0 && ownedBlueprints.length === 0 && (
                      <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет ресурсов, изделий или чертежей.</p>
                    )}
                  </div>
                </ScrollArea>
              </TabsContent>

              <TabsContent value="reagents">
                 <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <ItemGrid items={ownedReagents} imageMap={reagentImageMap} />
                  {ownedReagents.length === 0 && (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет реагентов.</p>
                  )}
                </ScrollArea>
              </TabsContent>
              
              <TabsContent value="tools">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <ItemGrid items={ownedCraftedTools} imageMap={craftedToolImageMap as any} />
                  {ownedCraftedTools.length === 0 && (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет готовых инструментов.</p>
                  )}
                </ScrollArea>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>

      {/* Main Chest Info Dialog */}
      <Dialog open={isModalOpen} onOpenChange={(open) => {
        setIsModalOpen(open);
        if (!open) setSelectedChest(null);
      }}>
        <DialogContent className="sm:max-w-[425px]">
          {currentChestDetails && (
            <>
              <DialogHeader>
                <DialogTitle className="text-primary text-2xl">{currentChestDetails.name}</DialogTitle>
                <DialogDescription>
                  {currentChestDetails.description}
                  <br />
                  <span className="font-bold text-foreground">У вас: {currentChestCount} шт.</span>
                </DialogDescription>
              </DialogHeader>
              <DialogFooter className="flex-col sm:flex-col sm:space-x-0 gap-2 mt-4">
                <Button onClick={() => handleOpenClick(1)} disabled={currentChestCount < 1}>Открыть</Button>
                <Button onClick={() => handleOpenClick(10)} disabled={currentChestCount < 10}>Открыть (10)</Button>
                <Button onClick={() => handleOpenClick(-1)} disabled={currentChestCount < 1}>Открыть все</Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
      
      {/* Confirmation Alert Dialog */}
      <AlertDialog open={isAlertOpen} onOpenChange={setIsAlertOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Подтверждение</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите открыть "{currentChestDetails?.name}" в количестве {openAmount} шт.?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отменить</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmOpen}>Подтвердить</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
