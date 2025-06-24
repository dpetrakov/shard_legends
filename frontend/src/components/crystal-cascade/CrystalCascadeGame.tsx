
"use client";

import React, { useState, useCallback, useEffect, useRef } from 'react';
import GameBoardComponent from './GameBoard';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import type { FloatingScoreItem } from '@/types/crystal-cascade';
import type { ChestType } from '@/types/profile';
import FloatingScoreManager from './FloatingScoreEffect';
import { animate, motion, AnimatePresence } from "framer-motion";
import { useIconSet } from '@/contexts/IconSetContext';
import { useChests } from '@/contexts/ChestContext';
import Image from 'next/image';
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

const SCORE_STORAGE_KEY = 'crystalCascadeGameScore';
const MAX_COMBO_STORAGE_KEY = 'crystalCascadeMaxCombo';
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
  const [score, setScore] = useState(0);
  const [displayScore, setDisplayScore] = useState(0);
  const [gameKey, setGameKey] = useState(() => Date.now());
  const [, setHasPossibleMoves] = useState(true); 
  const [floatingScores, setFloatingScores] = useState<FloatingScoreItem[]>([]);

  const [comboCount, setComboCount] = useState(0);
  const comboCountRef = useRef(0);
  const [maxCombo, setMaxCombo] = useState(0);
  const [showComboText, setShowComboText] = useState(false);
  const [comboDisplayKey, setComboDisplayKey] = useState(0);
  const comboTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const { iconSet } = useIconSet();
  const { awardChest } = useChests();

  const scoreDisplayElementRef = useRef<HTMLDivElement>(null);
  const floatingScoreSpawnRef = useRef<HTMLDivElement>(null);

  const [isBoardFlipped, setIsBoardFlipped] = useState(false);
  const [revealedGameCardIndex, setRevealedGameCardIndex] = useState<number | null>(null);
  const [isRevealingFlippedCard, setIsRevealingFlippedCard] = useState<boolean>(false);
  const [revealedFlippedCardImageUrl, setRevealedFlippedCardImageUrl] = useState<string | null>(null);
  const [revealedChestType, setRevealedChestType] = useState<ChestType | null>(null);
  const [awaitingCardPick, setAwaitingCardPick] = useState(false);
  const [currentRewardComboRequirement, setCurrentRewardComboRequirement] = useState(5);


  useEffect(() => {
    const loadedScore = parseInt(localStorage.getItem(SCORE_STORAGE_KEY) || '0', 10);
    setScore(loadedScore);
    setDisplayScore(loadedScore);

    const loadedMaxCombo = parseInt(localStorage.getItem(MAX_COMBO_STORAGE_KEY) || '0', 10);
    setMaxCombo(loadedMaxCombo);

    const loadedRewardRequirement = parseInt(localStorage.getItem(REWARD_REQUIREMENT_STORAGE_KEY) || '5', 10);
    setCurrentRewardComboRequirement(loadedRewardRequirement);
    
    setComboCount(0);
    setShowComboText(false);
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setIsBoardFlipped(false);
    setAwaitingCardPick(false);
    setRevealedGameCardIndex(null);
    setIsRevealingFlippedCard(false);
    console.log(`[DEBUG] Game initialized/reset (gameKey: ${gameKey}). Score: ${loadedScore}, MaxCombo: ${loadedMaxCombo}, RewardReq: ${loadedRewardRequirement}`);
  }, [gameKey]);

  useEffect(() => {
    comboCountRef.current = comboCount;
     console.log(`[DEBUG] useEffect: comboCount updated to ${comboCount}. comboCountRef.current is now ${comboCountRef.current}`);
  }, [comboCount]);


  useEffect(() => {
    setGameKey(Date.now()); 
  }, [iconSet]);


  useEffect(() => {
    localStorage.setItem(SCORE_STORAGE_KEY, score.toString());
  }, [score]);

  useEffect(() => {
    localStorage.setItem(MAX_COMBO_STORAGE_KEY, maxCombo.toString());
  }, [maxCombo]);

  useEffect(() => {
    localStorage.setItem(REWARD_REQUIREMENT_STORAGE_KEY, currentRewardComboRequirement.toString());
  }, [currentRewardComboRequirement]);

  useEffect(() => {
    const controls = animate(displayScore, score, {
      duration: 0.4,
      ease: "easeOut",
      onUpdate: (latestValue) => setDisplayScore(Math.round(latestValue)),
    });
    return () => controls.stop();
  }, [score, displayScore]); 

  const handleScoreUpdate = useCallback((scoreIncrement: number) => {
    if (scoreIncrement === -1) { 
      setGameKey(Date.now());
      setHasPossibleMoves(true); 
      setFloatingScores([]); 
    } else {
      setScore(prevScore => Math.max(0, prevScore + scoreIncrement));
    }
  }, []);

  const handleCreateFloatingScore = useCallback((points: number) => {
    const newFloatingScore: FloatingScoreItem = {
      id: `${Date.now()}-${Math.random()}`,
      value: points,
      key: `${Date.now()}-${Math.random()}-key`, 
    };
    setFloatingScores(prevScores => [...prevScores, newFloatingScore]);
  }, []);

  const handleFloatingScoreAnimationComplete = useCallback((id: string) => {
    setFloatingScores(prevScores => prevScores.filter(scoreItem => scoreItem.id !== id));
  }, []);

  const handlePossibleMoveUpdate = useCallback((possible: boolean) => {
    setHasPossibleMoves(possible);
  }, []);

  const handleMatchProcessed = useCallback((numberOfDistinctGroupsInStep: number, isFirstStepInChain: boolean) => {
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);

    setComboCount(prevCombo => {
      const newComboBase = isFirstStepInChain ? 0 : prevCombo;
      const newCombo = newComboBase + numberOfDistinctGroupsInStep;
      if (newCombo > maxCombo) {
        setMaxCombo(newCombo);
      }
      console.log(`[DEBUG] handleMatchProcessed: newCombo: ${newCombo}, isFirstStep: ${isFirstStepInChain}, groups: ${numberOfDistinctGroupsInStep}`);
      return newCombo;
    });

    setShowComboText(true);
    setComboDisplayKey(k => k + 1); 

    comboTimeoutRef.current = setTimeout(() => {
      setShowComboText(false);
    }, 1800); 
  }, [maxCombo]);

  const handleNoMatchOrComboEnd = useCallback(() => {
    console.log(`[DEBUG] handleNoMatchOrComboEnd called. Current comboCount (from ref): ${comboCountRef.current}`);
    const finalComboCount = comboCountRef.current; 
    console.log(`[DEBUG] finalComboCount: ${finalComboCount}`);
    
    if (comboTimeoutRef.current) clearTimeout(comboTimeoutRef.current);
    setShowComboText(false);
    
    if (finalComboCount >= currentRewardComboRequirement && currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD) {
      console.log(`[DEBUG] Combo threshold met (${finalComboCount} >= ${currentRewardComboRequirement}). Flipping board.`);
      setTimeout(() => { 
        setIsBoardFlipped(true);
        setTimeout(() => { 
          setAwaitingCardPick(true);
        }, 600); 
      }, 500); 
    } else {
      console.log(`[DEBUG] Combo threshold NOT met (${finalComboCount} < ${currentRewardComboRequirement} or req > ${MAX_REWARD_COMBO_THRESHOLD}). Not flipping board.`);
    }
    setComboCount(0); 
  }, [currentRewardComboRequirement, setIsBoardFlipped, setAwaitingCardPick, setShowComboText]);
  

  const currentComboTextDisplay = comboCount > 1 ? `COMBO ${comboCount}x` : (comboCount === 1 ? "MATCH!" : "");
  const currentComboStyle = comboCount > 0
    ? comboStyles[Math.min(comboCount, comboStyles.length - 1)] 
    : comboStyles[0];

  const maxComboStyleColor = maxCombo > 0 && comboStyles[Math.min(maxCombo, comboStyles.length - 1)]
    ? comboStyles[Math.min(maxCombo, comboStyles.length - 1)].text
    : 'hsl(var(--foreground))';

  const handleGameCardClick = (cardIndex: number) => {
    if (isRevealingFlippedCard || !isBoardFlipped || !awaitingCardPick) {
      return;
    }
    setAwaitingCardPick(false); 
    setIsRevealingFlippedCard(true);
    setRevealedGameCardIndex(cardIndex);
    
    const chestWon = determineChestReward();
    awardChest(chestWon);
    setRevealedChestType(chestWon);
    
    const chestData = chestVisualData[chestWon];
    const placeholderText = chestData.name.replace(/ /g, '+');
    setRevealedFlippedCardImageUrl(`https://placehold.co/150x200/663399/FFFFFF.png?text=${placeholderText}`);

    const showDuration = 2500; 
    const animationDuration = 600; 

    setTimeout(() => {
      setIsRevealingFlippedCard(false); 
      
      setTimeout(() => {
        setRevealedGameCardIndex(null); 
        setRevealedFlippedCardImageUrl(null);
        setRevealedChestType(null);
        setIsBoardFlipped(false); 
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
    <div className="flex flex-col items-center justify-end px-2 h-full w-full relative">
      <div className="w-full max-w-md">
        <Card className="shadow-2xl bg-card/80 backdrop-blur-md">
          <CardHeader className="text-center items-center justify-center flex flex-col pt-4 pb-1">
            <div ref={scoreDisplayElementRef} className="flex items-center justify-center">
              <span className="font-code text-4xl text-foreground tracking-wider">
                {displayScore.toString().padStart(12, '0')}
              </span>
            </div>
          </CardHeader>
          <CardContent className="flex flex-col items-center text-center pt-0 pb-3 gap-1">
            <div className="flex items-center gap-1">
              <span className="text-muted-foreground text-xs">Max Combo: </span>
              <span
                className="font-headline text-base"
                style={{ color: maxComboStyleColor }}
              >
                {maxCombo > 0 ? `${maxCombo}x` : '0'}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground text-xs">
                {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD 
                  ? "Приз за:" 
                  : "Наград:"}
              </span>
              <span className="font-headline text-base text-primary">
                {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD
                  ? `Комбо ${currentRewardComboRequirement}+`
                  : "сегодня нет"}
              </span>
              {currentRewardComboRequirement <= MAX_REWARD_COMBO_THRESHOLD && (
                <Button 
                  onClick={handleResetRewardRequirement} 
                  variant="ghost" 
                  size="icon" 
                  className="h-5 w-5 text-muted-foreground hover:text-primary"
                  aria-label="Сбросить требование комбо"
                >
                  <RefreshCcw className="h-4 w-4" />
                </Button>
              )}
              <Button
                onClick={handleAddTestChests}
                variant="ghost"
                size="icon"
                className="h-5 w-5 text-muted-foreground hover:text-accent"
                aria-label="Добавить 10 сундуков каждого типа"
              >
                <PlusCircle className="h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        </Card>

        <div
          ref={floatingScoreSpawnRef}
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
          className="relative w-full mx-auto z-10"
          style={{ perspective: '1200px', aspectRatio: `${BOARD_COLS}/${BOARD_ROWS}` }}
        >
          <motion.div
            className="relative w-full h-full"
            style={{ transformStyle: 'preserve-3d' }}
            animate={{ rotateY: isBoardFlipped ? 180 : 0 }}
            transition={{ duration: 0.6 }}
          >
            <motion.div
              className="absolute inset-0 w-full h-full backface-hidden flex items-center justify-center"
            >
              <GameBoardComponent
                gameKeyProp={gameKey}
                onScoreUpdate={handleScoreUpdate}
                onPossibleMoveUpdate={handlePossibleMoveUpdate}
                onCreateFloatingScore={handleCreateFloatingScore}
                onMatchProcessed={handleMatchProcessed}
                onNoMatchOrComboEnd={handleNoMatchOrComboEnd}
                isProcessingExternally={isBoardFlipped || isRevealingFlippedCard}
              />
            </motion.div>

            <motion.div
              className="absolute inset-0 w-full h-full backface-hidden bg-card/90 backdrop-blur-sm p-2 sm:p-4 rounded-lg shadow-xl flex items-center justify-center"
              style={{ transform: 'rotateY(180deg)' }}
            >
              <div className="grid grid-cols-3 grid-rows-2 gap-2 sm:gap-3 w-full h-full">
                {[...Array(6)].map((_, index) => (
                  <Card
                    key={`game-card-wrapper-${index}`}
                    className="bg-transparent shadow-md flex flex-col items-center justify-center p-0 cursor-pointer overflow-hidden relative rounded-lg"
                    style={{ perspective: '1000px' }} 
                    onClick={() => handleGameCardClick(index)}
                    role="button"
                    tabIndex={0}
                    aria-label={revealedGameCardIndex === index && revealedFlippedCardImageUrl ? `Revealed Card ${index + 1}` : `Card ${index + 1}`}
                  >
                    <AnimatePresence initial={false} mode="wait">
                      {isRevealingFlippedCard && revealedGameCardIndex === index && revealedFlippedCardImageUrl ? (
                         <motion.div
                          key={`game-card-front-${index}`}
                          className="absolute inset-0 w-full h-full bg-secondary/90 rounded-lg flex flex-col items-center justify-center backface-hidden p-1"
                          initial={{ rotateY: -180 }}
                          animate={{ rotateY: 0 }}
                          exit={{ rotateY: 180 }}
                          transition={{ duration: 0.6 }}
                         >
                          <div className="relative w-full h-full">
                            <Image
                              src={revealedFlippedCardImageUrl} 
                              alt={`Открытая карта ${index + 1}`}
                              layout="fill"
                              objectFit="contain"
                              className="rounded-sm"
                              data-ai-hint={revealedChestType ? chestVisualData[revealedChestType]?.hint : "treasure card"} 
                            />
                          </div>
                        </motion.div>
                      ) : (
                        <motion.div
                          key={`game-card-back-${index}`}
                          className="absolute inset-0 w-full h-full bg-secondary/70 hover:bg-secondary/80 rounded-lg flex flex-col items-center justify-center backface-hidden p-1"
                          initial={{ rotateY: 0 }}
                          animate={{ rotateY: 0 }} 
                          exit={{ rotateY: -180 }} 
                          transition={{ duration: 0.6 }}
                        >
                          <div className="relative w-full h-full">
                             <Image
                              src="/images/card-back.png"
                              alt={`Карта ${index + 1}`}
                              layout="fill"
                              objectFit="contain"
                              className="rounded-sm"
                              data-ai-hint="card back"
                            />
                          </div>
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
          </motion.div>
        </div>
      </div>

      <FloatingScoreManager
        floatingScores={floatingScores}
        spawnElement={floatingScoreSpawnRef.current}
        scoreElement={scoreDisplayElementRef.current}
        onAnimationComplete={handleFloatingScoreAnimationComplete}
      />
    </div>
  );
};

export default CrystalCascadeGame;
