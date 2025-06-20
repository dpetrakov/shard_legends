
export type Theme = 'neon' | 'fantasy-casual';

export interface ThemeContextType {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}
