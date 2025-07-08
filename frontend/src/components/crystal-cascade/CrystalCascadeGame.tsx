
"use client";

import React, { useState, useCallback, useEffect, useRef } from 'react';
import GameBoardComponent from './GameBoard';
import { Card, CardContent } from '@/components/ui/card';
import { motion, AnimatePresence } from "framer-motion";
import { useIconSet } from '@/contexts/IconSetContext';
import { BOARD_COLS, BOARD_ROWS } from '@/components/crystal-cascade/crystal-definitions';
import { useAuth } from '@/contexts/AuthContext';
import { useInventory } from '@/contexts/InventoryContext';
import { Skeleton } from '@/components/ui/skeleton';
import Image from 'next/image';
import { Button } from '@/components/ui/button';
import { PlusCircle } from 'lucide-react';

interface ComboStyle {
  background: string;
  text: string;
}

const comboStyles: ComboStyle[] = [
  { background: 'rgb(30, 30, 30)', text: 'rgb(220, 220, 220)' }, 
  { background: 'rgb(75, 0, 130)', text: 'rgb(230, 230, 250)' }, 
  { background: 'rgb(128, 0, 128)', text: 'rgb(240, 240, 240)' }, 
  { background: 'rgb(0, 0, 205)', text: 'rgb(240, 240, 240)' },   
  { background: 'rgb(0, 100, 0)', text: 'rgb(245, 245, 245)' },   
  { background: 'rgb(0, 191, 255)', text: 'rgb(20, 20, 20)' },    
  { background: 'rgb(0, 128, 128)', text: 'rgb(240, 240, 240)' }, 
  { background: 'rgb(50, 205, 50)', text: 'rgb(20, 20, 20)' },    
  { background: 'rgb(173, 255, 47)', text: 'rgb(15, 15, 15)' },   
  { background: 'rgb(255, 255, 0)', text: 'rgb(50, 50, 50)' },    
  { background: 'rgb(255, 215, 0)', text: 'rgb(40, 40, 40)' },    
  { background: 'rgb(255, 165, 0)', text: 'rgb(30, 30, 30)' },    
  { background: 'rgb(255, 140, 0)', text: 'rgb(240, 240, 240)'},   
  { background: 'rgb(255, 69, 0)', text: 'rgb(245, 245, 245)' },  
  { background: 'rgb(205, 0, 0)', text: 'rgb(245, 245, 245)' }    
];

interface RevealedCardData {
  imageUrl: string;
  altText: string;
}

// Light rays animation component, adapted for the card reveal
const LightRays = () => (
    <motion.div
        className="absolute inset-0 z-0"
        initial={{ opacity: 0, scale: 0.5 }}
        animate={{ 
            opacity: 1, 
            scale: 1.5,
            transition: { duration: 0.5, ease: 'easeOut' }
        }}
        exit={{ opacity: 0, scale: 0.8, transition: { duration: 0.3 } }}
    >
        <div className="absolute inset-0 animate-rays-spin-slow">
            {[...Array(8)].map((_, i) => (
                <div 
                    key={i}
                    className="absolute top-1/2 left-0 w-full h-px bg-gradient-to-r from-yellow-200/0 via-yellow-200 to-yellow-200/0"
                    style={{ transform: `rotate(${i * 22.5}deg)` }}
                />
            ))}
        </div>
        <div className="absolute inset-0 bg-yellow-300/20 rounded-full blur-2xl animate-pulse" />
    </motion.div>
);


