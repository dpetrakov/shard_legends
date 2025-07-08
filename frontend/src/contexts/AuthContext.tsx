
"use client";

import React, { createContext, useContext, useState, useEffect, type ReactNode, useCallback } from 'react';
import type { AuthContextType, User } from '@/types/auth';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const apiUrl = 'https://dev-forly.slcw.dimlight.online/api';

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  // Simplified login: only updates component state, no localStorage
  const login = useCallback((newToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
  }, []);

  // Simplified logout
  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
  }, []);
  
  // Simplified updateUser
  const updateUser = useCallback((partialUser: Partial<User>) => {
    setUser(prevUser => {
      if (!prevUser) return null;
      const updatedUser = { ...prevUser, ...partialUser };
      return updatedUser;
    });
  }, []);

  const autoLogin = useCallback(async (initData: string) => {
    if (!initData) {
      console.error("Auth Error: autoLogin called with no initData.");
      setIsLoaded(true); // Ensure we unblock UI
      return;
    }

    try {
      const requestUrl = `${apiUrl}/auth`;
      const response = await fetch(requestUrl, {
        method: 'POST',
        mode: 'cors',
        headers: {
          'Accept': 'application/json',
          'X-Telegram-Init-Data': initData,
        }
      });

      const responseBodyText = await response.text();

      if (!response.ok) {
        console.error("Auto-login failed. Server responded with an error.", {
          status: response.status,
          responseBody: responseBodyText,
        });
        return;
      }
      
      const data = JSON.parse(responseBodyText);

      if (data.success && data.token && data.user) {
        const serverUser = data.user;
        const clientUser: User = {
          id: serverUser.id,
          telegramId: serverUser.telegram_id,
          firstName: serverUser.first_name,
          lastName: serverUser.last_name,
          username: serverUser.username,
          languageCode: serverUser.language_code,
          isPremium: serverUser.is_premium,
          photoUrl: serverUser.photo_url,
          isNewUser: serverUser.is_new_user,
          parameters: serverUser.parameters || {},
        };
        login(data.token, clientUser);
        console.log(`Auth: Fresh login successful for user: ${clientUser.firstName || clientUser.username}.`);
      } else {
        console.error("Auto-login response was not successful.", { responseBody: data });
      }
    } catch (error) {
      console.error("A network or other error occurred during auto-login.", error);
    } finally {
        // Unblock the UI after the auth attempt
        setIsLoaded(true);
    }
  }, [login, apiUrl]);

  // Main effect to trigger authentication on every app load
  useEffect(() => {
    const initAuth = async () => {
      if (typeof window !== 'undefined' && (window as any).Telegram?.WebApp) {
        const tg = (window as any).Telegram.WebApp;
        
        // Use tg.ready() to ensure all Telegram-related objects are available.
        tg.ready();
        
        if (tg.initData) {
            console.log("Auth: Found initData. Performing fresh authentication.");
            await autoLogin(tg.initData);
        } else {
            console.error("Auth Error: No Telegram initData found. Cannot authenticate.");
            setIsLoaded(true); // Still unblock UI
        }
      } else {
        console.error("Auth Error: Not running in a Telegram WebApp environment.");
        // For local development outside Telegram, you might want a mock login here.
        // For now, just unblock the UI.
        setIsLoaded(true);
      }
    };

    initAuth();
  }, [autoLogin]);

  const isAuthenticated = !!token && isLoaded;

  return (
    <AuthContext.Provider value={{ token, user, isAuthenticated, login, logout, updateUser }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
