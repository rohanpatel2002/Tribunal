'use client';

import React, { useState, useEffect } from 'react';
import { GitPullRequest, ChevronDown, Check, Plus } from 'lucide-react';
import { clsx } from 'clsx';

interface Repository {
  id: string;
  name: string;
  fullName: string;
  url: string;
  isMonitored: boolean;
}

interface RepositorySelectorProps {
  selectedRepo: string;
  onSelectRepo: (repoName: string, fullName: string) => void;
  repositories?: Repository[];
  isLoading?: boolean;
}

export function RepositorySelector({
  selectedRepo,
  onSelectRepo,
  repositories = [],
  isLoading = false,
}: RepositorySelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredRepos, setFilteredRepos] = useState<Repository[]>(repositories);

  useEffect(() => {
    const filtered = repositories.filter(
      (repo) =>
        repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        repo.fullName.toLowerCase().includes(searchQuery.toLowerCase())
    );
    setFilteredRepos(filtered);
  }, [searchQuery, repositories]);

  const currentRepo = repositories.find((r) => r.fullName === selectedRepo);

  return (
    <div className="relative w-full max-w-sm">
      {/* Selector Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between px-3 py-2 bg-[#1A1A1E] border border-[#27272A] rounded-lg text-sm text-slate-300 hover:border-[#3F3F46] transition-all group"
      >
        <div className="flex items-center gap-2 min-w-0">
          <GitPullRequest size={16} className="text-slate-500 flex-shrink-0" />
          <span className="truncate">
            {currentRepo?.name || 'Select Repository'}
          </span>
        </div>
        <ChevronDown
          size={16}
          className={clsx(
            'text-slate-500 flex-shrink-0 transition-transform',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-[#0F0F11] border border-[#1F1F22] rounded-lg shadow-xl z-50 overflow-hidden">
          {/* Search Input */}
          <div className="p-2 border-b border-[#1F1F22]">
            <input
              type="text"
              placeholder="Search repositories..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-2 py-1.5 bg-[#1A1A1E] border border-[#27272A] rounded text-xs text-slate-300 placeholder-slate-600 focus:outline-none focus:border-indigo-500/50"
            />
          </div>

          {/* Repository List */}
          <div className="max-h-64 overflow-y-auto">
            {isLoading ? (
              <div className="px-3 py-8 text-center text-xs text-slate-500">
                Loading repositories...
              </div>
            ) : filteredRepos.length === 0 ? (
              <div className="px-3 py-8 text-center text-xs text-slate-500">
                No repositories found
              </div>
            ) : (
              filteredRepos.map((repo) => (
                <button
                  key={repo.id}
                  onClick={() => {
                    onSelectRepo(repo.name, repo.fullName);
                    setIsOpen(false);
                    setSearchQuery('');
                  }}
                  className={clsx(
                    'w-full px-3 py-2 text-left text-xs transition-colors flex items-center justify-between group',
                    selectedRepo === repo.fullName
                      ? 'bg-indigo-500/10 text-indigo-400'
                      : 'text-slate-300 hover:bg-[#1A1A1E]'
                  )}
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <GitPullRequest size={14} className="text-slate-500 flex-shrink-0" />
                    <div className="min-w-0">
                      <p className="font-medium truncate">{repo.name}</p>
                      <p className="text-[10px] text-slate-500 truncate">{repo.fullName}</p>
                    </div>
                  </div>
                  {selectedRepo === repo.fullName && (
                    <Check size={14} className="text-indigo-400 flex-shrink-0 ml-2" />
                  )}
                </button>
              ))
            )}
          </div>

          {/* Add New Repository */}
          <div className="border-t border-[#1F1F22] p-2">
            <button className="w-full flex items-center gap-2 px-2 py-1.5 text-xs text-slate-400 hover:text-slate-200 hover:bg-[#1A1A1E] rounded transition-all">
              <Plus size={14} />
              <span>Add Repository</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
