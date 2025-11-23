import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { PoliciesListTab } from './PoliciesListTab';
import { PermissionsPage } from './PermissionsPage';

export function PoliciesPage() {
  const [activeTab, setActiveTab] = useState('policies');

  return (
    <div className="p-8">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="policies">Policies</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>
        
        <TabsContent value="policies" className="mt-6">
          <PoliciesListTab />
        </TabsContent>
        
        <TabsContent value="permissions" className="mt-6">
          <PermissionsPage />
        </TabsContent>
      </Tabs>
    </div>
  );
}
