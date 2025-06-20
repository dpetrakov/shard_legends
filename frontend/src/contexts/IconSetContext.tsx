
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { IconSet, IconSetContextType as ImportedIconSetContextType } from '@/types/icon-set';
import type { ShardIcon } from '@/types/shard-legends';
import { ICON_SETS, CLASSIC_SHARD_ICONS } from '@/components/shard-legends/shard-definitions';

const ICON_SET_STORAGE_KEY = 'shardLegendsIconSet';

const IconSetContext = createContext<ImportedIconSetContextType | undefined>(undefined);

export const IconSetProvider = ({ children }: { children: ReactNode }) => {
  const [iconSet, setIconSetState] = useState<IconSet>('classic');

  useEffect(() => {
    const storedIconSet = localStorage.getItem(ICON_SET_STORAGE_KEY) as IconSet | null;
    const initialIconSet = storedIconSet || 'classic';
    setIconSetState(initialIconSet);
  }, []);

  const setIconSet = useCallback((newIconSet: IconSet) => {
    localStorage.setItem(ICON_SET_STORAGE_KEY, newIconSet);
    setIconSetState(newIconSet);
  }, []);

  const getActiveIconList = useCallback((): ShardIcon[] => {
    return ICON_SETS[iconSet] || CLASSIC_SHARD_ICONS; // Fallback to classic
  }, [iconSet]);

  return (
    <IconSetContext.Provider value={{ iconSet, setIconSet, getActiveIconList }}>
      {children}
    </IconSetContext.Provider>
  );
};

export const useIconSet = (): ImportedIconSetContextType => {
  const context = useContext(IconSetContext);
  if (context === undefined) {
    throw new Error('useIconSet must be used within an IconSetProvider');
  }
  return context;
};