const CrystalCascadeGame: React.FC = () => {
  const { token, isAuthenticated } = useAuth();
  const { syncWithServer: syncInventory } = useInventory();
  const apiUrl = 'https://dev-forly.slcw.dimlight.online';

  const [gameKey, setGameKey] = useState(() => Date.now());
  const [comboCount, setComboCount] = useState(0);
  const [showComboText, setShowComboText] = useState(false);
  const [comboDisplayKey, setComboDisplayKey] = useState(0);
  const comboTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const { iconSet } = useIconSet();
  
  const [dailyChestStatus, setDailyChestStatus] = useState<{ expected_combo: number | null; finished: boolean }>({
    expected_combo: null,
    finished: false,
  });
  const [isLoadingStatus, setIsLoadingStatus] = useState(true);
  
  const [viewState, setViewState] = useState<'board' | 'reward'>('board');
  const [revealedGameCardIndex, setRevealedGameCardIndex] = useState<number | null>(null);
  const [isRevealingFlippedCard, setIsRevealingFlippedCard] = useState<boolean>(false);
  const [showChestImage, setShowChestImage] = useState<boolean>(false);
  const [revealedCardData, setRevealedCardData] = useState<RevealedCardData | null>(null);
  const [awaitingCardPick, setAwaitingCardPick] = useState(false);
  
  const comboCountRef = useRef(comboCount);

  useEffect(() => {
    comboCountRef.current = comboCount;
  }, [comboCount]);

  const fetchStatus = useCallback(async () => {
    if (!isAuthenticated || !token) {
      return;
    }
    setIsLoadingStatus(true);
      
    try {
      const requestUrl = `${apiUrl}/api/deck/daily-chest/status`;
      const headers = { 
        'Authorization': `Bearer ${token}`,
        'Accept': 'application/json',
      };
      
      const response = await fetch(requestUrl, { 
        method: 'GET',
        mode: 'cors',
        headers 
      });
      
      const responseBodyText = await response.text();

      if (!response.ok) {
        throw new Error(`Failed to fetch daily chest status: ${responseBodyText}`);
      }
      
      if (!responseBodyText) {
          setDailyChestStatus({ expected_combo: null, finished: true });
          return;
      }
      
      let data;
      try {
          data = JSON.parse(responseBodyText);
      } catch (e) {
          console.error("Failed to parse daily chest status JSON:", e);
          setDailyChestStatus({ expected_combo: null, finished: true });
          return;
      }

      const expectedCombo = data.expected_combo ?? null;
      setDailyChestStatus({
        expected_combo: expectedCombo,
        finished: data.finished || expectedCombo === null,
      });

    } catch (error: any) {
      console.error("Fetch error:", error);
    } finally {
      setIsLoadingStatus(false);
    }
  }, [isAuthenticated, token, apiUrl]);

  useEffect(() => {
    if (isAuthenticated) {
        fetchStatus();
    }
  }, [isAuthenticated, gameKey, fetchStatus]);

  useEffect(() => {
    setComboCount(0);
    setShowComboText(false);
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setViewState('board');
    setAwaitingCardPick(false);
    setRevealedGameCardIndex(null);
    setIsRevealingFlippedCard(false);
    setShowChestImage(false);
    setRevealedCardData(null);
  }, [gameKey]);


  useEffect(() => {
    setGameKey(Date.now()); 
  }, [iconSet]);

  const handleMatchProcessed = useCallback((numberOfDistinctGroupsInStep: number, isFirstStepInChain: boolean) => {
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);

    setComboCount(prevCombo => {
      const newComboBase = isFirstStepInChain ? 0 : prevCombo;
      const newCombo = newComboBase + numberOfDistinctGroupsInStep;
      return newCombo;
    });

    setShowComboText(true);
    setComboDisplayKey(k => k + 1); 

    comboTimeoutRef.current = setTimeout(() => {
      setShowComboText(false);
    }, 1800); 
  }, []);

  const { expected_combo, finished } = dailyChestStatus;
  const handleNoMatchOrComboEnd = useCallback(() => {
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setShowComboText(false);
    
    if (expected_combo && comboCountRef.current >= expected_combo && !finished) {
      setTimeout(() => { 
        setViewState('reward');
        setTimeout(() => { 
          setAwaitingCardPick(true);
        }, 500);
      }, 500); 
    } else {
        setComboCount(0);
    }
    
  }, [expected_combo, finished]);
  

  const currentComboTextDisplay = comboCount > 1 ? `КОМБО ${comboCount}x` : (comboCount === 1 ? "ЕСТЬ!" : "");
  const currentComboStyle = comboCount > 0
    ? comboStyles[Math.min(comboCount, comboStyles.length - 1)] 
    : comboStyles[0];

  const handleGameCardClick = useCallback(async (cardIndex: number) => {
    if (isRevealingFlippedCard || viewState !== 'reward' || !awaitingCardPick || !expected_combo || !token) {
      return;
    }
    setAwaitingCardPick(false); 
    setIsRevealingFlippedCard(true);
    setShowChestImage(false);
    setRevealedGameCardIndex(cardIndex);
    
    try {
      const requestUrl = `${apiUrl}/api/deck/daily-chest/claim`;
      const requestBody = {
          combo: comboCountRef.current,
          chest_indices: [cardIndex],
        };
      const headers = {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        };

      const response = await fetch(requestUrl, {
        method: 'POST',
        mode: 'cors',
        headers,
        body: JSON.stringify(requestBody),
      });

      const responseBodyText = await response.text();
      
      if (!response.ok) {
        console.error("Failed to claim reward.", { body: responseBodyText });
        throw new Error(`Failed to claim reward.`);
      }

      if (!responseBodyText) {
          throw new Error('Claim reward response was empty');
      }

      let data;
      try {
          data = JSON.parse(responseBodyText);
      } catch (e) {
          console.error("Failed to parse claim reward JSON:", e);
          throw new Error('Invalid JSON in claim reward response');
      }
      
      const rewardItem = data.items?.[0];
      if (rewardItem) {
        const imageUrl = rewardItem.image_url
          ? `${apiUrl.replace('/api', '')}${rewardItem.image_url}`
          : ``;

        setRevealedCardData({
          imageUrl: imageUrl,
          altText: rewardItem.name || 'Награда'
        });
      } else {
        setRevealedCardData({
          imageUrl: '',
          altText: 'Пусто'
        });
      }
      
      setTimeout(() => {
        setShowChestImage(true);
      }, 3000);

      await syncInventory();

      const nextExpectedCombo = data.next_expected_combo ?? null;
      setDailyChestStatus({
        expected_combo: nextExpectedCombo,
        finished: nextExpectedCombo === null,
      });

    } catch (error: any) {
      console.error("Fetch error:", error);
      setTimeout(() => {
        setViewState('board');
        setComboCount(0);
      }, 1000);
    }

    const showDuration = 7000;
    const animationDuration = 500; 

    setTimeout(() => {
      setViewState('board'); 
      setComboCount(0);
      
      setTimeout(() => {
        setIsRevealingFlippedCard(false); 
        setRevealedGameCardIndex(null); 
        setRevealedCardData(null);
        setShowChestImage(false);
      }, animationDuration);

    }, showDuration); 
  }, [isRevealingFlippedCard, viewState, awaitingCardPick, expected_combo, token, apiUrl, syncInventory]);
  
  const handleDebugTriggerReward = () => {
    if (!dailyChestStatus.expected_combo || dailyChestStatus.finished) {
        console.log("Debug: No reward to trigger.");
        return;
    }
    console.log(`Debug: Triggering reward for combo ${dailyChestStatus.expected_combo}`);
    setComboCount(dailyChestStatus.expected_combo);
    setViewState('reward');
    setTimeout(() => {
        setAwaitingCardPick(true);
    }, 500);
  };
  
  const renderRewardStatus = () => {
    if (isLoadingStatus) {
      return <Skeleton className="h-6 w-48" />;
    }
    if (dailyChestStatus.finished) {
      return (
        <div className="flex items-center gap-2">
          <span className="text-muted-foreground text-sm uppercase tracking-wider font-semibold">
            Наград:
          </span>
          <span className="font-headline text-lg text-primary">
            сегодня нет
          </span>
        </div>
      );
    }
    return (
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground hover:text-primary" onClick={handleDebugTriggerReward}>
          <PlusCircle className="h-5 w-5" />
        </Button>
        <span className="text-muted-foreground text-sm uppercase tracking-wider font-semibold">
          Приз за:
        </span>
        <span className="font-headline text-lg text-primary">
          Комбо {dailyChestStatus.expected_combo}+
        </span>
      </div>
    );
  };


  return (
    <div className="flex flex-col items-center justify-center px-2 h-full w-full relative">
      <div className="w-full max-w-md space-y-2">
        <Card className="shadow-2xl bg-card/80 backdrop-blur-md">
          <CardContent className="flex flex-col items-center text-center p-2 gap-1 h-10 justify-center">
            {renderRewardStatus()}
          </CardContent>
        </Card>
        
        <div
          className="relative h-5 w-full flex items-center justify-center"
        >
          <AnimatePresence>
            {showComboText && comboCount > 0 && ( 
              <motion.div
                key={comboDisplayKey}
                initial={{ opacity: 0, y: 20, scale: 0.8 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -20, scale: 0.8, transition: { duration: 0.3 } }}
                className="absolute text-3xl md:text-4xl font-headline px-4 py-1 md:px-6 md:py-2 rounded-lg shadow-xl z-20"
                style={{
                  backgroundColor: currentComboStyle.background,
                  color: currentComboStyle.text,
                  textShadow: '1px 1px 2px rgba(0,0,0,0.2), 2px 2px 4px rgba(0,0,0,0.2)'
                }}
              >
                {currentComboTextDisplay}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
        
        <div 
          className="relative w-full mx-auto z-10 overflow-hidden"
          style={{ aspectRatio: `${BOARD_COLS}/${BOARD_ROWS}` }}
        >
            <motion.div
                className="relative w-full h-full"
                animate={{ x: viewState === 'board' ? '0%' : '-100%' }}
                transition={{ duration: 0.5, ease: 'easeInOut' }}
            >
                <GameBoardComponent
                    gameKeyProp={gameKey}
                    onMatchProcessed={handleMatchProcessed}
                    onNoMatchOrComboEnd={handleNoMatchOrComboEnd}
                    isProcessingExternally={viewState !== 'board'}
                />
            </motion.div>

            <motion.div
                className="absolute inset-0 w-full h-full bg-card/90 backdrop-blur-sm p-2 sm:p-4 rounded-lg shadow-xl"
                initial={{ x: '100%' }}
                animate={{ x: viewState === 'reward' ? '0%' : '100%' }}
                transition={{ duration: 0.5, ease: 'easeInOut' }}
            >
                <div className="grid grid-cols-3 grid-rows-2 gap-2 sm:gap-3 w-full h-full">
                {[...Array(6)].map((_, index) => (
                  <Card
                    key={`game-card-wrapper-${index}`}
                    className="bg-transparent shadow-md flex flex-col items-center justify-center p-0 cursor-pointer overflow-hidden relative rounded-lg"
                    onClick={() => handleGameCardClick(index)}
                    role="button"
                    tabIndex={0}
                    aria-label={`Card ${index + 1}`}
                  >
                    <AnimatePresence mode="wait">
                      {isRevealingFlippedCard && revealedGameCardIndex === index ? (
                         <motion.div
                            key={`revealed-card-container-${index}`}
                            className="w-full h-full flex items-center justify-center"
                            initial={{opacity: 0}}
                            animate={{opacity: 1}}
                            exit={{opacity: 0}}
                         >
                            <LightRays />
                            {showChestImage && revealedCardData && (
                                <motion.div
                                    key={`revealed-chest-${index}`}
                                    className="relative z-10 w-2/3 h-2/3"
                                    initial={{ opacity: 0, scale: 0.5 }}
                                    animate={{ opacity: 1, scale: 1, transition: { duration: 0.5, ease: 'backOut' } }}
                                >
                                    {revealedCardData.imageUrl ? (
                                        <Image
                                            src={revealedCardData.imageUrl}
                                            alt={revealedCardData.altText}
                                            layout="fill"
                                            objectFit="contain"
                                            className="drop-shadow-2xl"
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center text-white font-bold text-lg">
                                            {revealedCardData.altText}
                                        </div>
                                    )}
                                </motion.div>
                            )}
                        </motion.div>
                      ) : (
                        <motion.div
                          key={`game-card-back-${index}`}
                          className="relative w-full h-full"
                           initial={{ opacity: 1 }}
                           animate={{ opacity: 1 }}
                           exit={{ opacity: 0 }}
                           transition={{ duration: 0.3 }}
                        >
                           <Image
                            src="/images/card/card-back.jpg"
                            alt={`Карта ${index + 1}`}
                            layout="fill"
                            objectFit="cover"
                            className="rounded-lg"
                          />
                          {awaitingCardPick && !isRevealingFlippedCard && (
                             <span className="absolute bottom-2 text-xs text-center text-primary-foreground bg-black/60 px-2 py-0.5 rounded">
                               Выбери карту!
                             </span>
                          )}
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </Card>
                ))}
              </div>
            </motion.div>
        </div>
      </div>
    </div>
  );
};

export default CrystalCascadeGame;

