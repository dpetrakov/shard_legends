
"use client";

import React, { useEffect, useCallback } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { User, Plus, Axe, Pickaxe, Shovel, Wheat, Users, Dices, Anvil, Banknote, FlaskConical, LogOut } from "lucide-react";
import Image from 'next/image';
import { Separator } from "@/components/ui/separator";
import { useAuth } from '@/contexts/AuthContext';
import type { User as UserType } from '@/types/auth';

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
  const { user, isAuthenticated, token, updateUser } = useAuth();
  const apiUrl = 'https://dev-forly.slcw.dimlight.online';

  const fetchProfile = useCallback(async () => {
    if (!token) {
      return;
    }
    
    try {
      const requestUrl = `${apiUrl}/api/user/profile`;
      const response = await fetch(requestUrl, {
        method: 'GET',
        mode: 'cors',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'application/json'
        }
      });
      
      const responseBodyText = await response.text();
      if (!response.ok) {
          console.error("Failed to fetch profile, error response:", { status: response.status, body: responseBodyText });
          return;
      }
      
      if (!responseBodyText) {
          console.log("Profile fetch response was empty.");
          return;
      }
      
      let data;
      try {
        data = JSON.parse(responseBodyText);
      } catch (jsonError) {
        console.error("Failed to parse profile JSON response:", { status: response.status, body: responseBodyText });
        return;
      }

      if (data.user) {
        const serverUser = data.user;
        const partialClientUser: Partial<UserType> = {
          id: serverUser.id,
          telegramId: serverUser.telegram_id,
          firstName: serverUser.first_name,
          lastName: serverUser.last_name,
          username: serverUser.username,
          languageCode: serverUser.language_code,
          isPremium: serverUser.is_premium,
          photoUrl: serverUser.photo_url,
          isNewUser: serverUser.is_new_user,
          parameters: serverUser.parameters || {},
        };
        updateUser(partialClientUser);
        console.log("User profile updated successfully:", partialClientUser);
      } else {
        console.error("Failed to fetch user profile:", data.message || "Unknown error");
      }
    } catch (error) {
      console.error("Network error fetching user profile:", error);
    }
  }, [token, updateUser, apiUrl]);

  useEffect(() => {
    if (isAuthenticated) {
      fetchProfile();
    }
  }, [isAuthenticated, fetchProfile]);


  const parameters = [
    { id: 'builder', icon: <Shovel className="w-6 h-6 text-primary" />, title: 'Строитель', description: 'Увеличивает эффективность и прочность лопат.' },
    { id: 'reaper', icon: <Wheat className="w-6 h-6 text-primary" />, title: 'Жнец', description: 'Повышает скорость и удачу при использовании серпа.' },
    { id: 'leader', icon: <Users className="w-6 h-6 text-primary" />, title: 'Лидер', description: 'Позволяет приглашать больше друзей и увеличивает бонусы.' },
    { id: 'lumberjack', icon: <Axe className="w-6 h-6 text-primary" />, title: 'Лесоруб', description: 'Улучшает навыки владения топором для добычи дерева.' },
    { id: 'miner', icon: <Pickaxe className="w-6 h-6 text-primary" />, title: 'Шахтёр', description: 'Увеличивает силу и скорость добычи киркой.' },
    { id: 'player', icon: <Dices className="w-6 h-6 text-primary" />, title: 'Игрок', description: 'Повышает шанс найти редкие предметы и сокровища.' },
    { id: 'blacksmith', icon: <Anvil className="w-6 h-6 text-primary" />, title: 'Кузнец', description: 'Ускоряет процессы переплавки, ковки и починки.' },
    { id: 'trader', icon: <Banknote className="w-6 h-6 text-primary" />, title: 'Торговец', description: 'Снижает комиссии на рынке и улучшает условия обмена.' },
    { id: 'researcher', icon: <FlaskConical className="w-6 h-6 text-primary" />, title: 'Исследователь', description: 'Открывает новые рецепты и технологии быстрее.' },
  ];

  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
      {/* User Info Section */}
      <Card className="w-full max-w-md backdrop-blur-md shadow-xl">
        <CardContent className="pt-6 flex flex-col items-center space-y-4">
          <Avatar className="w-24 h-24 border-2 border-primary">
            <AvatarImage src={user?.photoUrl || undefined} alt={user?.username || "User Avatar"} />
            <AvatarFallback>
              <User className="w-12 h-12" />
            </AvatarFallback>
          </Avatar>
          <span className="text-2xl font-headline">
            {isAuthenticated && user ? user.username || user.firstName : 'Загрузка...'}
          </span>
          <div className="flex flex-col sm:flex-row gap-2">
            <Button variant="outline" className="border-primary text-primary hover:bg-primary/10 hover:text-primary-foreground">
              Подключить кошелек
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
                            level={user?.parameters?.[param.id] || 0}
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
