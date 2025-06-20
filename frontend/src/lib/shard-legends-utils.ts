
import type { GameBoard, Shard, Position, ShardIcon } from '@/types/shard-legends';
import { BOARD_ROWS, BOARD_COLS } from '@/components/shard-legends/shard-definitions';

let nextShardId = 0;

export function createShard(row: number, col: number, icons: ShardIcon[], specificTypeIndex?: number): Shard {
  const typeIndex = specificTypeIndex !== undefined ? specificTypeIndex : Math.floor(Math.random() * icons.length);
  if (typeIndex < 0 || typeIndex >= icons.length) {
    console.error("Invalid typeIndex in createShard", typeIndex, icons.length);
    // Fallback to a valid index, e.g., the first icon
    return {
        id: nextShardId++,
        type: icons[0],
        row,
        col,
        isMatched: false,
      };
  }
  return {
    id: nextShardId++,
    type: icons[typeIndex],
    row,
    col,
    isMatched: false,
  };
}

export function generateInitialBoard(icons: ShardIcon[]): GameBoard {
  let board: GameBoard = [];
  nextShardId = 0;

  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  while (true) {
    board = [];
    for (let r = 0; r < BOARD_ROWS; r++) {
      board[r] = [];
      for (let c = 0; c < BOARD_COLS; c++) {
        const shard = createShard(r, c, icons);
        board[r][c] = shard;
      }
    }

    const initialMatchGroups = findMatchGroups(board);
    if (initialMatchGroups.length === 0) {
      // Ensure all shards have correct row/col before returning
      for (let r = 0; r < BOARD_ROWS; r++) {
        for (let c = 0; c < BOARD_COLS; c++) {
          const shard = board[r][c];
          if (shard) {
            shard.row = r;
            shard.col = c;
          }
        }
      }
      break;
    }
  }
  return board;
}


export function isAdjacent(pos1: Position, pos2: Position): boolean {
  const rowDiff = Math.abs(pos1.row - pos2.row);
  const colDiff = Math.abs(pos1.col - pos2.col);
  return (rowDiff === 1 && colDiff === 0) || (rowDiff === 0 && colDiff === 1);
}

export function swapShards(board: GameBoard, pos1: Position, pos2: Position): GameBoard {
  if (
    !board || !Array.isArray(board) || board.length !== BOARD_ROWS ||
    !board.every(row => Array.isArray(row) && row.length === BOARD_COLS) ||
    pos1.row < 0 || pos1.row >= BOARD_ROWS || pos1.col < 0 || !board[pos1.row] || pos1.col >= board[pos1.row].length ||
    pos2.row < 0 || pos2.row >= BOARD_ROWS || pos2.col < 0 || !board[pos2.row] || pos2.col >= board[pos2.row].length
  ) {
    const safeCopy: GameBoard = [];
    for (let r = 0; r < BOARD_ROWS; r++) {
      safeCopy[r] = [];
      for (let c = 0; c < BOARD_COLS; c++) {
        const originalCell = board?.[r]?.[c];
        // This part needs to be careful if icons array is not available here for createShard
        // For now, just copy or null. If createShard is needed, icons must be passed.
        safeCopy[r][c] = (originalCell && typeof originalCell === 'object' && 'id' in originalCell) ? { ...originalCell, row:r, col:c } : null;
      }
    }
    return safeCopy;
  }

  const newBoard = board.map(row => row.map(crystal => (crystal ? {...crystal} : null)));
  const crystal1 = newBoard[pos1.row][pos1.col];
  const crystal2 = newBoard[pos2.row][pos2.col];

  if (crystal1) { crystal1.row = pos2.row; crystal1.col = pos2.col; }
  if (crystal2) { crystal2.row = pos1.row; crystal2.col = pos1.col; }

  newBoard[pos1.row][pos1.col] = crystal2;
  newBoard[pos2.row][pos2.col] = crystal1;
  return newBoard;
}

