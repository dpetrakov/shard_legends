
"use client";

import React, { useState, useEffect, useCallback, useRef } from 'react';
import type { GameBoard, Position, CrystalIcon } from '@/types/crystal-cascade';
import CrystalCell from './CrystalCell';
import { BOARD_ROWS, BOARD_COLS } from './crystal-definitions';
import { isAdjacent, swapCrystals as logicalSwap, findMatchGroups, shiftAndFillCrystals, generateInitialBoard } from '@/lib/crystal-cascade-utils';
import { useIconSet } from '@/contexts/IconSetContext';

interface GameBoardProps {
  onScoreUpdate: (scoreIncrement: number) => void;
  onPossibleMoveUpdate: (hasPossibleMoves: boolean) => void;
  onCreateFloatingScore: (points: number) => void;
  onMatchProcessed: (numberOfDistinctGroupsInStep: number, isFirstStepInChain: boolean) => void;
  onNoMatchOrComboEnd: () => void;
  gameKeyProp: number;
  isProcessingExternally?: boolean;
}

const GameBoardComponent: React.FC<GameBoardProps> = ({
  onScoreUpdate,
  onPossibleMoveUpdate,
  onCreateFloatingScore,
  onMatchProcessed,
  onNoMatchOrComboEnd,
  gameKeyProp,
  isProcessingExternally,
}) => {
  const { getActiveIconList } = useIconSet();
  const [board, setBoard] = useState<GameBoard>([]);
  const [selectedCrystal, setSelectedCrystal] = useState<Position | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const activeIconsRef = useRef<CrystalIcon[]>(getActiveIconList());


  useEffect(() => {
    activeIconsRef.current = getActiveIconList();
    setBoard(generateInitialBoard(activeIconsRef.current));
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

  useEffect(() => {
    if (board.length === BOARD_ROWS &&
        board.every(row => Array.isArray(row) && row.length === BOARD_COLS &&
                         row.every(cell => cell === null || (typeof cell === 'object' && cell !== null && cell.type !== undefined))) &&
        !isProcessing && !isProcessingExternally) {
      const hasMoves = checkForPossibleMoves(board);
      onPossibleMoveUpdate(hasMoves);
    }
  }, [board, isProcessing, checkForPossibleMoves, onPossibleMoveUpdate, isProcessingExternally]);

  const processMatchesInternal = useCallback(async (currentBoard: GameBoard, isInitialPlayerMatchChain: boolean): Promise<GameBoard> => {
    let boardAfterMatches = currentBoard;
    let matchGroups = findMatchGroups(boardAfterMatches);
    let isFirstStepInChainLoop = isInitialPlayerMatchChain;

    while (matchGroups.length > 0) {
      onMatchProcessed(matchGroups.length, isFirstStepInChainLoop);
      isFirstStepInChainLoop = false;

      const allMatchedPositions = matchGroups.flat();
      const boardWithMarkedMatches = boardAfterMatches.map(row =>
        row.map(crystal => {
          if (crystal && allMatchedPositions.some(p => p.row === crystal.row && p.col === crystal.col)) {
            return { ...crystal, isMatched: true };
          }
          return crystal;
        })
      );
      setBoard(boardWithMarkedMatches);

      const pointsForThisMatch = allMatchedPositions.length * 10;
      onScoreUpdate(pointsForThisMatch);
      if (allMatchedPositions.length > 0) {
        onCreateFloatingScore(pointsForThisMatch);
      }

      await new Promise(resolve => setTimeout(resolve, 300)); // Duration for matched crystals to animate (e.g., fade out/shrink)

      const { newBoard: shiftedBoard } = shiftAndFillCrystals(boardAfterMatches, matchGroups, activeIconsRef.current);
      boardAfterMatches = shiftedBoard;
      setBoard(boardAfterMatches);

      await new Promise(resolve => setTimeout(resolve, 300)); // Duration for new crystals to fall and settle

      matchGroups = findMatchGroups(boardAfterMatches);
    }

    if (matchGroups.length === 0) { // Ensure this is called only once after all cascades
       onNoMatchOrComboEnd();
    }
    return boardAfterMatches;
  }, [onScoreUpdate, onCreateFloatingScore, onMatchProcessed, onNoMatchOrComboEnd]);

  const performSwapAndProcess = useCallback(async (pos1: Position, pos2: Position) => {
    if (isProcessing || isProcessingExternally) return;
    setIsProcessing(true);

    const tempSwappedBoard = logicalSwap(board, pos1, pos2);
    setBoard(tempSwappedBoard); // Update UI to show swapped crystals. `layout` in CrystalCell animates positions.
    
    // Wait for the visual swap animation to complete (or be clearly visible)
    await new Promise(resolve => setTimeout(resolve, 250)); 

    const matchGroupsFound = findMatchGroups(tempSwappedBoard);
    
    if (matchGroupsFound.length > 0) {
      // Matches found, proceed with match processing.
      // processMatchesInternal will handle further setBoard calls for match animations & refilling.
      const finalBoard = await processMatchesInternal(tempSwappedBoard, true);
      setBoard(finalBoard); // Set the board to the state after all cascades
    } else {
      // No matches found, animate swap back.
      // Board is currently showing tempSwappedBoard. Wait a bit for user to see the "no match" state.
      await new Promise(resolve => setTimeout(resolve, 200)); 
      
      const boardToSwapBack = logicalSwap(tempSwappedBoard, pos2, pos1);
      setBoard(boardToSwapBack); // Animate swap back
      
      // Wait for swap back animation to complete
      await new Promise(resolve => setTimeout(resolve, 250)); 
      
      onNoMatchOrComboEnd(); 
    }
    setIsProcessing(false);
  }, [board, isProcessing, processMatchesInternal, onNoMatchOrComboEnd, isProcessingExternally]);


  const handleCrystalClick = useCallback((position: Position) => {
    if (isProcessing || isProcessingExternally) return;

    const clickedCrystalOnBoard = board[position.row]?.[position.col];
    if (!clickedCrystalOnBoard) {
      setSelectedCrystal(null);
      return;
    }

    if (!selectedCrystal) {
      setSelectedCrystal(position);
    } else {
      if (selectedCrystal.row === position.row && selectedCrystal.col === position.col) {
        setSelectedCrystal(null);
      } else if (isAdjacent(selectedCrystal, position)) {
        const currentSelectedCrystalOnBoard = board[selectedCrystal.row]?.[selectedCrystal.col];
        if (currentSelectedCrystalOnBoard) {
           performSwapAndProcess(selectedCrystal, position);
        }
        setSelectedCrystal(null);
      } else {
        setSelectedCrystal(position); 
      }
    }
  }, [isProcessing, board, selectedCrystal, performSwapAndProcess, isProcessingExternally]);


  if (board.length !== BOARD_ROWS || !board.every(row => Array.isArray(row) && row.length === BOARD_COLS && row.every(cell => cell === null || (typeof cell === 'object' && cell !== null && 'id' in cell && 'type' in cell )))) {
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
          const cellKey = crystal ? `crystal-${crystal.id}` : `empty-${r}-${c}`;
          return (
            <CrystalCell
              key={cellKey}
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
