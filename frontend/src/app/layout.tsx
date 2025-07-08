
import type { Metadata } from 'next';
import './globals.css';
import BottomNavigationBar from '@/components/layout/BottomNavigationBar';
import { ChestProvider } from '@/contexts/ChestContext';
import { IconSetProvider } from '@/contexts/IconSetContext';
import { InventoryProvider } from '@/contexts/InventoryContext';
import { ProductionProvider } from '@/contexts/ProductionContext';
import { MiningProvider } from '@/contexts/MiningContext';
import Script from 'next/script';
import Header from '@/components/layout/Header';
import { cn } from '@/lib/utils';
import { AuthProvider } from '@/contexts/AuthContext';

export const metadata: Metadata = {
  title: 'Crystal Cascade',
  description: 'A match-3 game by Firebase Studio',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300..700&display=swap" rel="stylesheet" />
        <Script src="https://telegram.org/js/telegram-web-app.js" strategy="beforeInteractive" />
      </head>
      <body className={cn("theme-fantasy-casual font-body antialiased min-h-screen text-foreground flex flex-col relative")}>
        <AuthProvider>
          <IconSetProvider>
            <InventoryProvider>
              <ChestProvider>
                <ProductionProvider>
                  <MiningProvider>
                    <Header />
                    <main className="flex-grow overflow-y-auto pt-16 pb-28">
                    {children}
                    </main>
                    <BottomNavigationBar />
                  </MiningProvider>
                </ProductionProvider>
              </ChestProvider>
            </InventoryProvider>
          </IconSetProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
