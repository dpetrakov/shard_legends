
"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
// import { useRewards } from "@/contexts/RewardContext"; // Removed
// import { useChests } from "@/contexts/ChestContext"; // Removed for this page's direct use
// import type { ChestType } from "@/types/profile"; // Removed if not used elsewhere
// import Image from 'next/image'; // Removed
// import { Coins } from "lucide-react"; // Removed
// import { useState } from "react"; // Removed
// import { motion, AnimatePresence } from 'framer-motion'; // Removed

// const chestVisualData: Record<ChestType, { name: string; hint: string }> = {
//   small: { name: "Малый", hint: "small treasure" },
//   medium: { name: "Средний", hint: "medium treasure" },
//   large: { name: "Большой", hint: "large treasure" }
// }; // Moved to CrystalCascadeGame

export default function ShopPage() {
  // const { totalRewardPoints, spendRewardPoints } = useRewards(); // Removed
  // const { awardChest } = useChests(); // Removed

  // const [revealedChest, setRevealedChest] = useState<ChestType | null>(null); // Removed
  // const [isRevealingCardIndex, setIsRevealingCardIndex] = useState<number | null>(null); // Removed

  // const determineChestReward = (): ChestType => { ... }; // Moved to CrystalCascadeGame
  // const handleCardClick = (cardIndex: number) => { ... }; // Removed

  return (
    <div className="flex flex-col items-center justify-start min-h-screen p-4 pt-6 text-foreground pb-20 space-y-6">
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-3xl font-headline text-center text-primary">Магазин</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="keys" className="w-full"> {/* Default to "keys" or other existing tab */}
            <TabsList className="grid w-full grid-cols-2 mb-4"> {/* Adjusted grid-cols */}
              {/* <TabsTrigger value="reward">Награда</TabsTrigger> */} {/* Removed */}
              <TabsTrigger value="keys">Ключи</TabsTrigger>
              <TabsTrigger value="premium">Премиум</TabsTrigger>
            </TabsList>
            {/* <TabsContent value="reward"> ... </TabsContent> */} {/* Removed */}
            <TabsContent value="keys">
              <Card className="bg-background/50 shadow-inner">
                <CardHeader>
                  <CardTitle className="text-xl font-headline text-center text-accent">Ключи</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-center text-muted-foreground">Раздел ключей находится в разработке.</p>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="premium">
              <Card className="bg-background/50 shadow-inner">
                <CardHeader>
                  <CardTitle className="text-xl font-headline text-center text-accent">Премиум</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-center text-muted-foreground">Премиум раздел находится в разработке.</p>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
