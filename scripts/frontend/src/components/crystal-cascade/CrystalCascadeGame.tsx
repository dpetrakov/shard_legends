
"use client";

import React, { useState, useCallback, useEffect, useRef } from 'react';
import GameBoardComponent from './GameBoard';
import { Card, CardContent } from '@/components/ui/card';
import type { ChestType } from '@/types/profile';
import { motion, AnimatePresence } from "framer-motion";
import { useIconSet } from '@/contexts/IconSetContext';
import { useChests } from '@/contexts/ChestContext';
import { BOARD_COLS, BOARD_ROWS } from '@/components/crystal-cascade/crystal-definitions';
import { Button } from '@/components/ui/button';
import { PlusCircle, RefreshCcw } from 'lucide-react';
import { chestDetails as chestVisualData, allChestTypes } from '@/lib/chest-definitions';

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

const REWARD_REQUIREMENT_STORAGE_KEY = 'crystalCascadeRewardRequirement';
const MAX_REWARD_COMBO_THRESHOLD = 15; // Max combo requirement for reward

const determineChestReward = (): ChestType => {
  const categoryRand = Math.random();
  let category: 'resource' | 'reagent' | 'booster' | 'blueprint';

  if (categoryRand < 0.5) { // 50%
    category = 'resource';
  } else if (categoryRand < 0.8) { // 30% (0.5 + 0.3)
    category = 'reagent';
  } else if (categoryRand < 0.9) { // 10% (0.8 + 0.1)
    category = 'booster';
  } else { // 10%
    category = 'blueprint';
  }

  if (category === 'blueprint') {
    return 'blueprint';
  }

  const sizeRand = Math.random();
  let size: 'small' | 'medium' | 'large';

  if (sizeRand < 0.6) { // 60%
    size = 'small';
  } else if (sizeRand < 0.9) { // 30% (0.6 + 0.3)
    size = 'medium';
  } else { // 10%
    size = 'large';
  }

  return `${category}_${size}` as ChestType;
};


