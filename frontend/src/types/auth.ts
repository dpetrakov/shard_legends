
export interface User {
  id: string; // Changed from number to string for UUID
  telegramId?: number; // Was missing
  isBot?: boolean;
  firstName: string;
  lastName?: string;
  username?: string;
  languageCode?: string;
  isPremium?: boolean;
  photoUrl?: string; // Added
  isNewUser?: boolean; // Added
  parameters?: Record<string, number>;
}

export interface AuthContextType {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
  updateUser: (partialUser: Partial<User>) => void;
}
