
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { RefiningContextType, ActiveProcess, RefiningRecipe } from '@/types/refining';
import { useInventory } from './InventoryContext';
import type { ActiveCraftingProcess } from '@/types/crafting';

const REFINING_STORAGE_KEY = 'crystalCascadeRefiningProcesses';
const CRAFTING_STORAGE_KEY = 'crystalCascadeCraftingProcesses';
const MAX_FACTORY_SLOTS = 2;

const RefiningContext = createContext<RefiningContextType | undefined>(undefined);

export const RefiningProvider = ({ children }: { children: ReactNode }) => {
    const { inventory, spendItems, addItems } = useInventory();
    const [activeProcesses, setActiveProcesses] = useState<ActiveProcess[]>([]);
    const [isLoaded, setIsLoaded] = useState(false);

    useEffect(() => {
        const stored = localStorage.getItem(REFINING_STORAGE_KEY);
        if (stored) {
            try {
                const parsed = JSON.parse(stored) as ActiveProcess[];
                setActiveProcesses(parsed);
            } catch {
                setActiveProcesses([]);
            }
        }
        setIsLoaded(true);
    }, []);

    useEffect(() => {
        if (isLoaded) {
            localStorage.setItem(REFINING_STORAGE_KEY, JSON.stringify(activeProcesses));
        }
    }, [activeProcesses, isLoaded]);


    const startProcess = useCallback((recipe: RefiningRecipe, quantity: number): boolean => {
        const storedCrafting = localStorage.getItem(CRAFTING_STORAGE_KEY);
        const craftingProcesses: ActiveCraftingProcess[] = storedCrafting ? JSON.parse(storedCrafting) : [];

        if (activeProcesses.length + craftingProcesses.length >= MAX_FACTORY_SLOTS) {
            alert("Все слоты кузницы заняты.");
            return false;
        }

        if (quantity <= 0) {
            alert("Укажите корректное количество.");
            return false;
        }

        const itemsToSpend = recipe.input.reduce((acc, current) => {
            acc[current.item] = current.quantity * quantity;
            return acc;
        }, {} as Record<string, number>);

        const canAfford = Object.keys(itemsToSpend).every(itemKey => {
            return (inventory[itemKey as keyof typeof inventory] || 0) >= itemsToSpend[itemKey];
        });

        if (!canAfford) {
            alert("Недостаточно ресурсов для запуска процесса.");
            return false;
        }

        if (spendItems(itemsToSpend)) {
            const newProcess: ActiveProcess = {
                id: Date.now().toString(),
                recipe,
                quantity,
                endTime: Date.now() + recipe.durationSeconds * quantity * 1000,
            };
            setActiveProcesses(prev => [...prev, newProcess]);
            return true;
        } else {
            alert("Произошла ошибка при списании ресурсов.");
            return false;
        }

    }, [activeProcesses.length, inventory, spendItems]);


    const claimProcess = useCallback((processId: string) => {
        const process = activeProcesses.find(p => p.id === processId);
        if (!process || process.endTime > Date.now()) {
            return;
        }
        
        const itemsToAdd = {
            [process.recipe.output.item]: process.recipe.output.quantity * process.quantity
        };

        addItems(itemsToAdd);
        setActiveProcesses(prev => prev.filter(p => p.id !== processId));
    }, [activeProcesses, addItems]);

    const processSlots = {
        current: activeProcesses.length,
        max: MAX_FACTORY_SLOTS,
    };

    return (
        <RefiningContext.Provider value={{ activeProcesses, startProcess, claimProcess, processSlots }}>
            {children}
        </RefiningContext.Provider>
    );
};


export const useRefining = (): RefiningContextType => {
    const context = useContext(RefiningContext);
    if (context === undefined) {
        throw new Error('useRefining must be used within a RefiningProvider');
    }
    return context;
};
