
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { InventoryContextType, LootResult, InventoryItemType, CraftedToolType, ItemDetails } from '@/types/inventory';
import { itemNames } from '@/lib/item-definitions';
import { useAuth } from './AuthContext';
import type { ChestType } from '@/types/profile';

const INVENTORY_STORAGE_KEY = 'crystalCascadeInventory';
const ITEM_DETAILS_STORAGE_KEY = 'crystalCascadeItemDetails';

const initialInventory: LootResult = {};

const InventoryContext = createContext<InventoryContextType | undefined>(undefined);

export const InventoryProvider = ({ children }: { children: ReactNode }) => {
  const [inventory, setInventory] = useState<LootResult>(initialInventory);
  const [itemDetails, setItemDetails] = useState<Record<string, ItemDetails>>({});
  const { isAuthenticated, token } = useAuth();
  const apiUrl = 'https://dev-forly.slcw.dimlight.online';

  // Load inventory from localStorage on mount
  useEffect(() => {
    const storedInventory = localStorage.getItem(INVENTORY_STORAGE_KEY);
    if (storedInventory) {
      try {
        setInventory(JSON.parse(storedInventory));
      } catch (error) {
        console.error("Failed to parse inventory from localStorage", error);
      }
    }
    const storedItemDetails = localStorage.getItem(ITEM_DETAILS_STORAGE_KEY);
     if (storedItemDetails) {
      try {
        setItemDetails(JSON.parse(storedItemDetails));
      } catch (error) {
        console.error("Failed to parse item details from localStorage", error);
      }
    }
  }, []);

  // Persist inventory to localStorage when it changes
  useEffect(() => {
    localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(inventory));
    localStorage.setItem(ITEM_DETAILS_STORAGE_KEY, JSON.stringify(itemDetails));
  }, [inventory, itemDetails]);
  
  const syncWithServer = useCallback(async () => {
    if (!isAuthenticated || !token) {
        return;
    }

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

      if (!response.ok) {
          console.error("Failed to sync inventory, error response:", { status: response.status, body: responseBodyText });
          return;
      }
      if (!responseBodyText) {
          console.log("Inventory sync response was empty, assuming empty inventory.");
          setInventory({});
          setItemDetails({});
          return;
      }
      
      let data;
      try {
        data = JSON.parse(responseBodyText);
      } catch (e) {
        console.error("Failed to parse inventory JSON response:", { status: response.status, body: responseBodyText });
        return;
      }

      if (!Array.isArray(data.items)) {
        console.error("Failed to sync inventory: 'items' array not found in server response", data);
        return;
      }

      const newInventory: LootResult = {};
      const newItemDetails: Record<string, ItemDetails> = {};
      const baseImageUrl = apiUrl.replace('/api', '');
      
      for (const serverItem of data.items) {
        let slug: InventoryItemType | ChestType | null = null;
        const { item_id, item_class, item_type, quality_level, quantity, name, description, image_url } = serverItem;
    
        if (!item_class || !item_type || typeof quantity !== 'number') {
            console.warn('Skipping invalid item from server:', serverItem);
            continue;
        }
    
        switch (item_class) {
            case 'chests':
                if (item_type === 'blueprint' || item_type === 'blueprint_chest') {
                    slug = 'blueprint';
                } else if (quality_level) {
                    const baseType = item_type.replace('_chest', '');
                    slug = `${baseType}_${quality_level}` as ChestType;
                }
                break;
            case 'resources':
            case 'reagents':
            case 'processed_items':
            case 'blueprints':
                slug = item_type as InventoryItemType;
                break;
            case 'tools':
                if (quality_level) {
                    slug = `${quality_level}_${item_type}` as CraftedToolType;
                }
                break;
        }
    
        if (slug) {
            newInventory[slug] = (newInventory[slug] || 0) + quantity;
            // Always update details to get the latest info, including image_url
            newItemDetails[slug] = {
                id: item_id,
                slug: slug,
                name: name,
                description: description,
                imageUrl: image_url ? `${baseImageUrl}${image_url}` : undefined,
            };
        } else {
            console.warn('Could not determine slug for server item:', serverItem);
        }
      }
      
      setInventory(newInventory);
      setItemDetails(newItemDetails);
      console.log("Inventory synced with server successfully.");

    } catch (error: any) {
      console.error("Network error syncing inventory:", error);
       if (error.name === 'TypeError' && error.message.includes('Failed to fetch')) {
            console.error('This might be a CORS issue. Please check the server CORS policy.');
        }
    }
  }, [isAuthenticated, token, apiUrl]);
  
  useEffect(() => {
      if (isAuthenticated) {
          syncWithServer();
      }
  }, [isAuthenticated, syncWithServer]);

  const addItems = useCallback((itemsToAdd: LootResult) => {
    setInventory(prevInventory => {
      const newInventory = { ...prevInventory };
      for (const key in itemsToAdd) {
          const itemKey = key as InventoryItemType | ChestType;
          const amountToAdd = itemsToAdd[itemKey] || 0;
          newInventory[itemKey] = (newInventory[itemKey] || 0) + amountToAdd;
      }
      return newInventory;
    });
  }, []);

  const spendItems = useCallback((itemsToSpend: LootResult): boolean => {
    const canAfford = Object.keys(itemsToSpend).every(key => {
        const itemKey = key as InventoryItemType | ChestType;
        return (inventory[itemKey] || 0) >= (itemsToSpend[itemKey] || 0);
    });

    if (canAfford) {
        setInventory(prevInventory => {
            const newInventory = { ...prevInventory };
            for (const key in itemsToSpend) {
                const itemKey = key as InventoryItemType | ChestType;
                newInventory[itemKey] = (newInventory[itemKey] || 0) - (itemsToSpend[itemKey] || 0);
                if (newInventory[itemKey] <= 0) {
                    delete newInventory[itemKey];
                }
            }
            return newInventory;
        });
        return true;
    }
    return false;
  }, [inventory]);

  const getItemName = useCallback((item: InventoryItemType | ChestType): string => {
    return itemDetails[item]?.name || itemNames[item as InventoryItemType] || "Неизвестный предмет";
  }, [itemDetails]);

  const getItemImage = useCallback((item: InventoryItemType | ChestType): string | undefined => {
      return itemDetails[item]?.imageUrl;
  }, [itemDetails]);

  return (
    <InventoryContext.Provider value={{ inventory, itemDetails, addItems, spendItems, getItemName, getItemImage, syncWithServer }}>
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
