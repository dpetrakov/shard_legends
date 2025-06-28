
"use client";

import type { Theme, ThemeContextType } from '@/types/theme';
import React, { createContext, useContext, useEffect, useState, type ReactNode, useCallback } from 'react';

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setThemeState] = useState<Theme>('fantasy-casual'); // Default theme

  useEffect(() => {
    const storedTheme = localStorage.getItem('appTheme') as Theme | null;
    const initialTheme = storedTheme || 'fantasy-casual'; // Use stored or default
    setThemeState(initialTheme);
    if (typeof window !== 'undefined') {
      document.documentElement.className = `theme-${initialTheme}`;
    }
  }, []);

  const setTheme = useCallback((newTheme: Theme) => {
    localStorage.setItem('appTheme', newTheme);
    setThemeState(newTheme);
    if (typeof window !== 'undefined') {
      document.documentElement.className = `theme-${newTheme}`;
    }
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};
