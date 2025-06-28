import type { RefiningRecipe } from '@/types/refining';

export const REFINING_RECIPES: RefiningRecipe[] = [
    {
        output: { item: 'wood_plank', quantity: 1 },
        input: [
            { item: 'wood', quantity: 100 },
            { item: 'disc', quantity: 4 },
        ],
        durationSeconds: 60,
    },
    {
        output: { item: 'stone_block', quantity: 1 },
        input: [
            { item: 'stone', quantity: 100 },
            { item: 'abrasive', quantity: 4 },
        ],
        durationSeconds: 60,
    },
    {
        output: { item: 'metal_ingot', quantity: 1 },
        input: [
            { item: 'ore', quantity: 100 },
            { item: 'inductor', quantity: 4 },
        ],
        durationSeconds: 60,
    },
    {
        output: { item: 'cut_diamond', quantity: 1 },
        input: [
            { item: 'diamond', quantity: 100 },
            { item: 'paste', quantity: 4 },
        ],
        durationSeconds: 60,
    },
];
