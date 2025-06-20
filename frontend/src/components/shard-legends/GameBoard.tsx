
"use client";

import React, { useState, useEffect, useCallback, useRef } from 'react';
import type { GameBoard, Position, ShardIcon } from '@/types/shard-legends';
import ShardCell from './ShardCell';
import { BOARD_ROWS, BOARD_COLS } from './shard-definitions';
import { isAdjacent, swapShards as logicalSwap, findMatchGroups, shiftAndFillShards, generateInitialBoard } from '@/lib/shard-legends-utils';
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
  const [selectedShard, setSelectedShard] = useState<Position | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const activeIconsRef = useRef<ShardIcon[]>(getActiveIconList());


  useEffect(() => {
    activeIconsRef.current = getActiveIconList();
    setBoard(generateInitialBoard(activeIconsRef.current));
    setSelectedShard(null);
    setIsProcessing(false);
  }, [getActiveIconList, gameKeyProp]);


  const checkForPossibleMoves = useCallback((currentBoard: GameBoard): boolean => {
    if (!currentBoard || currentBoard.length !== BOARD_ROWS) return false;
    for (let r = 0; r < BOARD_ROWS; r++) {
      if (!currentBoard[r] || currentBoard[r].length !== BOARD_COLS) return false;
      for (let c = 0; c < BOARD_COLS; c++) {
        const shard = currentBoard[r][c];
        if (!shard) continue;

        if (c < BOARD_COLS - 1) {
          const shardToSwap = currentBoard[r][c+1];
          if (shardToSwap) {
            const testBoard = logicalSwap(currentBoard, {r,c}, {r, c: c+1});
            if (findMatchGroups(testBoard).length > 0) return true;
          }
        }
        if (r < BOARD_ROWS - 1) {
           const shardToSwap = currentBoard[r+1]?.[c];
           if (shardToSwap) {
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
        row.map(shard => {
          if (shard && allMatchedPositions.some(p => p.row === shard.row && p.col === shard.col)) {
            return { ...shard, isMatched: true };
          }
          return shard;
        })
      );
      setBoard(boardWithMarkedMatches);

      const pointsForThisMatch = allMatchedPositions.length * 10;
      onScoreUpdate(pointsForThisMatch);
      if (allMatchedPositions.length > 0) {
        onCreateFloatingScore(pointsForThisMatch);
      }

      await new Promise(resolve => setTimeout(resolve, 300));

      const { newBoard: shiftedBoard } = shiftAndFillShards(boardAfterMatches, matchGroups, activeIconsRef.current);
      boardAfterMatches = shiftedBoard;
      setBoard(boardAfterMatches);

      await new Promise(resolve => setTimeout(resolve, 300));

      matchGroups = findMatchGroups(boardAfterMatches);
    }

    if (matchGroups.length === 0) {
       onNoMatchOrComboEnd();
    }
    return boardAfterMatches;
  }, [onScoreUpdate, onCreateFloatingScore, onMatchProcessed, onNoMatchOrComboEnd]);

  const performSwapAndProcess = useCallback(async (pos1: Position, pos2: Position) => {
    if (isProcessing || isProcessingExternally) return;
    setIsProcessing(true);

    const tempSwappedBoard = logicalSwap(board, pos1, pos2);
    
    const matchGroupsFound = findMatchGroups(tempSwappedBoard);
    if (matchGroupsFound.length > 0) {
      setBoard(tempSwappedBoard); 
      const finalBoard = await processMatchesInternal(tempSwappedBoard, true);
      setBoard(finalBoard);
    } else {
      setBoard(tempSwappedBoard); 
      await new Promise(resolve => setTimeout(resolve, 150)); 
      setBoard(logicalSwap(tempSwappedBoard, pos2, pos1)); 
      onNoMatchOrComboEnd(); 
      await new Promise(resolve => setTimeout(resolve, 300)); 
    }
    setIsProcessing(false);
  }, [board, isProcessing, processMatchesInternal, onNoMatchOrComboEnd, isProcessingExternally]);


  const handleShardClick = useCallback((position: Position) => {
    if (isProcessing || isProcessingExternally) return;

    const clickedShardOnBoard = board[position.row]?.[position.col];
    if (!clickedShardOnBoard) {
      setSelectedShard(null);
      return;
    }

    if (!selectedShard) {
      setSelectedShard(position);
    } else {
      if (selectedShard.row === position.row && selectedShard.col === position.col) {
        setSelectedShard(null);
      } else if (isAdjacent(selectedShard, position)) {
        const currentSelectedShardOnBoard = board[selectedShard.row]?.[selectedShard.col];
        if (currentSelectedShardOnBoard) {
           performSwapAndProcess(selectedShard, position);
        }
        setSelectedShard(null);
      } else {
        setSelectedShard(position); 
      }
    }
  }, [isProcessing, board, selectedShard, performSwapAndProcess, isProcessingExternally]);


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
      aria-label="Shard Legends game board"
    >
      {board.map((row, r) =>
        row.map((shard, c) => {
          const currentPosition = { row: r, col: c };
          const cellKey = shard ? `shard-${shard.id}-${r}-${c}` : `empty-${r}-${c}`;
          return (
            <ShardCell
              key={cellKey}
              shard={shard}
              position={currentPosition}
              isSelected={!!selectedShard && selectedShard.row === currentPosition.row && selectedShard.col === currentPosition.col && !!shard}
              onShardClick={handleShardClick}
            />
          );
        })
      )}
    </div>
  );
};

export default GameBoardComponent;
