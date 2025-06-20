
export interface RewardContextType {
  totalRewardPoints: number;
  addRewardPoints: (points: number) => void;
  spendRewardPoints: (points: number) => void;
}
