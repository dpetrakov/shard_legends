
import { Loader2 } from 'lucide-react';

export default function FactoryLoading() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4 text-foreground pb-20">
      <Loader2 className="h-12 w-12 animate-spin text-primary" />
      <p className="text-xl font-headline text-muted-foreground mt-4">Загрузка завода...</p>
    </div>
  );
}
