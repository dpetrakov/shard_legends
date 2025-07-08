
"use client";

import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Gift, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';

// A single digit reel for the slot machine effect
const Reel = ({ digit }: { digit: string }) => (
    <div className="w-12 h-16 sm:w-16 sm:h-20 bg-background/50 rounded-lg flex items-center justify-center shadow-inner overflow-hidden">
        <span className="text-4xl sm:text-5xl font-bold text-primary text-shadow-lg font-mono">
            {digit}
        </span>
    </div>
);

export default function GiftsPage() {
    const bankAmount = 1000000;
    const [digits, setDigits] = useState<string[]>(Array(6).fill('0'));
    const [isSpinning, setIsSpinning] = useState(false);
    const [prize, setPrize] = useState<number | null>(null);

    const handleSpin = () => {
        if (isSpinning) return;

        setIsSpinning(true);
        setPrize(null);

        // Calculate prize upfront
        const minPercent = 0.0000001; // 0.00001%
        const maxPercent = 0.001;     // 0.1%
        const randomPercent = Math.random() * (maxPercent - minPercent) + minPercent;
        const wonAmount = Math.floor(bankAmount * randomPercent);
        const prizeString = wonAmount.toString().padStart(6, '0');

        const totalSpinDuration = 3000; // Total duration in ms
        const reelStopDelay = 400;    // Delay between each reel stopping
        const digitChangeInterval = 50; // How fast digits flicker
        const initialSpinTime = totalSpinDuration - (reelStopDelay * 6); // Time before first reel stops

        const reelIntervals: (NodeJS.Timeout | null)[] = Array(6).fill(null);

        // Start all reels spinning
        for (let i = 0; i < 6; i++) {
            reelIntervals[i] = setInterval(() => {
                setDigits(prev => {
                    const newDigits = [...prev];
                    newDigits[i] = Math.floor(Math.random() * 10).toString();
                    return newDigits;
                });
            }, digitChangeInterval);
        }

        // Schedule stops for each reel from right to left
        for (let i = 5; i >= 0; i--) {
            const stopTime = initialSpinTime + (reelStopDelay * (5 - i));
            setTimeout(() => {
                if (reelIntervals[i]) {
                    clearInterval(reelIntervals[i]!);
                }
                setDigits(prev => {
                    const newDigits = [...prev];
                    newDigits[i] = prizeString[i];
                    return newDigits;
                });
            }, stopTime);
        }

        // Final cleanup after the last reel has stopped
        setTimeout(() => {
            reelIntervals.forEach(interval => interval && clearInterval(interval));
            setPrize(wonAmount);
            setIsSpinning(false);
        }, totalSpinDuration);
    };

    return (
        <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
            {/* Bank Info Panel */}
            <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl text-center">
                <CardHeader>
                    <CardTitle className="text-xl font-headline text-primary">Сумма банка</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-4xl font-bold text-foreground text-shadow-md">
                        {bankAmount.toLocaleString()} <span className="text-2xl text-primary font-semibold">$SLCW</span>
                    </p>
                </CardContent>
            </Card>

            {/* Main Drum Card */}
            <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
                <CardHeader className="text-center">
                    <div className="flex justify-center items-center gap-4">
                        <Gift className="w-8 h-8 text-primary" />
                        <CardTitle className="text-3xl font-headline text-primary">Подарки</CardTitle>
                    </div>
                </CardHeader>
                <CardContent className="flex flex-col items-center space-y-6">
                    {/* The Drum */}
                    <div className="flex items-center justify-center gap-2">
                        {digits.map((digit, index) => (
                            <Reel key={index} digit={digit} />
                        ))}
                    </div>

                    {/* Spin Button */}
                    <Button 
                        size="lg" 
                        className="w-full font-bold text-lg"
                        onClick={handleSpin}
                        disabled={isSpinning}
                    >
                        {isSpinning ? (
                            <RefreshCw className="mr-2 h-5 w-5 animate-spin" />
                        ) : null}
                        {isSpinning ? 'Крутится...' : 'Крутить'}
                    </Button>
                    
                    {/* Prize Display */}
                    <div className="h-8">
                        <AnimatePresence>
                            {prize !== null && !isSpinning && (
                                <motion.div
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    exit={{ opacity: 0, y: -20 }}
                                    className="text-center"
                                >
                                    <p className="text-lg font-semibold text-accent">Ваш выигрыш:</p>
                                    <p className="text-2xl font-bold text-primary">
                                        {prize.toLocaleString()} <span className="text-lg font-semibold">$SLCW</span>
                                    </p>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
