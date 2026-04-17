'use client';

import React, { useState } from 'react';
import { Calendar, ChevronDown } from 'lucide-react';

interface DateRangeFilterProps {
  onApplyFilter: (startDate: Date | null, endDate: Date | null) => void;
  onClearFilter: () => void;
  isOpen: boolean;
  onToggle: () => void;
}

type RelativeDateRange = 'today' | 'yesterday' | 'last7days' | 'last30days' | 'last90days' | 'custom' | null;

export function DateRangeFilter({ onApplyFilter, onClearFilter, isOpen, onToggle }: DateRangeFilterProps) {
  const [selectedRange, setSelectedRange] = useState<RelativeDateRange>(null);
  const [customStartDate, setCustomStartDate] = useState<string>('');
  const [customEndDate, setCustomEndDate] = useState<string>('');

  const calculateDateRange = (range: RelativeDateRange): [Date | null, Date | null] => {
    if (!range) return [null, null];

    const now = new Date();
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());

    switch (range) {
      case 'today':
        return [startOfToday, now];
      case 'yesterday':
        const yesterday = new Date(startOfToday);
        yesterday.setDate(yesterday.getDate() - 1);
        return [yesterday, startOfToday];
      case 'last7days':
        const start7 = new Date(startOfToday);
        start7.setDate(start7.getDate() - 7);
        return [start7, now];
      case 'last30days':
        const start30 = new Date(startOfToday);
        start30.setDate(start30.getDate() - 30);
        return [start30, now];
      case 'last90days':
        const start90 = new Date(startOfToday);
        start90.setDate(start90.getDate() - 90);
        return [start90, now];
      case 'custom':
        const start = customStartDate ? new Date(customStartDate) : null;
        const end = customEndDate ? new Date(customEndDate) : null;
        return [start, end];
      default:
        return [null, null];
    }
  };

  const handleApplyFilter = () => {
    const [startDate, endDate] = calculateDateRange(selectedRange);
    onApplyFilter(startDate, endDate);
  };

  const handleClearFilter = () => {
    setSelectedRange(null);
    setCustomStartDate('');
    setCustomEndDate('');
    onClearFilter();
  };

  const isFilterActive = selectedRange !== null;

  return (
    <div className="relative">
      {/* Toggle Button */}
      <button
        onClick={onToggle}
        className={`inline-flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition-all ${
          isFilterActive
            ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
            : 'bg-slate-500/10 text-slate-400 border border-slate-500/20 hover:border-slate-500/40'
        }`}
      >
        <Calendar size={14} />
        <span>{isFilterActive ? `${selectedRange}` : 'Date Range'}</span>
        <ChevronDown size={12} className={`transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {/* Dropdown Panel */}
      {isOpen && (
        <div className="absolute top-full right-0 mt-2 w-80 bg-[#0F0F11] border border-[#1F1F22] rounded-lg shadow-lg z-50 p-4">
          {/* Preset Ranges */}
          <div className="mb-4">
            <p className="text-xs font-semibold text-slate-400 mb-2 uppercase tracking-wider">Quick Select</p>
            <div className="grid grid-cols-2 gap-2">
              {[
                { key: 'today', label: 'Today' },
                { key: 'yesterday', label: 'Yesterday' },
                { key: 'last7days', label: 'Last 7 Days' },
                { key: 'last30days', label: 'Last 30 Days' },
                { key: 'last90days', label: 'Last 90 Days' },
              ].map(({ key, label }) => (
                <button
                  key={key}
                  onClick={() => setSelectedRange(key as RelativeDateRange)}
                  className={`px-3 py-2 rounded text-xs font-medium transition-all ${
                    selectedRange === key
                      ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                      : 'bg-slate-500/5 text-slate-300 border border-slate-500/10 hover:border-slate-500/20'
                  }`}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>

          {/* Custom Date Range */}
          <div className="mb-4 pb-4 border-t border-[#1F1F22]">
            <p className="text-xs font-semibold text-slate-400 mt-4 mb-2 uppercase tracking-wider">Custom Range</p>
            <div className="flex flex-col gap-2">
              <div>
                <label className="text-xs text-slate-500 block mb-1">Start Date</label>
                <input
                  type="date"
                  value={customStartDate}
                  onChange={(e) => {
                    setCustomStartDate(e.target.value);
                    setSelectedRange('custom');
                  }}
                  className="w-full px-2 py-1.5 bg-[#1A1A1E] border border-[#27272A] rounded text-xs text-slate-300 focus:outline-none focus:border-blue-500/50"
                />
              </div>
              <div>
                <label className="text-xs text-slate-500 block mb-1">End Date</label>
                <input
                  type="date"
                  value={customEndDate}
                  onChange={(e) => {
                    setCustomEndDate(e.target.value);
                    setSelectedRange('custom');
                  }}
                  className="w-full px-2 py-1.5 bg-[#1A1A1E] border border-[#27272A] rounded text-xs text-slate-300 focus:outline-none focus:border-blue-500/50"
                />
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-2 pt-2 border-t border-[#1F1F22]">
            <button
              onClick={handleApplyFilter}
              disabled={selectedRange === null}
              className="flex-1 px-3 py-1.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded text-xs font-medium hover:bg-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              Apply
            </button>
            <button
              onClick={handleClearFilter}
              className="flex-1 px-3 py-1.5 bg-slate-500/10 text-slate-400 border border-slate-500/20 rounded text-xs font-medium hover:bg-slate-500/20 transition-all"
            >
              Clear
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
