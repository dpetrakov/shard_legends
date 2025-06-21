
"use client";

import { useState } from "react";
import "@/types/telegram";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
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
    setPingModalMessage("Testing auth service...");
    setIsPingModalOpen(true);
    const apiUrl = process.env.NEXT_PUBLIC_API_URL;

    if (!apiUrl) {
      setPingModalMessage("Ошибка: Переменная окружения NEXT_PUBLIC_API_URL не установлена.");
      setIsPinging(false);
      return;
    }

    try {
      // Get Telegram Web App data
      let initData = '';
      
      // Проверяем, доступен ли Telegram Web App
      if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
        // Инициализируем Telegram Web App если не инициализирован
        if (!window.Telegram.WebApp.isExpanded) {
          window.Telegram.WebApp.ready();
          window.Telegram.WebApp.expand();  
        }
        
        // Получаем initData
        const telegramInitData = window.Telegram.WebApp.initData;
        if (telegramInitData && telegramInitData.trim()) {
          initData = telegramInitData;
          setPingModalMessage("Отправляем запрос с Telegram данными...");
        } else {
          setPingModalMessage("Отправляем запрос без Telegram данных...");
        }
      } else {
        setPingModalMessage("Отправляем запрос без Telegram данных...");
      }

      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };

      // Add Telegram data header if available
      if (initData) {
        headers['X-Telegram-Init-Data'] = initData;
      }

      const response = await fetch(`${apiUrl}/api/auth`, {
        method: 'POST',
        headers: headers,
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP error! status: ${response.status}, message: ${errorText || 'Unknown server error'}`);
      }
      
      const data = await response.json();
      
      if (data.success) {
        setPingModalMessage(`✅ Авторизация успешна!\n\nПользователь: ${data.user.first_name} ${data.user.last_name || ''}\nTelegram ID: ${data.user.telegram_id}\nUsername: ${data.user.username || 'не указан'}\nНовый пользователь: ${data.user.is_new_user ? 'Да' : 'Нет'}\n\nТокен получен: ${data.token.substring(0, 20)}...`);
      } else {
        setPingModalMessage(`❌ Ошибка авторизации:\n${data.message || data.error}`);
      }
    } catch (error) {
      if (error instanceof Error) {
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

      {/* Auth Test Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl font-headline text-center text-primary">Тест Авторизации</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center">
          <Button onClick={handlePingServer} variant="outline" className="border-accent text-accent hover:bg-accent/10 hover:text-accent-foreground" disabled={isPinging}>
            {isPinging ? (
              <Loader2 className="mr-2 h-5 w-5 animate-spin" />
            ) : (
              <User className="mr-2 h-5 w-5" />
            )}
            Тест Auth Service
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

      <AlertDialog open={isPingModalOpen} onOpenChange={(isOpen) => {
        if (!isPinging) { // Only allow closing if not actively pinging
          setIsPingModalOpen(isOpen);
        }
      }}>
        <AlertDialogContent className="bg-card/90 backdrop-blur-md">
          <AlertDialogHeader>
            <AlertDialogTitle className="text-primary">Результат Авторизации</AlertDialogTitle>
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
