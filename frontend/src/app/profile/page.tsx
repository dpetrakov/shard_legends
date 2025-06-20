
"use client";

import { useState } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { api, ApiError } from "@/lib/api";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useChests } from "@/contexts/ChestContext";
import type { ChestType } from "@/types/profile";
import { User, Gem, Candy, Skull as SkullIcon, Dog as DogIcon, Image as ImageIcon, Wifi, Loader2 } from "lucide-react";
import NextImage from 'next/image';
import { useIconSet } from "@/contexts/IconSetContext";
import { cn } from "@/lib/utils";

const chestDisplayOrder: ChestType[] = ['small', 'medium', 'large'];
const chestVisualData: Record<ChestType, { name: string; hint: string }> = {
  small: { name: "Малый", hint: "small treasure" },
  medium: { name: "Средний", hint: "medium treasure" },
  large: { name: "Большой", hint: "large treasure" }
};

export default function ProfilePage() {
  const { chestCounts } = useChests();
  const { iconSet, setIconSet } = useIconSet();

  const [isPingModalOpen, setIsPingModalOpen] = useState(false);
  const [pingModalMessage, setPingModalMessage] = useState("");
  const [isPinging, setIsPinging] = useState(false);

  const handlePingServer = async () => {
    setIsPinging(true);
    setPingModalMessage("Pinging server...");
    setIsPingModalOpen(true);
    try {
      const data = await api.getJson<{message: string}>('/api/ping');
      setPingModalMessage(`Ответ сервера: ${data.message}`);
    } catch (error) {
      if (error instanceof ApiError) {
        setPingModalMessage(`Ошибка: ${error.message} (Код: ${error.status})`);
      } else if (error instanceof Error) {
        setPingModalMessage(`Ошибка: ${error.message}`);
      } else {
        setPingModalMessage(`Произошла неизвестная ошибка.`);
      }
    } finally {
      setIsPinging(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-start min-h-screen p-4 pt-6 space-y-6 text-foreground pb-20">
      {/* User Info Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardContent className="pt-6 flex flex-col items-center space-y-4">
          <Avatar className="w-24 h-24 border-2 border-primary">
            <AvatarImage src="https://placehold.co/100x100.png" alt="User Avatar" data-ai-hint="cyborg avatar" />
            <AvatarFallback>
              <User className="w-12 h-12" />
            </AvatarFallback>
          </Avatar>
          <span className="text-2xl font-headline">Имя пользователя</span>
          <Button variant="outline" className="border-primary text-primary hover:bg-primary/10 hover:text-primary-foreground">
            Подключить кошелек
          </Button>
        </CardContent>
      </Card>

      {/* Chests Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl font-headline text-center text-primary">Мои Сундуки</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-3 gap-4">
          {chestDisplayOrder.map((chestType) => (
            <div key={chestType} className="flex flex-col items-center space-y-2 p-3 bg-background/50 rounded-lg shadow-md hover:shadow-primary/30 transition-shadow">
              <NextImage
                src="https://placehold.co/80x80.png"
                alt={`${chestVisualData[chestType].name} сундук`}
                width={80}
                height={80}
                className="rounded"
                data-ai-hint={chestVisualData[chestType].hint}
              />
              <span className="text-sm font-semibold">{chestVisualData[chestType].name} ({chestCounts[chestType]})</span>
              <Button size="sm" variant="secondary" className="w-full">Открыть</Button>
            </div>
          ))}
        </CardContent>
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
            <SkullIcon className="mr-2 h-5 w-5" />
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
            <DogIcon className="mr-2 h-5 w-5" />
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

      {/* Backend Test Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl font-headline text-center text-primary">Тест Бэкенда</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center">
          <Button onClick={handlePingServer} variant="outline" className="border-accent text-accent hover:bg-accent/10 hover:text-accent-foreground" disabled={isPinging}>
            {isPinging ? (
              <Loader2 className="mr-2 h-5 w-5 animate-spin" />
            ) : (
              <Wifi className="mr-2 h-5 w-5" />
            )}
            Ping Сервера
          </Button>
        </CardContent>
      </Card>

      <AlertDialog open={isPingModalOpen} onOpenChange={(isOpen) => {
        if (!isPinging) { // Only allow closing if not actively pinging
          setIsPingModalOpen(isOpen);
        }
      }}>
        <AlertDialogContent className="bg-card/90 backdrop-blur-md">
          <AlertDialogHeader>
            <AlertDialogTitle className="text-primary">Результат Ping</AlertDialogTitle>
            <AlertDialogDescription className="text-card-foreground whitespace-pre-wrap">
              {pingModalMessage}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction
              onClick={() => setIsPingModalOpen(false)}
              className="bg-primary text-primary-foreground hover:bg-primary/90"
              disabled={isPinging}
            >
              OK
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

    </div>
  );
}
