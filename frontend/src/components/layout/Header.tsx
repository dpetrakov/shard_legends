
"use client";

import { Compass, Settings } from 'lucide-react';
import Link from 'next/link';
import React from 'react';
import Image from 'next/image';
import { useAuth } from '@/contexts/AuthContext';

const ResourceIcon = ({ children, value }: { children: React.ReactNode, value: string }) => (
  <div className="flex items-center gap-1.5 bg-black/20 backdrop-blur-sm px-2 py-1 rounded-full border border-white/10 shadow-inner">
    {children}
    <span className="font-code text-xs sm:text-sm text-white font-semibold tabular-nums tracking-tighter">
      {value}
    </span>
  </div>
);

const NavIcon = ({ href, children }: { href: string, children: React.ReactNode }) => (
    <Link href={href} className="flex items-center justify-center p-2 rounded-full bg-black/20 hover:bg-black/40 transition-colors">
        {children}
    </Link>
);


const Header: React.FC = () => {
  const { user } = useAuth();

  return (
    <header 
        className="fixed top-0 left-0 right-0 z-40 bg-transparent backdrop-blur-sm"
        style={{ paddingTop: 'env(safe-area-inset-top)' }}
    >
        <div 
            style={{ paddingLeft: 'env(safe-area-inset-left)', paddingRight: 'env(safe-area-inset-right)' }} 
            className="flex items-center justify-between h-16 max-w-md mx-auto px-4"
        >
            {/* Left Side: Profile Icon & Name */}
            <div className="flex flex-col items-center">
              <Link href="/profile" className="flex items-center justify-center p-0 rounded-full bg-transparent hover:bg-black/40 transition-colors">
                <Image src="/images/menu-heroes.png" alt="Профиль" width={48} height={48} />
              </Link>
              {user && (
                <span className="text-xs font-bold text-white text-shadow -mt-1 truncate max-w-[80px]">
                  {user.username || user.firstName}
                </span>
              )}
            </div>

            {/* Center: Balances */}
            <div className="flex items-center justify-center gap-2">
                <ResourceIcon value="1.2M">
                  <Image src="/images/gold.png" alt="Gold" width={20} height={20} />
                </ResourceIcon>
                <ResourceIcon value="345K">
                  <Image src="/images/diamond.png" alt="Diamonds" width={20} height={20} />
                </ResourceIcon>
                <ResourceIcon value="8/10"><Image src="/images/sl.png" alt="SL" width={20} height={20} /></ResourceIcon>
            </div>
            
            {/* Right Side: Nav Icons */}
            <div className="flex items-center gap-2">
              <NavIcon href="#">
                  <Compass className="w-6 h-6 text-secondary" />
              </NavIcon>
              <NavIcon href="/settings">
                  <Settings className="w-6 h-6 text-secondary" />
              </NavIcon>
            </div>
        </div>
    </header>
  );
};

export default Header;
