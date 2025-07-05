
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import { Axe, Pickaxe, Shovel, Wheat } from 'lucide-react';
import type { MiningContextType, EquippedTools, ToolType, EquippedTool, ActiveMiningProcess, MiningLocation } from '@/types/mining';
import type { CraftedToolType } from '@/types/inventory';
import { useInventory } from './InventoryContext';

const MINING_STORAGE_KEY = 'crystalCascadeMiningSetup';
const MINING_PROCESS_KEY = 'crystalCascadeMiningProcess';

const initialEquippedTools: EquippedTools = {};

const MiningContext = createContext<MiningContextType | undefined>(undefined);

const toolMeta: Record<ToolType, { name: string; icon: React.ComponentType<{ className?: string }> }> = {
    axe: { name: 'Топор', icon: Axe },
    pickaxe: { name: 'Кирка', icon: Pickaxe },
    shovel: { name: 'Лопата', icon: Shovel },
    sickle: { name: 'Серп', icon: Wheat },
};

export const MiningProvider = ({ children }: { children: ReactNode }) => {
    const { addItems, spendItems } = useInventory();
    const [equippedTools, setEquippedTools] = useState<EquippedTools>(initialEquippedTools);
    const [activeProcess, setActiveProcess] = useState<ActiveMiningProcess | null>(null);
    const [isLoaded, setIsLoaded] = useState(false);

    useEffect(() => {
        const storedTools = localStorage.getItem(MINING_STORAGE_KEY);
        if (storedTools) {
            try { setEquippedTools(JSON.parse(storedTools)); } catch {}
        }
        const storedProcess = localStorage.getItem(MINING_PROCESS_KEY);
        if (storedProcess) {
             try { 
                const parsed = JSON.parse(storedProcess);
                // Basic validation to prevent loading invalid data
                if (parsed && typeof parsed === 'object' && 'endTime' in parsed) {
                    setActiveProcess(parsed);
                } else {
                    setActiveProcess(null);
                }
             } catch {
                setActiveProcess(null);
             }
        }
        setIsLoaded(true);
    }, []);

    useEffect(() => {
        if (isLoaded) {
            localStorage.setItem(MINING_STORAGE_KEY, JSON.stringify(equippedTools));
            localStorage.setItem(MINING_PROCESS_KEY, JSON.stringify(activeProcess));
        }
    }, [equippedTools, activeProcess, isLoaded]);

    const getToolName = useCallback((toolType: ToolType) => toolMeta[toolType].name, []);

    const equipTool = useCallback((toolType: ToolType, item: CraftedToolType): boolean => {
        if (equippedTools[toolType]) {
            alert(`Слот для инструмента "${getToolName(toolType)}" уже занят.`);
            return false;
        }

        const success = spendItems({ [item]: 1 });
        if (success) {
            const newTool: EquippedTool = {
                item,
                stats: { strength: 10, speed: 10, luck: 10, durability: 100 }, // Hardcoded stats
            };
            setEquippedTools(prev => ({ ...prev, [toolType]: newTool }));
            return true;
        }
        return false;
    }, [equippedTools, spendItems, getToolName]);

    const unequipTool = useCallback((toolType: ToolType): boolean => {
        const toolToUnequip = equippedTools[toolType];
        if (!toolToUnequip) return false;

        addItems({ [toolToUnequip.item]: 1 });
        setEquippedTools(prev => {
            const newTools = { ...prev };
            delete newTools[toolType];
            return newTools;
        });
        return true;
    }, [equippedTools, addItems]);

    const startMining = useCallback((location: MiningLocation): boolean => {
        if (activeProcess) {
            alert("Вы уже в процессе добычи.");
            return false;
        }

        const requiredToolType: ToolType = location === 'cave' ? 'pickaxe' : 'axe';
        const tool = equippedTools[requiredToolType];

        if (!tool) {
            alert(`Для этой локации требуется ${getToolName(requiredToolType)}.`);
            return false;
        }

        const newProcess: ActiveMiningProcess = {
            location,
            endTime: Date.now() + 1 * 60 * 60 * 1000, // 1 hour
        };
        
        // Decrease durability
        const newStats = { ...tool.stats, durability: tool.stats.durability - 1 };
        const updatedTool = { ...tool, stats: newStats };

        setEquippedTools(prev => ({
            ...prev,
            [requiredToolType]: updatedTool,
        }));

        setActiveProcess(newProcess);
        return true;
    }, [activeProcess, equippedTools, getToolName]);

    const claimMining = useCallback(() => {
        if (!activeProcess || activeProcess.endTime > Date.now()) {
            return null;
        }

        const location = activeProcess.location;
        // Placeholder rewards
        const rewards = location === 'cave' 
            ? { stone: 100, ore: 5 } 
            : { wood: 100, diamond: 1 };
        
        addItems(rewards);
        const claimedProcess = activeProcess;
        setActiveProcess(null);
        
        return { location: claimedProcess.location, rewards };
    }, [activeProcess, addItems]);

    const getToolIcon = useCallback((toolType: ToolType) => toolMeta[toolType].icon, []);


    return (
        <MiningContext.Provider value={{ equippedTools, equipTool, unequipTool, getToolIcon, getToolName, activeProcess, startMining, claimMining }}>
            {children}
        </MiningContext.Provider>
    );
};

export const useMining = (): MiningContextType => {
    const context = useContext(MiningContext);
    if (context === undefined) {
        throw new Error('useMining must be used within a MiningProvider');
    }
    return context;
};
