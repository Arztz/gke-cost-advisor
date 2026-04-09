import { create } from 'zustand';
import type { Cluster, Recommendation, ComparisonRequest, ComparisonResult } from '@/types/api';

interface AppState {
  // Sidebar
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  
  // Selected cluster
  selectedClusterId: string | null;
  setSelectedClusterId: (id: string | null) => void;
  
  // Recommendations filters
  recommendationFilters: {
    type: string;
    priority: string;
    namespace: string;
  };
  setRecommendationFilters: (filters: Partial<AppState['recommendationFilters']>) => void;
  
  // Comparison state
  comparisonRequest: ComparisonRequest | null;
  setComparisonRequest: (request: ComparisonRequest | null) => void;
  comparisonResult: ComparisonResult | null;
  setComparisonResult: (result: ComparisonResult | null) => void;
  
  // Dark mode
  darkMode: boolean;
  setDarkMode: (dark: boolean) => void;
}

export const useAppStore = create<AppState>((set) => ({
  sidebarOpen: true,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  
  selectedClusterId: null,
  setSelectedClusterId: (id) => set({ selectedClusterId: id }),
  
  recommendationFilters: {
    type: '',
    priority: '',
    namespace: '',
  },
  setRecommendationFilters: (filters) =>
    set((state) => ({
      recommendationFilters: { ...state.recommendationFilters, ...filters },
    })),
  
  comparisonRequest: null,
  setComparisonRequest: (request) => set({ comparisonRequest: request }),
  comparisonResult: null,
  setComparisonResult: (result) => set({ comparisonResult: result }),
  
  darkMode: false,
  setDarkMode: (dark) => set({ darkMode: dark }),
}));