const CrystalCascadeGame: React.FC = () => {
  const [gameKey, setGameKey] = useState(() => Date.now());
  const [comboCount, setComboCount] = useState(0);
  const [showComboText, setShowComboText] = useState(false);
  const [comboDisplayKey, setComboDisplayKey] = useState(0);
  const comboTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const { iconSet } = useIconSet();
  const { awardChest } = useChests();
  
  const [viewState, setViewState] = useState<'board' | 'reward'>('board');
  const [revealedGameCardIndex, setRevealedGameCardIndex] = useState<number | null>(null);
  const [isRevealingFlippedCard, setIsRevealingFlippedCard] = useState<boolean>(false);
  const [revealedFlippedCardImageUrl, setRevealedFlippedCardImageUrl] = useState<string | null>(null);
  const [revealedChestType, setRevealedChestType] = useState<ChestType | null>(null);
  const [awaitingCardPick, setAwaitingCardPick] = useState(false);
  const [currentRewardComboRequirement, setCurrentRewardComboRequirement] = useState(5);
  const comboCountRef = useRef(comboCount);

  useEffect(() => {
    comboCountRef.current = comboCount;
  }, [comboCount]);

  useEffect(() => {
    const loadedRewardRequirement = parseInt(localStorage.getItem(REWARD_REQUIREMENT_STORAGE_KEY) || '5', 10);
    setCurrentRewardComboRequirement(loadedRewardRequirement);
    
    setComboCount(0);
    setShowComboText(false);
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setViewState('board');
    setAwaitingCardPick(false);
    setRevealedGameCardIndex(null);
    setIsRevealingFlippedCard(false);
    console.log(`[DEBUG] Game initialized/reset (gameKey: ${gameKey}). RewardReq: ${loadedRewardRequirement}`);
  }, [gameKey]);


  useEffect(() => {
    setGameKey(Date.now()); 
  }, [iconSet]);

  useEffect(() => {
    localStorage.setItem(REWARD_REQUIREMENT_STORAGE_KEY, currentRewardComboRequirement.toString());
  }, [currentRewardComboRequirement]);

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

  const handleNoMatchOrComboEnd = useCallback(() => {
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setShowComboText(false);
    
    // Use the ref here to get the latest value of comboCount
    if (comboCountRef.current >= currentRewardComboRequirement && currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD) {
      setTimeout(() => { 
        setViewState('reward');
        setTimeout(() => { 
          setAwaitingCardPick(true);
        }, 500); // Wait for slide animation to start before enabling pick
      }, 500); 
    }
    
    setComboCount(0); 
  }, [currentRewardComboRequirement]);
  

  const currentComboTextDisplay = comboCount > 1 ? `КОМБО ${comboCount}x` : (comboCount === 1 ? "ЕСТЬ!" : "");
  const currentComboStyle = comboCount > 0
    ? comboStyles[Math.min(comboCount, comboStyles.length - 1)] 
    : comboStyles[0];

  const handleGameCardClick = (cardIndex: number) => {
    if (isRevealingFlippedCard || viewState !== 'reward' || !awaitingCardPick) {
      return;
    }
    setAwaitingCardPick(false); 
    setIsRevealingFlippedCard(true);
    setRevealedGameCardIndex(cardIndex);
    
    const chestWon = determineChestReward();
    awardChest(chestWon);
    setRevealedChestType(chestWon);
    
    let imageUrl = '';
    const chestData = chestVisualData[chestWon];
    
    switch (chestWon) {
        case 'resource_small':
            imageUrl = '/images/card/card-small-res.jpg';
            break;
        case 'resource_medium':
            imageUrl = '/images/card/card-medium-res.jpg';
            break;
        case 'resource_large':
            imageUrl = '/images/card/card-big-res.jpg';
            break;
        case 'reagent_small':
            imageUrl = '/images/card/card-small-ing.jpg';
            break;
        case 'reagent_medium':
            imageUrl = '/images/card/card-medium-ing.jpg';
            break;
        case 'reagent_large':
            imageUrl = '/images/card/card-big-ing.jpg';
            break;
        case 'blueprint':
            imageUrl = '/images/card/card-blueprint.jpg';
            break;
        default:
            const placeholderText = chestData ? chestData.name.replace(/ /g, '+') : 'reward';
            imageUrl = `https://placehold.co/150x200/663399/FFFFFF.png?text=${placeholderText}`;
    }
    setRevealedFlippedCardImageUrl(imageUrl);

    const showDuration = 2500; 
    const animationDuration = 500; 

    setTimeout(() => {
      setViewState('board'); 
      
      setTimeout(() => {
        setIsRevealingFlippedCard(false); 
        setRevealedGameCardIndex(null); 
        setRevealedFlippedCardImageUrl(null);
        setRevealedChestType(null);
        if (currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD) {
            setCurrentRewardComboRequirement(prev => prev + 1);
        }
      }, animationDuration);

    }, showDuration); 
  };

  const handleResetRewardRequirement = () => {
    setCurrentRewardComboRequirement(5);
    console.log("[DEBUG] Reward requirement reset to 5 by test button.");
  };

  const handleAddTestChests = useCallback(() => {
    allChestTypes.forEach(chestType => {
      for (let i = 0; i < 10; i++) {
        awardChest(chestType);
      }
    });
    console.log("[DEBUG] Added 10 of each chest type for testing.");
  }, [awardChest]);


  return (
    <div className="flex flex-col items-center justify-center px-2 h-full w-full relative">
      <div className="w-full max-w-md">
        <Card className="shadow-2xl bg-card/80 backdrop-blur-md">
          <CardContent className="flex flex-col items-center text-center p-2 gap-1">
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground text-sm uppercase tracking-wider font-semibold">
                {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD 
                  ? "Приз за:" 
                  : "Наград:"}
              </span>
              <span className="font-headline text-lg text-primary">
                {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD
                  ? `Комбо ${currentRewardComboRequirement}+`
                  : "сегодня нет"}
              </span>
              {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD && (
                <Button 
                  onClick={handleResetRewardRequirement} 
                  variant="ghost" 
                  size="icon" 
                  className="h-6 w-6 text-muted-foreground hover:text-primary"
                  aria-label="Сбросить требование комбо"
                >
                  <RefreshCcw className="h-4 w-4" />
                </Button>
              )}
              <Button
                onClick={handleAddTestChests}
                variant="ghost"
                size="icon"
                className="h-6 w-6 text-muted-foreground hover:text-accent"
                aria-label="Добавить 10 сундуков каждого типа"
              >
                <PlusCircle className="h-4 w-4" />
              </Button>
            </div>
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
                      {isRevealingFlippedCard && revealedGameCardIndex === index && revealedFlippedCardImageUrl ? (
                         <motion.div
                          key={`game-card-front-${index}`}
                          className="relative w-full h-full"
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                          transition={{ duration: 0.3 }}
                         >
                          <img
                            src={revealedFlippedCardImageUrl} 
                            alt={`Открытая карта ${index + 1}`}
                            className="rounded-lg object-cover w-full h-full"
                          />
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
                           <img
                            src="/images/card/card-back.jpg"
                            alt={`Карта ${index + 1}`}
                            className="rounded-lg object-cover w-full h-full"
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
