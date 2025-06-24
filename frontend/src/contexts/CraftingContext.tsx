
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { CraftingContextType, ActiveCraftingProcess, CraftingRecipeInfo } from '@/types/crafting';
import { useInventory } from './InventoryContext';
import { TOOL_QUALITIES, CRAFTING_COST_PER_ITEM, CRAFTING_DURATION_SECONDS_PER_ITEM, getCraftedToolId } from '@/lib/crafting-definitions';

const CRAFTING_STORAGE_KEY = 'crystalCascadeCraftingProcesses';
const MAX_CRAFTING_SLOTS = 2;

const CraftingContext = createContext<CraftingContextType | undefined>(undefined);

export const CraftingProvider = ({ children }: { children: ReactNode }) => {
    const { inventory, spendItems, addItems, getItemName } = useInventory();
    const [activeProcesses, setActiveProcesses] = useState<ActiveCraftingProcess[]>([]);
    const [isLoaded, setIsLoaded] = useState(false);

    useEffect(() => {
        const stored = localStorage.getItem(CRAFTING_STORAGE_KEY);
        if (stored) {
            try {
                const parsed = JSON.parse(stored) as ActiveCraftingProcess[];
                setActiveProcesses(parsed);
            } catch {
                setActiveProcesses([]);
            }
        }
        setIsLoaded(true);
    }, []);

    useEffect(() => {
        if (isLoaded) {
            localStorage.setItem(CRAFTING_STORAGE_KEY, JSON.stringify(activeProcesses));
        }
    }, [activeProcesses, isLoaded]);

    const startProcess = useCallback((recipeInfo: CraftingRecipeInfo, quantity: number): boolean => {
        if (activeProcesses.length >= MAX_CRAFTING_SLOTS) {
            alert("Все слоты крафта заняты.");
            return false;
        }
        if (!recipeInfo.tool || !recipeInfo.quality || quantity <= 0) {
            alert("Выберите инструмент, качество и корректное количество.");
            return false;
        }
        
        const qualityInfo = TOOL_QUALITIES.find(q => q.id === recipeInfo.quality);
        if (!qualityInfo) {
             alert("Выбрано некорректное качество.");
             return false;
        }

        const itemsToSpend = {
            [recipeInfo.tool]: quantity, // Blueprints
            [qualityInfo.material]: CRAFTING_COST_PER_ITEM * quantity, // Processed materials
        };

        const canAfford = Object.keys(itemsToSpend).every(itemKey => {
            const requiredAmount = itemsToSpend[itemKey as keyof typeof itemsToSpend];
            return (inventory[itemKey as keyof typeof inventory] || 0) >= requiredAmount;
        });

        if (!canAfford) {
            alert("Недостаточно ресурсов для запуска крафта.");
            return false;
        }

        if (spendItems(itemsToSpend)) {
            const outputItem = getCraftedToolId(recipeInfo.tool, recipeInfo.quality);
            const newProcess: ActiveCraftingProcess = {
                id: Date.now().toString(),
                recipe: recipeInfo,
                quantity,
                endTime: Date.now() + CRAFTING_DURATION_SECONDS_PER_ITEM * quantity * 1000,
                outputItem: outputItem,
                outputItemName: getItemName(outputItem),
            };
            setActiveProcesses(prev => [...prev, newProcess]);
            return true;
        } else {
            alert("Произошла ошибка при списании ресурсов.");
            return false;
        }

    }, [activeProcesses.length, inventory, spendItems, getItemName]);

    const claimProcess = useCallback((processId: string) => {
        const process = activeProcesses.find(p => p.id === processId);
        if (!process || process.endTime > Date.now()) {
            return;
        }
        
        const itemsToAdd = { [process.outputItem]: process.quantity };

        addItems(itemsToAdd);
        setActiveProcesses(prev => prev.filter(p => p.id !== processId));
    }, [activeProcesses, addItems]);

    const processSlots = {
        current: activeProcesses.length,
        max: MAX_CRAFTING_SLOTS,
    };

    return (
        <CraftingContext.Provider value={{ activeProcesses, startProcess, claimProcess, processSlots }}>
            {children}
        </CraftingContext.Provider>
    );
};

export const useCrafting = (): CraftingContextType => {
    const context = useContext(CraftingContext);
    if (context === undefined) {
        throw new Error('useCrafting must be used within a CraftingProvider');
    }
    return context;
};
