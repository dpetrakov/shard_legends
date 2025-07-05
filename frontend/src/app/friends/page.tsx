
"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Share2, Copy, Star } from "lucide-react";
import Image from 'next/image';

const CurrencyDisplay = ({ icon, text, className }: { icon: React.ReactNode, text: string, className?: string }) => (
    <div className={`flex items-center gap-1.5 font-semibold ${className}`}>
        {icon}
        <span>{text}</span>
    </div>
);

const ReferralLevel = ({ level, starPercent, diamondPercent, userCount }: { level: string, starPercent: number, diamondPercent: number, userCount: number }) => (
    <div className="flex justify-between items-center py-2">
        <div>
            <p className="font-bold text-sm text-foreground">{level}</p>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <CurrencyDisplay icon={<Star className="w-3 h-3 text-yellow-400" />} text={`${starPercent}%`} />
                <span>&</span>
                <CurrencyDisplay icon={<Image src="/images/diamond.png" alt="Бриллианты" width={12} height={12} className="inline-block" />} text={`${diamondPercent}%`} />
            </div>
        </div>
        <div className="text-sm text-muted-foreground">{userCount} пользователей</div>
    </div>
);


export default function FriendsPage() {
  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
      
      {/* Main Invite Card */}
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl overflow-hidden">
        <div className="relative p-6 bg-primary/20">
            <Image 
                src="https://placehold.co/400x150.png" 
                alt="Пригласительный баннер"
                width={400}
                height={150}
                className="absolute top-0 left-0 w-full h-full object-cover opacity-20"
                data-ai-hint="fantasy banner"
            />
             <div className="relative z-10 flex flex-col items-center text-center space-y-3">
                <h2 className="text-xl font-bold text-primary-foreground text-shadow">Приглашайте друзей и получайте</h2>
                <div className="flex flex-wrap justify-center items-center gap-x-4 gap-y-2 text-primary-foreground text-shadow">
                    <CurrencyDisplay icon={<Image src="/images/gold.png" alt="Золото" width={20} height={20} />} text="Золото" />
                    <CurrencyDisplay icon={<Star className="w-5 h-5 text-yellow-300" />} text="Звезды" />
                    <CurrencyDisplay icon={<Image src="/images/diamond.png" alt="USDT" width={20} height={20} />} text="USDT" />
                </div>
                 <div className="flex gap-3 pt-2">
                    <Button><Share2 className="mr-2"/> Поделиться</Button>
                    <Button variant="secondary"><Copy className="mr-2"/> Копировать</Button>
                </div>
            </div>
        </div>
      </Card>
      
      <Tabs defaultValue="details" className="w-full max-w-md">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="details">Детали</TabsTrigger>
          <TabsTrigger value="commissions">Комиссии</TabsTrigger>
          <TabsTrigger value="friends">Друзья</TabsTrigger>
        </TabsList>
        
        <TabsContent value="details" className="mt-6 space-y-6">
            {/* 'Mining alone is boring!' Card */}
            <Card className="w-full bg-accent/10 backdrop-blur-md shadow-xl border-dashed border-primary/50">
                <CardContent className="pt-6 flex items-center justify-between gap-4">
                    <div className="space-y-1">
                        <CardTitle className="text-lg font-headline text-primary">Играть в одиночку скучно!</CardTitle>
                        <CardDescription className="text-xs">Приглашайте друзей и получайте бонусы, которые помогут вам получить еще больше.</CardDescription>
                    </div>
                    <Image 
                        src="/images/menu-friend.png"
                        alt="Иконка друзей"
                        width={64}
                        height={64}
                        className="flex-shrink-0"
                    />
                </CardContent>
            </Card>

            {/* Rewards Section */}
            <Card className="w-full bg-card/80 backdrop-blur-md shadow-xl">
                <CardHeader>
                    <CardTitle className="text-xl font-headline text-center text-primary">Награды за рефералов</CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-1 md:grid-cols-3 gap-4 bg-background/50 p-4 rounded-lg">
                    <div className="md:col-span-1 p-3 bg-card/50 rounded-lg flex flex-col justify-center text-center md:text-left">
                        <p className="text-sm font-semibold text-muted-foreground">Тип операции</p>
                        <p className="text-xs mt-1 flex items-center gap-1 flex-wrap justify-center md:justify-start">
                            <span>Реферал потратил</span>
                            <Star className="inline w-3 h-3 text-yellow-400"/> 
                            <span>Telegram Stars на покупку</span>
                            <Image src="/images/diamond.png" alt="Бриллианты" width={14} height={14} className="inline-block align-middle" />
                            <span>Бриллиантов.</span>
                        </p>
                    </div>
                    <div className="md:col-span-2">
                        <ReferralLevel level="1-й уровень" starPercent={10} diamondPercent={5} userCount={0} />
                        <Separator className="my-1 bg-border/50" />
                        <ReferralLevel level="2-й уровень" starPercent={3} diamondPercent={2} userCount={0} />
                        <Separator className="my-1 bg-border/50" />
                        <ReferralLevel level="3-й уровень" starPercent={1} diamondPercent={1} userCount={0} />
                    </div>
                </CardContent>
            </Card>

            {/* Invite Bonuses Section */}
            <Card className="w-full bg-card/80 backdrop-blur-md shadow-xl">
                <CardHeader>
                    <CardTitle className="text-xl font-headline text-center text-primary">Бонусы за приглашение</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="text-left">Телеграмм версия</TableHead>
                                <TableHead className="text-center">Получает друг</TableHead>
                                <TableHead className="text-center">Получаете Вы</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            <TableRow>
                                <TableCell className="font-medium text-left">Обычная</TableCell>
                                <TableCell className="text-center">
                                    <span className="flex items-center justify-center gap-1.5">10 <Image src="/images/gold.png" alt="Gold" width={16} height={16} /></span>
                                </TableCell>
                                <TableCell className="text-center">
                                    <span className="flex items-center justify-center gap-1.5">10 <Image src="/images/gold.png" alt="Gold" width={16} height={16} /></span>
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell className="font-medium text-left">Премиум</TableCell>
                                <TableCell className="text-center">
                                     <span className="flex items-center justify-center gap-1.5">50 <Image src="/images/gold.png" alt="Gold" width={16} height={16} /></span>
                                </TableCell>
                                <TableCell className="text-center">
                                     <span className="flex items-center justify-center gap-1.5">50 <Image src="/images/gold.png" alt="Gold" width={16} height={16} /></span>
                                </TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </TabsContent>

        <TabsContent value="commissions">
            <div className="mt-6 space-y-6">
                <Card className="w-full bg-card/80 backdrop-blur-md shadow-xl">
                    <CardHeader>
                        <CardTitle className="text-xl font-headline text-primary">Заблокированный баланс</CardTitle>
                        <CardDescription>Будет доступен через 21 день</CardDescription>
                    </CardHeader>
                    <CardContent className="flex justify-around items-center pt-2">
                        <div className="flex flex-col items-center gap-1">
                             <CurrencyDisplay icon={<Star className="w-8 h-8 text-yellow-400" />} text="0" className="text-2xl" />
                             <span className="text-xs text-muted-foreground">Звезды</span>
                        </div>
                         <div className="flex flex-col items-center gap-1">
                             <CurrencyDisplay icon={<Image src="/images/diamond.png" alt="Бриллианты" width={32} height={32} />} text="0" className="text-2xl"/>
                             <span className="text-xs text-muted-foreground">Бриллианты</span>
                        </div>
                    </CardContent>
                </Card>

                <Card className="w-full bg-card/80 backdrop-blur-md shadow-xl">
                    <CardHeader>
                        <CardTitle className="text-xl font-headline text-primary text-center">Свободный баланс</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col items-center gap-2">
                        <CurrencyDisplay icon={<Star className="w-10 h-10 text-yellow-400" />} text="0" className="text-4xl" />
                        <span className="text-xs text-muted-foreground">Звезды</span>
                    </CardContent>
                </Card>

                <div className="flex flex-col gap-3 pt-2">
                    <Button variant="secondary" size="lg">Обменять звезды на USDT</Button>
                    <Button variant="default" size="lg">Обменять звезды на Бриллианты</Button>
                </div>
            </div>
        </TabsContent>
        
        <TabsContent value="friends">
             <Card className="mt-6">
                <CardContent className="pt-6">
                    <p className="text-center text-muted-foreground">Здесь будет отображаться список ваших друзей.</p>
                </CardContent>
            </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
