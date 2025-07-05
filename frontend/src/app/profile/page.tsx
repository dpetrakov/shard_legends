
"use client";

import React from 'react';
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { User, Plus, Axe, Pickaxe, Shovel, Wheat, Users, Dices, Anvil, Banknote, FlaskConical } from "lucide-react";
import Image from 'next/image';
import { Separator } from "@/components/ui/separator";

const ParameterItem = ({ icon, title, description, level }: { icon: React.ReactNode, title: string, description: string, level: number }) => (
    <div className="flex items-center gap-4 py-3">
        <div className="bg-primary/10 p-2 rounded-lg border border-primary/20">
            {icon}
        </div>
        <div className="flex-grow">
            <p className="font-semibold">{title}</p>
            <p className="text-xs text-muted-foreground">{description}</p>
        </div>
        <div className="flex items-center gap-2">
            <span className="font-bold text-lg w-12 text-center">Ур. {level}</span>
            <Button size="icon" className="w-8 h-8" disabled>
                <Plus className="h-4 w-4" />
            </Button>
        </div>
    </div>
);


export default function ProfilePage() {

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

  const parameters = [
    { id: 'builder', icon: <Shovel className="w-6 h-6 text-primary" />, title: 'Строитель', description: 'Увеличивает эффективность и прочность лопат.', level: 0 },
    { id: 'reaper', icon: <Wheat className="w-6 h-6 text-primary" />, title: 'Жнец', description: 'Повышает скорость и удачу при использовании серпа.', level: 0 },
    { id: 'leader', icon: <Users className="w-6 h-6 text-primary" />, title: 'Лидер', description: 'Позволяет приглашать больше друзей и увеличивает бонусы.', level: 0 },
    { id: 'lumberjack', icon: <Axe className="w-6 h-6 text-primary" />, title: 'Лесоруб', description: 'Улучшает навыки владения топором для добычи дерева.', level: 0 },
    { id: 'miner', icon: <Pickaxe className="w-6 h-6 text-primary" />, title: 'Шахтёр', description: 'Увеличивает силу и скорость добычи киркой.', level: 0 },
    { id: 'player', icon: <Dices className="w-6 h-6 text-primary" />, title: 'Игрок', description: 'Повышает шанс найти редкие предметы и сокровища.', level: 0 },
    { id: 'blacksmith', icon: <Anvil className="w-6 h-6 text-primary" />, title: 'Кузнец', description: 'Ускоряет процессы переплавки, ковки и починки.', level: 0 },
    { id: 'trader', icon: <Banknote className="w-6 h-6 text-primary" />, title: 'Торговец', description: 'Снижает комиссии на рынке и улучшает условия обмена.', level: 0 },
    { id: 'researcher', icon: <FlaskConical className="w-6 h-6 text-primary" />, title: 'Исследователь', description: 'Открывает новые рецепты и технологии быстрее.', level: 0 },
  ];

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
      
      {/* Parameters Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardHeader>
            <CardTitle className="text-2xl font-headline text-center text-primary">Параметры</CardTitle>
        </CardHeader>
        <CardContent>
            <div className="flex justify-between items-center bg-card/50 p-3 rounded-lg mb-4">
                <div>
                    <p className="text-sm text-muted-foreground">Свободные очки</p>
                    <p className="text-2xl font-bold">0</p>
                </div>
                <Button>
                    <Image src="/images/diamond.png" alt="Бриллианты" width={20} height={20} className="mr-2" />
                    Купить
                </Button>
            </div>
            
            <Separator className="my-4 bg-border/50" />

            <div>
                {parameters.map((param, index) => (
                    <React.Fragment key={param.id}>
                        <ParameterItem 
                            icon={param.icon}
                            title={param.title}
                            description={param.description}
                            level={param.level}
                        />
                        {index < parameters.length - 1 && <Separator className="bg-border/30" />}
                    </React.Fragment>
                ))}
            </div>
        </CardContent>
      </Card>
    </div>
  );
}
