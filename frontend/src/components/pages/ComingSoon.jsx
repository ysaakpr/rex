import React from 'react';
import { Construction } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function ComingSoon({ title, description }) {
  return (
    <div className="flex items-center justify-center h-full">
      <Card className="w-full max-w-md text-center">
        <CardHeader>
          <div className="flex justify-center mb-4">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/30">
              <Construction className="h-10 w-10 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
          <CardTitle className="text-2xl">{title || 'Coming Soon'}</CardTitle>
          <CardDescription className="text-base mt-2">
            {description || 'This feature is currently under development and will be available in the next phase.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Stay tuned for updates!
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

