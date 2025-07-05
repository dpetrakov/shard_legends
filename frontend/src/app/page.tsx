
"use client"; // Make it a client component for useEffect

import Image from "next/image";
import Link from "next/link";
import React, { useState, useEffect } from 'react'; // Import React hooks
import { cn } from "@/lib/utils";

const SideNavLink = ({
  href,
  icon,
  label,
  hint,
}: {
  href: string;
  icon: string;
  label: string;
  hint: string;
}) => (
  <Link href={href}>
    <div className="flex flex-col items-center text-center cursor-pointer hover:scale-105 transition-transform">
      <div>
        <Image
          src={icon}
          alt={label}
          width={64}
          height={64}
          data-ai-hint={hint}
        />
      </div>
    </div>
  </Link>
);

const Fireflies = () => {
  const [fireflies, setFireflies] = useState<JSX.Element[]>([]);

  useEffect(() => {
    // Generate a random number of fireflies between 5 and 15
    const fireflyCount = Math.floor(Math.random() * 11) + 5;
    
    const generatedFireflies = Array.from({ length: fireflyCount }).map((_, i) => {
      // Randomize size, animation duration, and starting point
      const size = 2 + Math.random() * 2; // Random size between 2px and 4px
      const animationDuration = 20 + Math.random() * 20; // Slower, random duration from 20s to 40s

      const style: React.CSSProperties = {
        top: `${Math.random() * 100}%`,
        left: `${Math.random() * 100}%`,
        width: `${size}px`,
        height: `${size}px`,
        animationDuration: `${animationDuration}s`,
        animationDelay: `-${Math.random() * animationDuration}s`, // Start at a random point in the cycle
      };
      return <div key={i} className="firefly" style={style}></div>;
    });
    setFireflies(generatedFireflies);
  }, []); // Empty dependency array ensures this runs once on the client, avoiding hydration errors

  return (
    <div className="firefly-wrapper">
      {fireflies}
    </div>
  );
};


export default function LobbyPage() {
  return (
    <div className="fixed inset-0 z-0 text-white">
      <div className="relative h-full w-full bg-lobby bg-cover bg-center bg-no-repeat">
        {/* Gradient overlay for text readability */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/70 to-black/20" />

        {/* Fireflies animation */}
        <Fireflies />

        {/* Main content container */}
        <div className="relative z-10 h-full w-full">

          {/* Container for Side Navs (Vertically Centered) */}
          <div className="absolute top-1/2 left-0 right-0 w-full -translate-y-1/2 px-4 sm:px-6">
            <div className="flex justify-between items-center">
              {/* Left Nav */}
              <div className="flex flex-col items-center gap-4">
                <SideNavLink href="/achievements" icon="/images/menu-achiv.png" label="Достижения" hint="achievements trophy" />
                <SideNavLink href="/quests" icon="/images/menu-quest.png" label="Квесты" hint="quest scroll" />
              </div>

              {/* Right Nav */}
              <div className="flex flex-col items-center gap-4">
                <SideNavLink href="/friends" icon="/images/menu-friend.png" label="Друзья" hint="friends handshake" />
                <SideNavLink href="/battlepass" icon="/images/battlepass.png" label="Боевой Пропуск" hint="battle pass" />
                <SideNavLink href="/shop" icon="/images/menu-key.png" label="Ключи" hint="shop keys" />
              </div>
            </div>
          </div>

          {/* Container for Play Button (Bottom Aligned) */}
          <div className="absolute bottom-32 left-1/2 -translate-x-1/2">
            <Link href="/game" className="flex flex-col items-center text-center cursor-pointer hover:scale-105 transition-transform group">
              <div className="bg-gradient-to-t from-primary/80 to-primary/40 rounded-full shadow-xl shadow-primary/50 flex items-center justify-center p-4">
                <Image src="/images/play.png" width={180} height={180} alt="Забрать награду" data-ai-hint="play button reward" />
              </div>
              <span className="text-2xl font-headline text-white text-shadow-lg mt-2 group-hover:text-primary transition-colors">
                Забрать награду
              </span>
            </Link>
          </div>
          
        </div>
      </div>
    </div>
  );
}
