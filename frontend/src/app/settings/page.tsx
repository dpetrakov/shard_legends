
"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Gem, Candy, Skull, Dog, Image as ImageIcon, Settings } from "lucide-react";
import { useIconSet } from "@/contexts/IconSetContext";
import { cn } from "@/lib/utils";

export default function SettingsPage() {
  const { iconSet, setIconSet } = useIconSet();

  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
        <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
            <CardHeader className="text-center">
                <div className="flex justify-center items-center gap-4">
                    <Settings className="w-8 h-8 text-primary" />
                    <CardTitle className="text-3xl font-headline text-primary">Настройки</CardTitle>
                </div>
            </CardHeader>
        </Card>

        {/* Icon Set Selection Section */}
        <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
            <CardHeader>
            <CardTitle className="text-2xl font-headline text-center text-primary">Набор иконок в игре</CardTitle>
            <CardDescription className="text-center text-muted-foreground pt-1">
                Выберите, какие фишки будут в игре "Поиск".
            </CardDescription>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-3">
            <Button
                variant={iconSet === 'classic' ? 'default' : 'outline'}
                className={cn(
                "py-3 text-sm sm:text-base",
                iconSet === 'classic' ? "bg-primary text-primary-foreground" : "border-primary text-primary hover:bg-primary/10"
                )}
                onClick={() => setIconSet('classic')}
            >
                <Gem className="mr-2 h-5 w-5" />
                Классика
            </Button>
            <Button
                variant={iconSet === 'sweets' ? 'default' : 'outline'}
                className={cn(
                "py-3 text-sm sm:text-base",
                iconSet === 'sweets' ? "bg-primary text-primary-foreground" : "border-primary text-primary hover:bg-primary/10"
                )}
                onClick={() => setIconSet('sweets')}
            >
                <Candy className="mr-2 h-5 w-5" />
                Сладости
            </Button>
            <Button
                variant={iconSet === 'gothic' ? 'default' : 'outline'}
                className={cn(
                "py-3 text-sm sm:text-base",
                iconSet === 'gothic' ? "bg-primary text-primary-foreground" : "border-primary text-primary hover:bg-primary/10"
                )}
                onClick={() => setIconSet('gothic')}
            >
                <Skull className="mr-2 h-5 w-5" />
                Готика
            </Button>
            <Button
                variant={iconSet === 'animals' ? 'default' : 'outline'}
                className={cn(
                "py-3 text-sm sm:text-base",
                iconSet === 'animals' ? "bg-primary text-primary-foreground" : "border-primary text-primary hover:bg-primary/10"
                )}
                onClick={() => setIconSet('animals')}
            >
                <Dog className="mr-2 h-5 w-5" />
                Животные
            </Button>
            <Button
                variant={iconSet === 'in-match3' ? 'default' : 'outline'}
                className={cn(
                "py-3 text-sm sm:text-base",
                iconSet === 'in-match3' ? "bg-primary text-primary-foreground" : "border-primary text-primary hover:bg-primary/10"
                )}
                onClick={() => setIconSet('in-match3')}
            >
                <ImageIcon className="mr-2 h-5 w-5" />
                IN-match3
            </Button>
            </CardContent>
        </Card>
    </div>
  );
}
