"use client";

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import Image from 'next/image';
import { cn } from '@/lib/utils';

interface NavItem {
  id: string;
  label: string;
  href: string;
  imageSrc?: string;
}

const navItems: NavItem[] = [
  { id: 'lobby', label: 'Лобби', href: '/', imageSrc: '/images/menu-lobby.png' },
  { id: 'shop', label: 'Магазин', href: '/shop', imageSrc: '/images/menu-trade.png' },
  { id: 'factory', label: 'Кузница', href: '/factory', imageSrc: '/images/menu-forge.png' },
  { id: 'inventory', label: 'Инвентарь', href: '/inventory', imageSrc: '/images/menu-inventar.png' },
  { id: 'mine', label: 'Добыча', href: '/mine', imageSrc: '/images/menu-mine.png' },
];


const BottomNavigationBar: React.FC = () => {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 h-28 bg-gradient-to-t from-background/90 via-background/70 to-transparent z-50">
      <div className="flex justify-around items-end h-full max-w-md mx-auto px-2 pb-5">
        {navItems.map((item) => {
          const isActive = item.href === '/' ? pathname === '/' : pathname.startsWith(item.href);
          
          return (
            <Link
              href={item.href}
              key={item.id}
              className={cn(
                "flex flex-col items-center justify-center w-16 h-16 rounded-2xl transition-all duration-300 ease-in-out",
                "text-white/80 hover:text-white",
                // When active, apply a more saturated gradient and a stronger shadow
                isActive && "bg-gradient-to-t from-primary/80 to-primary/40 text-white shadow-xl shadow-primary/50"
              )}
              aria-label={item.label}
            >
              {item.imageSrc ? (
                <Image src={item.imageSrc} alt={item.label} width={64} height={64} className="transition-transform duration-300" />
              ) : null}
            </Link>
          );
        })}
      </div>
    </nav>
  );
};

export default BottomNavigationBar;
