
"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { useAuth } from './AuthContext';
import { useInventory } from './InventoryContext';
import type { ProductionContextType, ProductionRecipe, ProductionTask } from '@/types/production';
import { AllProcessedItemTypes } from '@/types/inventory';
import { useToast } from '@/hooks/use-toast';

const ProductionContext = createContext<ProductionContextType | undefined>(undefined);

export const ProductionProvider = ({ children }: { children: ReactNode }) => {
    const { token, isAuthenticated } = useAuth();
    const { syncWithServer: syncInventory } = useInventory();
    const { toast } = useToast();
    const apiUrl = 'https://dev-forly.slcw.dimlight.online';

    const [recipes, setRecipes] = useState<ProductionRecipe[]>([]);
    const [tasks, setTasks] = useState<ProductionTask[]>([]);
    const [isLoading, setIsLoading] = useState(false);

    const getRecipeCategory = useCallback((recipe: ProductionRecipe): 'refining' | 'crafting' => {
        const outputItemSlug = recipe.output_items[0]?.item_slug;
        if (outputItemSlug && AllProcessedItemTypes.includes(outputItemSlug as any)) {
            return 'refining';
        }
        return 'crafting';
    }, []);

    const fetchData = useCallback(async () => {
        if (!isAuthenticated || !token) return;
        setIsLoading(true);

        try {
            const headers = { 
                'Authorization': `Bearer ${token}`,
                'Accept': 'application/json',
            };

            const [recipesRes, queueRes] = await Promise.all([
                fetch(`${apiUrl}/api/production/recipes`, { mode: 'cors', headers }),
                fetch(`${apiUrl}/api/production/factory/queue`, { mode: 'cors', headers })
            ]);

            if (!recipesRes.ok) {
                console.error("Failed to fetch recipes:", await recipesRes.text());
                throw new Error('Failed to fetch recipes');
            }
             if (!queueRes.ok) {
                console.error("Failed to fetch queue:", await queueRes.text());
                throw new Error('Failed to fetch queue');
            }

            const recipesData = await recipesRes.json();
            const queueData = await queueRes.json();
            
            setRecipes(recipesData.recipes || []);
            setTasks(queueData.tasks || []);

        } catch (error) {
            console.error("Error fetching production data:", error);
            toast({
                variant: "destructive",
                title: "Ошибка",
                description: "Не удалось загрузить данные для кузницы.",
            });
        } finally {
            setIsLoading(false);
        }
    }, [isAuthenticated, token, toast, apiUrl]);

    const startProduction = useCallback(async (recipeId: string, quantity: number): Promise<boolean> => {
        if (!isAuthenticated || !token) return false;
        
        try {
            const response = await fetch(`${apiUrl}/api/production/factory/start`, {
                method: 'POST',
                mode: 'cors',
                headers: { 
                    'Authorization': `Bearer ${token}`, 
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({ recipe_id: recipeId, execution_count: quantity }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to start production');
            }
            
            toast({ title: "Производство запущено!", description: "Задача добавлена в очередь." });
            await fetchData(); // Refresh data
            await syncInventory(); // Refresh inventory
            return true;
        } catch (error: any) {
            console.error("Error starting production:", error);
            toast({ variant: "destructive", title: "Ошибка", description: error.message || "Не удалось запустить производство." });
            return false;
        }
    }, [isAuthenticated, token, fetchData, syncInventory, toast, apiUrl]);
    
    const claimCompleted = useCallback(async () => {
        if (!isAuthenticated || !token) return;

        try {
            const response = await fetch(`${apiUrl}/api/production/factory/claim`, {
                method: 'POST',
                mode: 'cors',
                headers: { 
                    'Authorization': `Bearer ${token}`, 
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({}),
            });

            if (!response.ok) {
                 const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to claim items');
            }

            const claimedData = await response.json();
            const claimedCount = claimedData.claimed_tasks_count || 0;

            if (claimedCount > 0) {
                 toast({ title: "Успех!", description: `Вы забрали ${claimedCount} готовых заказов.` });
            } else {
                 toast({ title: "Нечего забирать", description: "Нет готовых заказов." });
            }

            await fetchData();
            await syncInventory();

        } catch (error: any) {
            console.error("Error claiming items:", error);
            toast({ variant: "destructive", title: "Ошибка", description: error.message || "Не удалось забрать предметы." });
        }
    }, [isAuthenticated, token, fetchData, syncInventory, toast, apiUrl]);

    const getRecipeById = useCallback((recipeId: string) => {
        return recipes.find(r => r.id === recipeId);
    }, [recipes]);
    
    useEffect(() => {
        if (isAuthenticated) {
            fetchData();
        }
    }, [isAuthenticated, fetchData]);


    const value = { recipes, tasks, isLoading, fetchData, startProduction, claimCompleted, getRecipeById };

    return (
        <ProductionContext.Provider value={value}>
            {children}
        </ProductionContext.Provider>
    );
};

export const useProduction = (): ProductionContextType => {
    const context = useContext(ProductionContext);
    if (context === undefined) {
        throw new Error('useProduction must be used within a ProductionProvider');
    }
    return context;
};