export function findMatchGroups(board: GameBoard): Position[][] {
  const allGroups: Position[][] = [];
  const cellsInAGroup: Set<string> = new Set();

  for (let r = 0; r < BOARD_ROWS; r++) {
    for (let c = 0; c < BOARD_COLS; ) {
      const currentCrystal = board[r]?.[c]; // Added optional chaining
      if (!currentCrystal || !currentCrystal.type) {
        c++;
        continue;
      }

      let matchLength = 1;
      for (let k = c + 1; k < BOARD_COLS; k++) {
        if (board[r]?.[k]?.type?.name === currentCrystal.type.name) {
          matchLength++;
        } else {
          break;
        }
      }

      if (matchLength >= 3) {
        const currentGroupToAdd: Position[] = [];
        let distinctGroup = false;
        for(let i=0; i < matchLength; ++i) {
           const posKey = `${r}-${c+i}`;
           if(!cellsInAGroup.has(posKey)) {
               distinctGroup = true;
           }
           currentGroupToAdd.push({row: r, col: c+i});
        }
        if(distinctGroup) {
            allGroups.push(currentGroupToAdd);
            currentGroupToAdd.forEach(p => cellsInAGroup.add(`${p.row}-${p.col}`));
        }
        c += matchLength;
      } else {
        c++;
      }
    }
  }

  // const cellsInVerticalGroupThisPass: Set<string> = new Set(); // Not used, can be removed
  for (let c = 0; c < BOARD_COLS; c++) {
    for (let r = 0; r < BOARD_ROWS; ) {
      const currentCrystal = board[r]?.[c]; // Added optional chaining
      if (!currentCrystal || !currentCrystal.type) {
        r++;
        continue;
      }

      let matchLength = 1;
      for (let k = r + 1; k < BOARD_ROWS; k++) {
        if (board[k]?.[c]?.type?.name === currentCrystal.type.name) {
          matchLength++;
        } else {
          break;
        }
      }

      if (matchLength >= 3) {
        const currentGroupToAdd: Position[] = [];
        let distinctGroup = false;
        for(let i=0; i < matchLength; ++i) {
            const posKey = `${r+i}-${c}`;
            if(!cellsInAGroup.has(posKey)) {
                 distinctGroup = true;
            }
            currentGroupToAdd.push({row: r+i, col: c});
        }
        if(distinctGroup) {
            allGroups.push(currentGroupToAdd);
            currentGroupToAdd.forEach(p => {
              cellsInAGroup.add(`${p.row}-${p.col}`);
              // cellsInVerticalGroupThisPass.add(`${p.row}-${p.col}`);
            });
        }
        r += matchLength;
      } else {
        r++;
      }
    }
  }
  return allGroups;
}


export function shiftAndFillShards(initialBoard: GameBoard, matchGroups: Position[][], icons: ShardIcon[] ): { newBoard: GameBoard, newScore: number } {
  const allMatchedPositionsSet = new Set<string>();
  matchGroups.flat().forEach(p => allMatchedPositionsSet.add(`${p.row}-${p.col}`));

  const workingBoard: GameBoard = initialBoard.map((row, r) => {
    if (!Array.isArray(row) || row.length !== BOARD_COLS) {
      // Ensure icons array is available and used if creating new crystals here
      return new Array(BOARD_COLS).fill(null).map((_, cIdx) => createShard(r, cIdx, icons));
    }
    return row.map(crystal => crystal ? { ...crystal, isMatched: false } : null);
  });

  allMatchedPositionsSet.forEach(posKey => {
    const [r, c] = posKey.split('-').map(Number);
    if (workingBoard[r]?.[c]) {
      workingBoard[r][c] = null;
    }
  });

  for (let c = 0; c < BOARD_COLS; c++) {
    let writeRow = BOARD_ROWS - 1;
    for (let r = BOARD_ROWS - 1; r >= 0; r--) {
      if (workingBoard[r]?.[c] !== null) {
        const crystalToMove = workingBoard[r][c]!;
        if (r !== writeRow) {
          workingBoard[writeRow][c] = crystalToMove;
          workingBoard[r][c] = null;
        }
        if (crystalToMove) {
            crystalToMove.row = writeRow;
            crystalToMove.col = c;
        }
        writeRow--;
      }
    }
  }

  for (let r = 0; r < BOARD_ROWS; r++) {
    if (!workingBoard[r]) {
        workingBoard[r] = new Array(BOARD_COLS).fill(null);
    }
    for (let c = 0; c < BOARD_COLS; c++) {
      if (workingBoard[r][c] === null) {
        workingBoard[r][c] = createShard(r, c, icons);
      }
    }
  }

  const finalBoard: GameBoard = workingBoard.map((row, r) =>
    row.map((crystal, c) => {
      if (crystal) {
        return { ...crystal, row: r, col: c, isMatched: false };
      }
      // This case should ideally not be hit if logic above is correct, but as a fallback:
      return createShard(r,c, icons);
    })
  );

  return { newBoard: finalBoard, newScore: 0 }; // score calculation removed for brevity, it's handled in GameBoardComponent
}
