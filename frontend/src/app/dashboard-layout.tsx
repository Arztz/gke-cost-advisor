'use client';

import { useAppStore } from '@/stores/appStore';
import { cn } from '@/lib/utils';
import { Sidebar, Header } from '@/components/layout/Sidebar';

function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { sidebarOpen } = useAppStore();

  return (
    <div className="flex min-h-screen bg-slate-50 dark:bg-slate-950">
      <Sidebar />
      <div
        className={cn(
          'flex-1 transition-all duration-200',
          sidebarOpen ? 'lg:ml-64' : 'lg:ml-0'
        )}
      >
        <Header />
        <main className="p-4 lg:p-6">{children}</main>
      </div>
    </div>
  );
}

export default DashboardLayout;