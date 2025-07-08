
"use client";

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Archive, Gem, FlaskConical, Wrench, Loader2, Code } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useInventory } from "@/contexts/InventoryContext";
import { useAuth } from '@/contexts/AuthContext';
import Image from 'next/image';
import { allChestTypes } from "@/lib/chest-definitions";
import type { ChestType } from '@/types/profile';
import type { InventoryItemType, BlueprintType, ResourceType, ProcessedItemType, ReagentType, CraftedToolType } from '@/types/inventory';
import { AllResourceTypes, AllReagentTypes, AllBlueprintTypes, AllProcessedItemTypes, AllCraftedToolTypes } from '@/types/inventory';
import { useToast } from '@/hooks/use-toast';
import { motion, AnimatePresence } from 'framer-motion';
import { Separator } from "@/components/ui/separator";

const TABS = [
  { value: 'chests', label: 'Сундуки', icon: <Archive className="w-6 h-6" /> },
  { value: 'resources', label: 'Ресурсы', icon: <Gem className="w-6 h-6" /> },
  { value: 'reagents', label: 'Реагенты', icon: <FlaskConical className="w-6 h-6" /> },
  { value: 'tools', label: 'Инструменты', icon: <Wrench className="w-6 h-6" /> },
  { value: 'debug', label: 'Отладка', icon: <Code className="w-6 h-6" /> },
] as const;


// Light rays animation component
const LightRays = () => {
    return (
        <motion.div
            className="absolute z-0 w-64 h-64" // z-0 so it's behind the chest
            initial={{ opacity: 0, scale: 0.5 }}
            animate={{ 
                opacity: 1, 
                scale: 1.2, 
                transition: { delay: 0.5, duration: 0.5 } // Appear after chest lands
            }}
            exit={{ opacity: 0, scale: 0.8, transition: { duration: 0.3 } }}
        >
            {/* Spinning rays */}
            <div className="absolute inset-0 animate-rays-spin-slow">
                {[...Array(8)].map((_, i) => (
                    <div 
                        key={i}
                        className="absolute top-1/2 left-0 w-full h-px bg-gradient-to-r from-yellow-200/0 via-yellow-200 to-yellow-200/0"
                        style={{ transform: `rotate(${i * 22.5}deg)` }}
                    />
                ))}
            </div>
            {/* Pulsating glow */}
            <div className="absolute inset-0 bg-yellow-300/20 rounded-full blur-2xl animate-pulse" />
        </motion.div>
    );
};

