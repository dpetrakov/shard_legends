
"use client";

import type { RewardContextType } from '@/types/rewards';
import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';

const REWARD_STORAGE_KEY = 'shardLegendsRewardPoints';

const RewardContext = createContext<RewardContextType | undefined>(undefined);

export const RewardProvider = ({ children }: { children: ReactNode }) => {
  const [totalRewardPoints, setTotalRewardPoints] = useState<number>(0);

  useEffect(() => {
    const storedPoints = localStorage.getItem(REWARD_STORAGE_KEY);
    if (storedPoints) {
      try {
        const parsedPoints = parseInt(storedPoints, 10);
        if (!isNaN(parsedPoints)) {
          setTotalRewardPoints(parsedPoints);
        } else {
          localStorage.setItem(REWARD_STORAGE_KEY, '0');
        }
      } catch (error) {
        console.error("Failed to parse reward points from localStorage", error);
        localStorage.setItem(REWARD_STORAGE_KEY, '0');
      }
    } else {
      localStorage.setItem(REWARD_STORAGE_KEY, '0');
    }
  }, []);

  const addRewardPoints = useCallback((points: number) => {
    setTotalRewardPoints(prevPoints => {
      const newTotal = prevPoints + points;
      localStorage.setItem(REWARD_STORAGE_KEY, newTotal.toString());
      return newTotal;
    });
  }, []);

  const spendRewardPoints = useCallback((points: number) => {
    setTotalRewardPoints(prevPoints => {
      const newTotal = Math.max(0, prevPoints - points); // Ensure points don't go negative
      localStorage.setItem(REWARD_STORAGE_KEY, newTotal.toString());
      return newTotal;
    });
  }, []);

  return (
    <RewardContext.Provider value={{ totalRewardPoints, addRewardPoints, spendRewardPoints }}>
      {children}
    </RewardContext.Provider>
  );
};

export const useRewards = (): RewardContextType => {
  const context = useContext(RewardContext);
  if (context === undefined) {
    throw new Error('useRewards must be used within a RewardProvider');
  }
  return context;
};
