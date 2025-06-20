
import ShardLegendsGame from '@/components/shard-legends/ShardLegendsGame';

export default function GamePage() {
  return (
    // The main container for the game will inherit theme from html/body
    // No specific theme class needed here unless overriding body styles.
    <main className="min-h-screen"> 
      <ShardLegendsGame />
    </main>
  );
}