// Animation component for chest opening
const ChestOpeningAnimation = ({ chestImageUrl, state }: { chestImageUrl: string | undefined, state: 'idle' | 'shaking' | 'flashing' }) => {
    if (!chestImageUrl || state === 'idle') return null;

    const containerVariants = {
        hidden: { opacity: 0, scale: 0.7 },
        visible: { 
            opacity: 1, 
            scale: 1,
            transition: { duration: 0.5, ease: 'easeOut' } 
        },
    };

    const shakeVariants = {
        shaking: {
            rotate: [0, -3, 3, -3, 3, 0],
            transition: { duration: 1.5, repeat: Infinity, delay: 0.5 }
        }
    };

    const flashVariants = {
        hidden: { opacity: 0 },
        visible: { opacity: [0, 1, 0], transition: { duration: 0.7, times: [0, 0.5, 1] } }
    };

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center pointer-events-none">
            <AnimatePresence>
                {state === 'shaking' && (
                    <motion.div
                        key="chest-container"
                        className="relative flex items-center justify-center"
                        variants={containerVariants}
                        initial="hidden"
                        animate="visible"
                        exit={{ opacity: 0, scale: 2, transition: {duration: 0.4} }}
                    >
                        <LightRays />
                        <motion.div
                            key="chest"
                            className="z-10"
                            variants={shakeVariants}
                            animate='shaking'
                        >
                            <Image src={chestImageUrl} alt="Открытие сундука" width={128} height={128} />
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
            <AnimatePresence>
                {state === 'flashing' && (
                    <motion.div
                        key="flash"
                        className="absolute inset-0 bg-white"
                        variants={flashVariants}
                        initial="hidden"
                        animate="visible"
                    />
                )}
            </AnimatePresence>
        </div>
    );
};


export default function InventoryPage() {
  const { inventory, itemDetails, getItemName, getItemImage, syncWithServer } = useInventory();
  const { token } = useAuth();
  const { toast } = useToast();
  const apiUrl = 'https://dev-forly.slcw.dimlight.online';

  const [activeTab, setActiveTab] = useState<(typeof TABS)[number]['value']>('chests');
  const [selectedChest, setSelectedChest] = useState<ChestType | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [openAmount, setOpenAmount] = useState(0);

  // States for server logic
  const [isLoading, setIsLoading] = useState(false);
  const [lastLoot, setLastLoot] = useState<any[] | null>(null);
  const [isLootModalOpen, setIsLootModalOpen] = useState(false);
  
  // Animation states
  const [animationState, setAnimationState] = useState<'idle' | 'shaking' | 'flashing'>('idle');
  const [animatingChest, setAnimatingChest] = useState<ChestType | null>(null);
  
  // States for debug tab
  const [debugResponse, setDebugResponse] = useState<string>('');
  const [isDebugLoading, setIsDebugLoading] = useState<boolean>(false);
  const [debugDefinitionsResponse, setDebugDefinitionsResponse] = useState<string>('');
  const [isDebugDefinitionsLoading, setIsDebugDefinitionsLoading] = useState<boolean>(false);

  const ownedChests = allChestTypes.filter(chestType => (inventory[chestType] || 0) > 0);

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
    const available = inventory[selectedChest] || 0;
    let amountToOpen;

    if (amount === -1) {
      amountToOpen = -1; // Use -1 as a special value for open_all
    } else {
      amountToOpen = Math.min(amount, available);
    }
    
    if (amountToOpen !== 0) { // Allow -1 and positive numbers
        setOpenAmount(amountToOpen);
        setIsModalOpen(false); // Close options modal
        setIsAlertOpen(true); // Open confirmation modal
    }
  };

  const handleConfirmOpen = async () => {
    if (!selectedChest || openAmount === 0 || !token || isLoading) return;

    setIsLoading(true);
    setIsAlertOpen(false);
    setAnimatingChest(selectedChest);
    
    try {
      // 1. Start shaking animation and call API
      setAnimationState('shaking');
      
      const [type, quality] = selectedChest.split('_');
      let requestBody: any;
      if (type === 'blueprint') {
          requestBody = { chest_type: 'blueprint_chest' };
      } else {
           requestBody = { chest_type: `${type}_chest`, quality_level: quality };
      }
      if (openAmount === -1) {
        requestBody.open_all = true;
      } else if (openAmount > 0) {
        requestBody.quantity = openAmount;
      }

      const response = await fetch(`${apiUrl}/api/deck/chest/open`, {
          method: 'POST',
          mode: 'cors',
          headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
              'Accept': 'application/json'
          },
          body: JSON.stringify(requestBody)
      });

      const responseData = await response.json();
      if (!response.ok) { throw new Error(responseData.message || 'Не удалось открыть сундук.'); }
      
      setLastLoot(responseData.items || []);
      await syncWithServer();

      // Wait for shaking animation to play out (2.5 seconds)
      await new Promise(res => setTimeout(res, 2500));

      // 2. Flash
      setAnimationState('flashing');
      await new Promise(res => setTimeout(res, 700));

      // 3. Reset state and show loot
      setAnimationState('idle');
      setAnimatingChest(null);
      if (responseData.items && responseData.items.length > 0) {
        setIsLootModalOpen(true);
      } else {
        toast({ title: "Сундук оказался пуст", description: "В этот раз вам не повезло." });
      }

    } catch (error: any) {
      toast({ variant: "destructive", title: "Ошибка", description: error.message });
      setAnimationState('idle');
      setAnimatingChest(null);
      setLastLoot(null);
    } finally {
      setIsLoading(false);
    }
  };


  const getLootItemSlug = (item: any): InventoryItemType | ChestType | null => {
    const { item_class, item_type, quality_level } = item;
    if (!item_class || !item_type) return null;
    
    switch (item_class) {
      case 'chests':
        if (item_type === 'blueprint' || item_type === 'blueprint_chest') {
          return 'blueprint';
        }
        if (quality_level) {
          const baseType = item_type.replace('_chest', '');
          return `${baseType}_${quality_level}` as ChestType;
        }
        break;
      case 'resources':
      case 'reagents':
      case 'processed_items':
      case 'blueprints':
        return item_type as InventoryItemType;
      case 'tools':
        if (quality_level) {
          return `${quality_level}_${item_type}` as CraftedToolType;
        }
        break;
    }
    return null;
  };

  const handleDebugSync = async () => {
    if (!token) {
        setDebugResponse('Ошибка: Пользователь не авторизован.');
        return;
    }
    setIsDebugLoading(true);
    setDebugResponse('Загрузка...');
    try {
        const requestUrl = `${apiUrl}/api/inventory/items`;
        const response = await fetch(requestUrl, {
            method: 'GET',
            mode: 'cors',
            headers: { 
                'Authorization': `Bearer ${token}`,
                'Accept': 'application/json'
            },
        });
        
        const responseBodyText = await response.text();
        
        try {
            const json = JSON.parse(responseBodyText);
            setDebugResponse(JSON.stringify(json, null, 2));
        } catch {
            setDebugResponse(responseBodyText);
        }

        await syncWithServer();

    } catch (error: any) {
        setDebugResponse(`Сетевая ошибка: ${error.message}`);
    } finally {
        setIsDebugLoading(false);
    }
  };
  
  const handleDebugDefinitions = async () => {
    if (!token) {
        setDebugDefinitionsResponse('Ошибка: Пользователь не авторизован.');
        return;
    }
    setIsDebugDefinitionsLoading(true);
    setDebugDefinitionsResponse('Загрузка...');
    try {
        const requestUrl = `${apiUrl}/api/definitions`;
        const response = await fetch(requestUrl, {
            method: 'GET',
            mode: 'cors',
            headers: { 
                'Authorization': `Bearer ${token}`,
                'Accept': 'application/json'
            },
        });
        
        const responseBodyText = await response.text();
        
        try {
            const json = JSON.parse(responseBodyText);
            setDebugDefinitionsResponse(JSON.stringify(json, null, 2));
        } catch {
            setDebugDefinitionsResponse(responseBodyText);
        }

    } catch (error: any) {
        setDebugDefinitionsResponse(`Сетевая ошибка: ${error.message}`);
    } finally {
        setIsDebugDefinitionsLoading(false);
    }
  };

  const currentChestDetails = selectedChest ? itemDetails[selectedChest] : null;
  const currentChestCount = selectedChest ? inventory[selectedChest] || 0 : 0;

  const ownedResources = AllResourceTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedReagents = AllReagentTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedBlueprints = AllBlueprintTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedProcessedItems = AllProcessedItemTypes.filter(item => (inventory[item] || 0) > 0);
  const ownedCraftedTools = AllCraftedToolTypes.filter(item => (inventory[item] || 0) > 0);

  const ItemGrid = ({ items, title }: { items: (InventoryItemType | ChestType)[], title?: string }) => (
    items.length > 0 ? (
      <div>
        {title && <h4 className="text-lg font-semibold text-primary mb-4">{title}</h4>}
        <div className="grid grid-cols-4 sm:grid-cols-5 gap-4">
          {items.map(item => (
            <div key={item} className="relative aspect-square flex items-center justify-center">
              <Image
                src={getItemImage(item) || `https://placehold.co/64x64.png`}
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
      <ChestOpeningAnimation 
        chestImageUrl={animatingChest ? getItemImage(animatingChest) : undefined} 
        state={animationState} 
      />
      
      <div className={cn(
        "flex flex-col items-center justify-start min-h-full p-4 text-foreground transition-opacity duration-300",
        animationState !== 'idle' && 'opacity-0 pointer-events-none'
      )}>
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
                        aria-label={`Открыть сундук: ${getItemName(chestType)}`}
                      >
                        <Image
                          src={getItemImage(chestType) ?? `https://placehold.co/64x64.png`}
                          alt={getItemName(chestType)}
                          width={64}
                          height={64}
                          className="w-full h-full object-contain"
                        />
                        <span className="absolute bottom-0 right-0 bg-background/80 backdrop-blur-sm text-foreground text-xs font-bold px-1.5 py-0.5 rounded-full shadow-md">
                          {formatNumber(inventory[chestType] || 0)}
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
                    <ItemGrid items={ownedResources} title="Сырые ресурсы" />
                    <ItemGrid items={ownedProcessedItems} title="Изделия" />
                    <ItemGrid items={ownedBlueprints} title="Чертежи" />
                    {ownedResources.length === 0 && ownedProcessedItems.length === 0 && ownedBlueprints.length === 0 && (
                      <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет ресурсов, изделий или чертежей.</p>
                    )}
                  </div>
                </ScrollArea>
              </TabsContent>
              <TabsContent value="reagents">
                 <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <ItemGrid items={ownedReagents} />
                  {ownedReagents.length === 0 && (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет реагентов.</p>
                  )}
                </ScrollArea>
              </TabsContent>
              <TabsContent value="tools">
                <ScrollArea className="h-96 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                  <ItemGrid items={ownedCraftedTools} />
                  {ownedCraftedTools.length === 0 && (
                    <p className="col-span-full text-center text-muted-foreground py-10">У вас пока нет готовых инструментов.</p>
                  )}
                </ScrollArea>
              </TabsContent>
              <TabsContent value="debug">
                <div className="space-y-6 p-2">
                  <div className="space-y-2">
                    <h3 className="text-lg font-semibold text-primary">Инвентарь игрока</h3>
                    <p className="text-sm text-muted-foreground">Запрашивает предметы, которые есть у текущего пользователя.</p>
                    <Button onClick={handleDebugSync} disabled={isDebugLoading}>
                      {isDebugLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Запросить инвентарь
                    </Button>
                    <ScrollArea className="h-48 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                      <pre className="text-xs whitespace-pre-wrap break-all">
                        <code>
                          {debugResponse || 'Ответ от сервера на запрос инвентаря будет здесь...'}
                        </code>
                      </pre>
                    </ScrollArea>
                  </div>
                  
                  <Separator />

                  <div className="space-y-2">
                    <h3 className="text-lg font-semibold text-primary">Справочник предметов</h3>
                    <p className="text-sm text-muted-foreground">Запрашивает полный список всех возможных предметов и их деталей с сервера.</p>
                    <Button onClick={handleDebugDefinitions} disabled={isDebugDefinitionsLoading}>
                      {isDebugDefinitionsLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Запросить детали всех предметов
                    </Button>
                    <ScrollArea className="h-48 w-full rounded-md border p-4 bg-background/50 shadow-inner">
                      <pre className="text-xs whitespace-pre-wrap break-all">
                        <code>
                          {debugDefinitionsResponse || 'Ответ от сервера на запрос деталей всех предметов будет здесь...'}
                        </code>
                      </pre>
                    </ScrollArea>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>

      {/* Main Chest Info Dialog */}
      <Dialog open={isModalOpen} onOpenChange={(open) => { if (!open) setSelectedChest(null); setIsModalOpen(open); }}>
        <DialogContent className="sm:max-w-[425px]">
          {currentChestDetails && (
            <>
              <DialogHeader>
                <DialogTitle className="text-primary text-2xl">{currentChestDetails.name}</DialogTitle>
                <DialogDescription>
                  {currentChestDetails.description}
                  <br />
                  <span className="font-bold text-foreground">У вас: {currentChestCount.toLocaleString()} шт.</span>
                </DialogDescription>
              </DialogHeader>
              <DialogFooter className="flex-col sm:flex-col sm:space-x-0 gap-2 mt-4">
                <Button onClick={() => handleOpenClick(1)} disabled={currentChestCount < 1 || isLoading}>Открыть</Button>
                <Button onClick={() => handleOpenClick(10)} disabled={currentChestCount < 10 || isLoading}>Открыть (10)</Button>
                <Button onClick={() => handleOpenClick(-1)} disabled={currentChestCount < 1 || isLoading}>Открыть все</Button>
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
              Вы уверены, что хотите открыть "{currentChestDetails?.name}" в количестве {openAmount === -1 ? 'все' : `${openAmount} шт.`}?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isLoading}>Отменить</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmOpen} disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Подтвердить
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Loot Display Dialog */}
      <Dialog open={isLootModalOpen} onOpenChange={(open) => { if (!open) setLastLoot(null); setIsLootModalOpen(open); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-2xl text-primary text-center">Ваша добыча!</DialogTitle>
          </DialogHeader>
          <div className="py-4 max-h-96 overflow-y-auto">
            {lastLoot && lastLoot.length > 0 ? (
                <div className="grid grid-cols-3 gap-4">
                  {lastLoot.map((item, index) => {
                    const itemSlug = getLootItemSlug(item);
                    const imageUrl = itemSlug ? getItemImage(itemSlug) : undefined;

                    if (!itemSlug) return null;

                    return (
                      <div key={index} className="flex flex-col items-center text-center gap-2">
                          <div className="relative w-20 h-20 bg-card/50 rounded-lg p-2">
                              <Image
                                  src={imageUrl ?? `https://placehold.co/64x64.png`}
                                  alt={item.name}
                                  width={64}
                                  height={64}
                                  className="w-full h-full object-contain"
                              />
                              <span className="absolute -bottom-2 -right-2 bg-primary text-primary-foreground text-sm font-bold px-2 py-0.5 rounded-full shadow-md border-2 border-background">
                                  x{item.quantity}
                              </span>
                          </div>
                          <p className="text-xs text-muted-foreground">{item.name}</p>
                      </div>
                    );
                  })}
                </div>
            ) : (
                <p className="text-center text-muted-foreground">Сундук оказался пуст.</p>
            )}
          </div>
          <DialogFooter>
            <Button onClick={() => setIsLootModalOpen(false)}>Отлично!</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

    
