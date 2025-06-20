
"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function FactoryPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4 text-foreground pb-20">
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl font-headline text-center text-primary">Завод</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-center text-muted-foreground">Раздел завода находится в разработке.</p>
        </CardContent>
      </Card>
    </div>
  );
}
