
import type { ChestType } from '@/types/profile';

export const allChestTypes: ChestType[] = [
    'resource_small', 'resource_medium', 'resource_large',
    'reagent_small', 'reagent_medium', 'reagent_large',
    'booster_small', 'booster_medium', 'booster_large',
    'blueprint'
];

export const chestDetails: Record<ChestType, { name: string; hint: string; description: string }> = {
    resource_small: { name: "Малый сундук ресурсов", hint: "small resource chest", description: "Содержит небольшое количество базовых ресурсов." },
    resource_medium: { name: "Средний сундук ресурсов", hint: "medium resource chest", description: "Содержит среднее количество базовых и редких ресурсов." },
    resource_large: { name: "Большой сундук ресурсов", hint: "large resource chest", description: "Содержит большое количество ценных и редких ресурсов." },
    reagent_small: { name: "Малый сундук реагентов", hint: "small reagent chest", description: "Содержит несколько обычных реагентов для крафта." },
    reagent_medium: { name: "Средний сундук реагентов", hint: "medium reagent chest", description: "Содержит разнообразные реагенты, включая редкие." },
    reagent_large: { name: "Большой сундук реагентов", hint: "large reagent chest", description: "Содержит множество редких и эпических реагентов." },
    booster_small: { name: "Малый сундук ускорителей", hint: "small booster chest", description: "Содержит небольшие ускорители для ваших процессов." },
    booster_medium: { name: "Средний сундук ускорителей", hint: "medium booster chest", description: "Содержит значительные ускорители, экономящие ваше время." },
    booster_large: { name: "Большой сундук ускорителей", hint: "large booster chest", description: "Содержит мощные ускорители, которые сильно ускорят ваш прогресс." },
    blueprint: { name: "Сундук-чертеж", hint: "blueprint chest", description: "Содержит уникальный чертеж для создания нового предмета или улучшения." },
};
