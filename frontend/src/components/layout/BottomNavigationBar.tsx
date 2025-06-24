
"use client";

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { SearchCode, ShoppingBasket, UserCog, Package, Anvil } from 'lucide-react';
import { cn } from '@/lib/utils';

const navItems = [
  { id: 'search', label: 'Поиск', icon: SearchCode, href: '/' },
  { id: 'shop', label: 'Магазин', icon: ShoppingBasket, href: '/shop' },
  { id: 'factory', label: 'Кузница', icon: Anvil, href: '/factory' },
  { id: 'inventory', label: 'Инвентарь', icon: Package, href: '/friends' },
  { id: 'profile', label: 'Профиль', icon: UserCog, href: '/profile' },
];


const BottomNavigationBar: React.FC = () => {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 h-16 bg-card border-t border-border shadow-top z-50">
      <div className="flex justify-around items-center h-full max-w-md mx-auto px-2">
        {navItems.map((item) => {
          const IconComponent = item.icon;
          const isActive = pathname === item.href;
          return (
            <Link
              href={item.href}
              key={item.id}
              className={cn(
                "flex flex-col items-center justify-center w-1/5 h-full text-xs transition-colors duration-200 ease-in-out",
                isActive ? "text-accent" : "text-muted-foreground hover:text-foreground"
              )}
              aria-label={item.label}
            >
              <IconComponent
                size={24}
                className={cn("mb-0.5", isActive ? "stroke-[2.5px]" : "stroke-[2px]")}
              />
              <span className={cn("truncate", isActive ? "font-semibold" : "font-normal")}>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
};

export default BottomNavigationBar;
