"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function ShopPage() {
  return (
    <div className="flex flex-col items-center justify-start min-h-full p-4 space-y-6 text-foreground">
      <Card className="w-full max-w-md bg-card/80 backdrop-blur-md shadow-xl">
        <CardHeader>
          <CardTitle className="text-3xl font-headline text-center text-primary">Магазин</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="keys" className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-4">
              <TabsTrigger value="keys">Ключи</TabsTrigger>
              <TabsTrigger value="premium">Премиум</TabsTrigger>
            </TabsList>
            <TabsContent value="keys">
              <Card className="bg-background/50 shadow-inner">
                <CardHeader>
                  <CardTitle className="text-xl font-headline text-center text-accent">Ключи</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <p className="text-center text-muted-foreground">Раздел ключей находится в разработке.</p>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="premium">
              <Card className="bg-background/50 shadow-inner">
                <CardHeader>
                  <CardTitle className="text-xl font-headline text-center text-accent">Премиум</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
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
