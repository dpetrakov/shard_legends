
import { Coins, Gem, Wallet } from 'lucide-react';
import React from 'react';

const BalanceDisplay: React.FC<{ icon: React.ElementType; value: string | number; label: string }> = ({ icon: Icon, value, label }) => (
  <div className="flex items-center gap-1.5 sm:gap-2 bg-black/10 backdrop-blur-sm px-2 sm:px-3 py-1 rounded-full border border-border/50 shadow-inner">
    <Icon className="h-5 w-5 text-primary shrink-0" />
    <span className="font-code text-xs sm:text-sm text-foreground font-semibold text-left tabular-nums tracking-tighter">
      {typeof value === 'number' ? value.toLocaleString() : value}
    </span>
  </div>
);

const Header: React.FC = () => {
  return (
    <header className="sticky top-0 z-40 w-full bg-card/80 backdrop-blur-md border-b border-border">
      <div className="flex items-center justify-around h-16 max-w-md mx-auto px-2 gap-2">
        <BalanceDisplay icon={Coins} value={0} label="Gold" />
        <BalanceDisplay icon={Gem} value={0} label="Diamonds" />
        <BalanceDisplay icon={Wallet} value={0} label="$SLCW" />
      </div>
    </header>
  );
};

export default Header;
