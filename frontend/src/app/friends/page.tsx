
"use client";

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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

import { useChests } from "@/contexts/ChestContext";
import { useInventory } from "@/contexts/InventoryContext";
import Image from 'next/image';
import { allChestTypes, chestDetails } from "@/lib/chest-definitions";
import { openChest } from '@/lib/loot-tables';
import type { ChestType } from '@/types/profile';
import type { LootResult, InventoryItemType, BlueprintType } from '@/types/inventory';
import { AllResourceTypes, AllReagentTypes, AllBlueprintTypes, AllProcessedItemTypes, AllCraftedToolTypes } from '@/types/inventory';


export default function InventoryPage() {
  const { chestCounts, getChestName, spendChests } = useChests();
  const { inventory, addItems, getItemName } = useInventory();

  const [selectedChest, setSelectedChest] = useState<ChestType | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [openAmount, setOpenAmount] = useState(0);

  const ownedChests = allChestTypes.filter(chestType => (chestCounts[chestType] || 0) > 0);

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

  return (
    <>
      <div className="flex flex-col items-center justify-start min-h-full p-4 text-foreground">
        <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
          <CardHeader>
            <CardTitle className="text-3xl font-headline text-center text-primary">Инвентарь</CardTitle>
          </CardHeader>
          <CardContent className="pt-2">
            <Tabs defaultValue="chests" className="w-full">
              <TabsList className="grid w-full grid-cols-5 mb-4 text-xs sm:text-sm">
                <TabsTrigger value="chests">Сундуки</TabsTrigger>
                <TabsTrigger value="resources">Ресурсы</TabsTrigger>
                <TabsTrigger value="reagents">Реагенты</TabsTrigger>
                <TabsTrigger value="processed">Изделия</TabsTrigger>
                <TabsTrigger value="tools">Инструменты</TabsTrigger>
              </TabsList>

              <TabsContent value="chests">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <div className="grid grid-cols-3 sm:grid-cols-4 gap-4">
                    {ownedChests.map(chestType => (
                      <div
                        key={chestType}
                        className="relative flex flex-col items-center justify-start p-1 space-y-1 cursor-pointer transition-transform hover:scale-105"
                        onClick={() => handleChestClick(chestType)}
                        role="button"
                        tabIndex={0}
                        aria-label={`Открыть сундук: ${getChestName(chestType)}`}
                      >
                        <Image
                          src={`https://placehold.co/64x64.png`}
                          alt={getChestName(chestType)}
                          width={64}
                          height={64}
                          className="rounded-md"
                          data-ai-hint={chestDetails[chestType]?.hint || 'treasure chest'}
                        />
                        <p className="text-center text-xs text-muted-foreground leading-tight min-h-[2rem] flex items-center justify-center">
                          {getChestName(chestType)}
                        </p>
                        <span className="absolute -bottom-1 -right-1 bg-primary text-primary-foreground text-xs font-bold px-1.5 py-0.5 rounded-full shadow-md">
                          {chestCounts[chestType]}
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
                  {ownedResources.length > 0 ? (
                    <ul className="space-y-2">
                      {ownedResources.map(item => (
                        <li key={item} className="flex justify-between items-center p-2 rounded-md bg-card/50">
                          <span className="text-foreground">{getItemName(item)}</span>
                          <span className="font-bold text-primary">{inventory[item]?.toLocaleString() ?? 0} шт.</span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет ресурсов.</p>
                  )}
                </ScrollArea>
              </TabsContent>

              <TabsContent value="reagents">
                 <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  {ownedReagents.length > 0 ? (
                    <ul className="space-y-2">
                      {ownedReagents.map(item => (
                        <li key={item} className="flex justify-between items-center p-2 rounded-md bg-card/50">
                          <span className="text-foreground">{getItemName(item)}</span>
                          <span className="font-bold text-primary">{inventory[item]?.toLocaleString() ?? 0} шт.</span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет реагентов.</p>
                  )}
                </ScrollArea>
              </TabsContent>
              
              <TabsContent value="processed">
                 <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  {ownedProcessedItems.length > 0 ? (
                    <ul className="space-y-2">
                      {ownedProcessedItems.map(item => (
                        <li key={item} className="flex justify-between items-center p-2 rounded-md bg-card/50">
                          <span className="text-foreground">{getItemName(item)}</span>
                          <span className="font-bold text-primary">{inventory[item]?.toLocaleString() ?? 0} шт.</span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет изделий.</p>
                  )}
                </ScrollArea>
              </TabsContent>

              <TabsContent value="tools">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  {ownedBlueprints.length > 0 || ownedCraftedTools.length > 0 ? (
                    <ul className="space-y-2">
                      {ownedCraftedTools.length > 0 && (
                        <>
                          <p className="text-sm font-semibold text-muted-foreground mb-2">Готовые инструменты</p>
                          {ownedCraftedTools.map(item => (
                            <li key={item} className="flex justify-between items-center p-2 rounded-md bg-card/50">
                              <span className="text-foreground">{getItemName(item)}</span>
                              <span className="font-bold text-primary">{inventory[item]?.toLocaleString() ?? 0} шт.</span>
                            </li>
                          ))}
                        </>
                      )}
                      {ownedBlueprints.length > 0 && (
                        <>
                          <p className="text-sm font-semibold text-muted-foreground my-2 pt-2 border-t border-border">Чертежи</p>
                          {ownedBlueprints.map(item => (
                            <li key={item} className="flex justify-between items-center p-2 rounded-md bg-card/50">
                              <span className="text-foreground">{getItemName(item)}</span>
                              <span className="font-bold text-primary">{inventory[item]?.toLocaleString() ?? 0} шт.</span>
                            </li>
                          ))}
                        </>
                      )}
                    </ul>
                  ) : (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет инструментов и чертежей.</p>
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
