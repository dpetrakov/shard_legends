
"use client";

import Link from "next/link";
import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import { useAuth } from "@/contexts/AuthContext";
import { Award, Calendar, Gift } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

const FireflyAnimation = () => {
    const [count, setCount] = useState(0);

    useEffect(() => {
        // Random number between 7 and 15.
        // This runs only on the client, after hydration, to avoid mismatch.
        setCount(Math.floor(Math.random() * 9) + 7);
    }, []); // Empty dependency array ensures this runs once on mount.

    // Don't render anything on the server or before the count is set.
    if (count === 0) {
        return null;
    }
    
    return (
        <div className="firefly-wrapper fixed inset-0 z-0" aria-hidden="true">
            {Array.from({ length: count }).map((_, i) => (
                <div
                    key={i}
                    className="firefly"
                    style={{
                        width: `${Math.random() * 3 + 2}px`,
                        height: `${Math.random() * 3 + 2}px`,
                        top: `${Math.random() * 100}%`,
                        left: `${Math.random() * 100}%`,
                        // Slower animation: Tripled the duration from 5-15s to 15-45s
                        animationDuration: `${(Math.random() * 10 + 5) * 3}s`,
                        // Increased delay range for more staggered appearance
                        animationDelay: `${Math.random() * 15}s`,
                    }}
                />
            ))}
        </div>
    );
};

const LobbyMenuItem = ({ href, label, imageSrc, icon, side }: { href: string; label:string; imageSrc?: string; icon?: React.ReactNode; side: 'left' | 'right' }) => (
    <Tooltip>
        <TooltipTrigger asChild>
            <Link href={href} className={cn(
                "group flex items-center justify-center w-16 h-16 rounded-2xl transition-all duration-300",
                "bg-black/20 backdrop-blur-sm border border-white/10 shadow-lg hover:bg-primary/30 hover:border-primary/50 hover:scale-110"
            )}>
                {imageSrc ? <Image src={imageSrc} alt={label} width={40} height={40} /> : icon}
            </Link>
        </TooltipTrigger>
        <TooltipContent side={side === 'left' ? 'right' : 'left'}>
            <p>{label}</p>
        </TooltipContent>
    </Tooltip>
);


export default function LobbyPage() {
  const { user } = useAuth(); // Auth is working, keep this hook for context.

  const leftMenuItems = [
    { href: '/quests', label: 'Задания', icon: <Calendar className="w-8 h-8 text-primary" /> },
    { href: '/battlepass', label: 'Пропуск', imageSrc: '/images/battlepass.png' },
    { href: '/friends', label: 'Друзья', imageSrc: '/images/menu-friend.png' },
  ];

  const rightMenuItems = [
    { href: '/achievements', label: 'Достижения', icon: <Award className="w-8 h-8 text-primary" /> },
    { href: '/gifts', label: 'Подарки', icon: <Gift className="w-8 h-8 text-primary" /> },
    { href: '/profile', label: 'Профиль', imageSrc: '/images/menu-heroes.png' },
  ];

  return (
    <TooltipProvider>
        <div className="fixed inset-0 bg-cover bg-center bg-no-repeat bg-lobby">
            <FireflyAnimation />
            <div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />

            <div className="relative z-10 h-full w-full flex justify-between items-center p-4">
                {/* Left Menu */}
                <div className="flex flex-col gap-4">
                    {leftMenuItems.map(item => (
                        <LobbyMenuItem key={item.href} {...item} side="left" />
                    ))}
                </div>

                {/* Center Play Button */}
                <div className="flex flex-col items-center">
                    <Link href="/game" className="flex flex-col items-center text-center cursor-pointer hover:scale-105 transition-transform group">
                        <div className="bg-gradient-to-t from-primary/80 to-primary/40 rounded-full shadow-xl shadow-primary/50 flex items-center justify-center p-4">
                            <Image 
                                src="/images/play.png" 
                                alt="Играть" 
                                width={180} 
                                height={180} 
                                data-ai-hint="play game button"
                                priority
                            />
                        </div>
                        <span className="text-3xl font-headline text-white text-shadow-lg mt-4 group-hover:text-primary transition-colors">
                            ИГРАТЬ
                        </span>
                    </Link>
                </div>

                {/* Right Menu */}
                <div className="flex flex-col gap-4">
                    {rightMenuItems.map(item => (
                        <LobbyMenuItem key={item.href} {...item} side="right" />
                    ))}
                </div>
            </div>
        </div>
    </TooltipProvider>
  );
}
