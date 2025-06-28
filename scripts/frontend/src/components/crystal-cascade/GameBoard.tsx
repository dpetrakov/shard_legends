
"use client";

import React, { useState, useEffect, useCallback } from 'react';
import type { GameBoard, Position, CrystalIcon } from '@/types/crystal-cascade';
import CrystalCell from './CrystalCell';
import { BOARD_ROWS, BOARD_COLS } from './crystal-definitions';
import { isAdjacent, swapCrystals as logicalSwap, findMatchGroups, shiftAndFillCrystals, generateInitialBoard } from '@/lib/crystal-cascade-utils';
import { useIconSet } from '@/contexts/IconSetContext';

interface GameBoardProps {
  onMatchProcessed: (numberOfDistinctGroupsInStep: number, isFirstStepInChain: boolean) => void;
  onNoMatchOrComboEnd: () => void;
  gameKeyProp: number;
  isProcessingExternally?: boolean;
}

const GameBoardComponent: React.FC<GameBoardProps> = ({
  onMatchProcessed,
  onNoMatchOrComboEnd,
  gameKeyProp,
  isProcessingExternally,
}) => {
  const { getActiveIconList } = useIconSet();
  const [board, setBoard] = useState<GameBoard>([]);
  const [selectedCrystal, setSelectedCrystal] = useState<Position | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  useEffect(() => {
    // This effect re-initializes the board when the game key or icon set changes.
    const activeIcons = getActiveIconList();
    if (activeIcons && activeIcons.length > 0) {
      const newBoard = generateInitialBoard(activeIcons);
      if (checkForPossibleMoves(newBoard)) {
        setBoard(newBoard);
      } else {
        // Retry generation if no moves are possible, to avoid a stuck board
        setBoard(generateInitialBoard(activeIcons));
      }
    }
    setSelectedCrystal(null);
    setIsProcessing(false);
  }, [getActiveIconList, gameKeyProp]);


  const checkForPossibleMoves = useCallback((currentBoard: GameBoard): boolean => {
    if (!currentBoard || currentBoard.length !== BOARD_ROWS) return false;
    for (let r = 0; r < BOARD_ROWS; r++) {
      if (!currentBoard[r] || currentBoard[r].length !== BOARD_COLS) return false;
      for (let c = 0; c < BOARD_COLS; c++) {
        const crystal = currentBoard[r][c];
        if (!crystal) continue;

        if (c < BOARD_COLS - 1) {
          const crystalToSwap = currentBoard[r][c+1];
          if (crystalToSwap) {
            const testBoard = logicalSwap(currentBoard, {r,c}, {r, c: c+1});
            if (findMatchGroups(testBoard).length > 0) return true;
          }
        }
        if (r < BOARD_ROWS - 1) {
           const crystalToSwap = currentBoard[r+1]?.[c];
           if (crystalToSwap) {
            const testBoard = logicalSwap(currentBoard, {r,c}, {r: r+1, c});
            if (findMatchGroups(testBoard).length > 0) return true;
          }
        }
      }
    }
    return false;
  }, []);

  const processMatchesAndRefill = useCallback((boardAfterSwap: GameBoard, isInitialPlayerMatch: boolean) => {
    let currentBoard = boardAfterSwap;
    let isFirstStepInChain = isInitialPlayerMatch;
    let chainEnded = false;
    const activeIcons = getActiveIconList();

    const processStep = () => {
      const matchGroups = findMatchGroups(currentBoard);

      if (matchGroups.length === 0) {
        if (!chainEnded) {
          onNoMatchOrComboEnd();
          chainEnded = true;
        }
        setIsProcessing(false);
        return;
      }
      
      onMatchProcessed(matchGroups.length, isFirstStepInChain);
      isFirstStepInChain = false;

      let boardWithMatchesRemoved = currentBoard.map(row => [...row]);
      matchGroups.flat().forEach(pos => {
        boardWithMatchesRemoved[pos.row][pos.col] = null;
      });
      setBoard(boardWithMatchesRemoved);

      // Wait for exit animation
      setTimeout(() => {
        const { newBoard: boardAfterRefill } = shiftAndFillCrystals(boardWithMatchesRemoved, [], activeIcons);
        setBoard(boardAfterRefill);
        currentBoard = boardAfterRefill;
        
        // Wait for fall animation then check for new matches
        setTimeout(processStep, 400); 
      }, 350);
    };

    processStep();

  }, [onMatchProcessed, onNoMatchOrComboEnd, getActiveIconList]);

  const performSwapAndProcess = useCallback((pos1: Position, pos2: Position) => {
    if (isProcessing || isProcessingExternally) return;
    
    setIsProcessing(true);
    
    const originalBoard = board;
    const swappedBoard = logicalSwap(board, pos1, pos2);
    setBoard(swappedBoard);
    
    // Wait for swap animation to finish
    setTimeout(() => {
      const matchGroupsFound = findMatchGroups(swappedBoard);
      
      if (matchGroupsFound.length > 0) {
        processMatchesAndRefill(swappedBoard, true);
      } else {
        // No match, swap back
        setBoard(originalBoard); // Re-render with original board state to trigger swap back animation
        setTimeout(() => {
          setIsProcessing(false);
          onNoMatchOrComboEnd();
        }, 300);
      }
    }, 300); // Duration should be close to the swap animation time

  }, [board, isProcessing, isProcessingExternally, processMatchesAndRefill, onNoMatchOrComboEnd]);


  const handleCrystalClick = useCallback((position: Position) => {
    if (isProcessing || isProcessingExternally) return;

    if (!selectedCrystal) {
      setSelectedCrystal(position);
    } else {
      if (selectedCrystal.row === position.row && selectedCrystal.col === position.col) {
        setSelectedCrystal(null);
      } else if (isAdjacent(selectedCrystal, position)) {
        performSwapAndProcess(selectedCrystal, position);
        setSelectedCrystal(null);
      } else {
        setSelectedCrystal(position); 
      }
    }
  }, [isProcessing, selectedCrystal, performSwapAndProcess, isProcessingExternally]);


  if (board.length === 0) {
    return <div className="text-center p-8 font-headline text-accent">Loading Game Board...</div>;
  }

  return (
    <div
      className="grid gap-px p-1 bg-primary/20 rounded-lg shadow-xl w-full h-full"
      style={{
        gridTemplateColumns: `repeat(${BOARD_COLS}, minmax(0, 1fr))`,
        gridTemplateRows: `repeat(${BOARD_ROWS}, minmax(0, 1fr))`,
      }}
      role="grid"
      aria-label="Crystal Cascade game board"
    >
      {board.map((row, r) =>
        row.map((crystal, c) => {
          const currentPosition = { row: r, col: c };
          return (
            <CrystalCell
              key={`${r}-${c}`} // Use stable cell key
              crystal={crystal}
              position={currentPosition}
              isSelected={!!selectedCrystal && selectedCrystal.row === currentPosition.row && selectedCrystal.col === currentPosition.col && !!crystal}
              onCrystalClick={handleCrystalClick}
            />
          );
        })
      )}
    </div>
  );
};

export default GameBoardComponent;
