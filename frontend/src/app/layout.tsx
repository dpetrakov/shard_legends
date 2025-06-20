
import type { Metadata } from 'next';
import './globals.css';
import BottomNavigationBar from '@/components/layout/BottomNavigationBar';
import { ChestProvider } from '@/contexts/ChestContext';
import { IconSetProvider } from '@/contexts/IconSetContext'; // Import IconSetProvider

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
    <html lang="en" className="theme-fantasy-casual">
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300..700&display=swap" rel="stylesheet" />
      </head>
      <body className="font-body antialiased min-h-screen text-foreground flex flex-col">
        <IconSetProvider> {/* Wrap with IconSetProvider */}
          <ChestProvider>
            <main className="flex-grow overflow-y-auto">
             {children}
            </main>
            <BottomNavigationBar />
          </ChestProvider>
        </IconSetProvider>
      </body>
    </html>
  );
}
