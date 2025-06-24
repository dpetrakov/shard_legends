
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { InventoryContextType, LootResult, InventoryItemType } from '@/types/inventory';
import { itemNames } from '@/lib/item-definitions';

const INVENTORY_STORAGE_KEY = 'crystalCascadeInventory';

const initialInventory: LootResult = {};

const InventoryContext = createContext<InventoryContextType | undefined>(undefined);

export const InventoryProvider = ({ children }: { children: ReactNode }) => {
  const [inventory, setInventory] = useState<LootResult>(initialInventory);

  useEffect(() => {
    const storedInventory = localStorage.getItem(INVENTORY_STORAGE_KEY);
    if (storedInventory) {
      try {
        const parsed = JSON.parse(storedInventory);
        if (typeof parsed === 'object' && parsed !== null) {
          setInventory(parsed);
        }
      } catch (error) {
        console.error("Failed to parse inventory from localStorage", error);
        localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(initialInventory));
      }
    } else {
       localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(initialInventory));
    }
  }, []);

  const addItems = useCallback((itemsToAdd: LootResult) => {
    setInventory(prevInventory => {
      const newInventory = { ...prevInventory };
      for (const key in itemsToAdd) {
          const itemKey = key as InventoryItemType;
          const amountToAdd = itemsToAdd[itemKey] || 0;
          newInventory[itemKey] = (newInventory[itemKey] || 0) + amountToAdd;
      }
      localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(newInventory));
      return newInventory;
    });
  }, []);

  const spendItems = useCallback((itemsToSpend: LootResult): boolean => {
    let canAfford = true;
    for (const key in itemsToSpend) {
        const itemKey = key as InventoryItemType;
        const amountToSpend = itemsToSpend[itemKey] || 0;
        if ((inventory[itemKey] || 0) < amountToSpend) {
            canAfford = false;
            break;
        }
    }

    if (canAfford) {
        setInventory(prevInventory => {
            const newInventory = { ...prevInventory };
            for (const key in itemsToSpend) {
                const itemKey = key as InventoryItemType;
                const amountToSpend = itemsToSpend[itemKey] || 0;
                newInventory[itemKey] = (newInventory[itemKey] || 0) - amountToSpend;
                if (newInventory[itemKey] <= 0) {
                    delete newInventory[itemKey];
                }
            }
            localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(newInventory));
            return newInventory;
        });
        return true;
    } else {
        return false;
    }
  }, [inventory]);

  const getItemName = useCallback((item: InventoryItemType): string => {
    return itemNames[item] || "Неизвестный предмет";
  }, []);

  return (
    <InventoryContext.Provider value={{ inventory, addItems, spendItems, getItemName }}>
      {children}
    </InventoryContext.Provider>
  );
};

export const useInventory = (): InventoryContextType => {
  const context = useContext(InventoryContext);
  if (context === undefined) {
    throw new Error('useInventory must be used within a InventoryProvider');
  }
  return context;
};
