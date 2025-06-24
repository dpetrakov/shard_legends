
"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { User, Gem, Candy, Skull as SkullIcon, Dog as DogIcon, Image as ImageIcon } from "lucide-react";
import { useIconSet } from "@/contexts/IconSetContext";
import { cn } from "@/lib/utils";

export default function ProfilePage() {
  const { iconSet, setIconSet } = useIconSet();

  const handleTelegramAuth = async () => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL;
    if (!apiUrl || apiUrl === "YOUR_BACKEND_URL_HERE") {
      alert("Критическая ошибка: URL бэкенда не настроен в .env файле.");
      console.error("Error: NEXT_PUBLIC_API_URL is not defined in .env file.");
      return;
    }

    const tg = (window as any).Telegram;
    if (!tg || !tg.WebApp) {
        alert("Ошибка: Не удалось найти API Telegram. Убедитесь, что приложение запущено внутри Telegram.");
        return;
    }
    
    const initData = tg.WebApp.initData;
    const initDataUnsafe = tg.WebApp.initDataUnsafe || {};
    const userInfo = initDataUnsafe.user || { id: 'неизвестно', username: 'неизвестно' };

    if (!initData) {
      alert(`Ошибка: Данные Telegram (initData) не найдены для пользователя ${userInfo.username} (ID: ${userInfo.id}).\n\nУбедитесь, что приложение открыто через кнопку в боте, а не по прямой ссылке.`);
      console.error("Error: Telegram initData is not available.");
      return;
    }
    
    try {
      const response = await fetch(`${apiUrl}/api/auth`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Telegram-Init-Data': initData
        }
      });

      const responseBodyText = await response.text();
      
      let data;
      try {
        data = JSON.parse(responseBodyText);
      } catch (jsonError) {
        alert(`Ошибка ответа сервера: Не удалось разобрать JSON.\nСтатус: ${response.status} ${response.statusText}\nОтвет: ${responseBodyText.substring(0, 200)}...`);
        console.error("Failed to parse JSON response:", { status: response.status, body: responseBodyText });
        return;
      }

      if (response.ok && data.success) {
        alert(`Авторизация успешна! Привет, ${data.user.firstName} (ID: ${data.user.id})!`);
        console.log(`Пользователь: ${data.user.firstName} ${data.user.lastName}`);
        console.log(`JWT токен: ${data.token}`);
        console.log(`Новый пользователь: ${data.isNewUser}`);
      } else {
        const errorMessage = data.message || "Неизвестная ошибка сервера";
        let alertMessage = `Ошибка авторизации для пользователя ${userInfo.username} (ID: ${userInfo.id}).\n\nСообщение от сервера: "${errorMessage}"\nСтатус: ${response.status} ${response.statusText}`;

        if (response.status === 401) {
            alertMessage += `\n\n(Ошибка 401 Unauthorized обычно означает, что токен бота на сервере не совпадает с токеном бота, в котором запущено приложение. Проверьте это с бэкенд-разработчиком.)`;
        }

        alert(alertMessage);
        console.error("Authentication failed. Server responded with:", {
          status: response.status,
          statusText: response.statusText,
          responseBody: data,
        });
      }
    } catch (error: any) {
      alert(`Сетевая ошибка: Не удалось отправить запрос на сервер для пользователя ${userInfo.username} (ID: ${userInfo.id}).\n\nПроверьте URL и ваше интернет-соединение.\nДетали: ${error.message}`);
      console.error("Fetch error:", error);
    }
  };

  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
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
          <div className="flex flex-col sm:flex-row gap-2">
            <Button variant="outline" className="border-primary text-primary hover:bg-primary/10 hover:text-primary-foreground">
              Подключить кошелек
            </Button>
            <Button onClick={handleTelegramAuth} variant="default">
              Авторизация
            </Button>
          </div>
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
    </div>
  );
}
