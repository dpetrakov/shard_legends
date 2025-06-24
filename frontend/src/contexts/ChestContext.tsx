
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { ChestType, ChestCounts, ChestContextType } from '@/types/profile';
import { chestDetails } from '@/lib/chest-definitions';

const CHEST_STORAGE_KEY = 'crystalCascadeChestCounts';

const initialChestCounts: ChestCounts = {};

const ChestContext = createContext<ChestContextType | undefined>(undefined);

export const ChestProvider = ({ children }: { children: ReactNode }) => {
  const [chestCounts, setChestCounts] = useState<ChestCounts>(initialChestCounts);

  useEffect(() => {
    const storedCounts = localStorage.getItem(CHEST_STORAGE_KEY);
    if (storedCounts) {
      try {
        const parsedCounts = JSON.parse(storedCounts);
        if (typeof parsedCounts === 'object' && parsedCounts !== null) {
          setChestCounts(parsedCounts);
        } else {
          localStorage.setItem(CHEST_STORAGE_KEY, JSON.stringify(initialChestCounts));
        }
      } catch (error) {
        console.error("Failed to parse chest counts from localStorage", error);
        localStorage.setItem(CHEST_STORAGE_KEY, JSON.stringify(initialChestCounts));
      }
    } else {
       localStorage.setItem(CHEST_STORAGE_KEY, JSON.stringify(initialChestCounts));
    }
  }, []);

  const awardChest = useCallback((chestType: ChestType) => {
    setChestCounts(prevCounts => {
      const newCounts = {
        ...prevCounts,
        [chestType]: (prevCounts[chestType] || 0) + 1,
      };
      localStorage.setItem(CHEST_STORAGE_KEY, JSON.stringify(newCounts));
      return newCounts;
    });
  }, []);

  const spendChests = useCallback((chestType: ChestType, amount: number) => {
    setChestCounts(prevCounts => {
      const currentAmount = prevCounts[chestType] || 0;
      const newCounts = {
        ...prevCounts,
        [chestType]: Math.max(0, currentAmount - amount),
      };
      
      if (newCounts[chestType] === 0) {
        delete newCounts[chestType];
      }

      localStorage.setItem(CHEST_STORAGE_KEY, JSON.stringify(newCounts));
      return newCounts;
    });
  }, []);

  const getChestName = useCallback((chestType: ChestType): string => {
    return chestDetails[chestType]?.name || "Неизвестный сундук";
  }, []);

  return (
    <ChestContext.Provider value={{ chestCounts, awardChest, spendChests, getChestName }}>
      {children}
    </ChestContext.Provider>
  );
};

export const useChests = (): ChestContextType => {
  const context = useContext(ChestContext);
  if (context === undefined) {
    throw new Error('useChests must be used within a ChestProvider');
  }
  return context;
};
