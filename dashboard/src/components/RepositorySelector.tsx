'use client';

import React, { useMemo, useState } from 'react';
import { GitPullRequest, ChevronDown, Check, Plus } from 'lucide-react';
import { clsx } from 'clsx';

interface Repository {
  id: string;
  name: string;
  fullName: string;
  url: string;
  isMonitored: boolean;
  analysisCount?: number;
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
  const normalizeRepo = (value: string) => value.trim().toLowerCase();

  const filteredRepos = useMemo(
    () =>
      repositories.filter(
        (repo) =>
          repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          repo.fullName.toLowerCase().includes(searchQuery.toLowerCase())
      ),
    [searchQuery, repositories]
  );

  const currentRepo = repositories.find(
    (r) => normalizeRepo(r.fullName) === normalizeRepo(selectedRepo)
  );

  return (
    <div className="relative w-full max-w-sm">
      {/* Selector Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between px-3 py-2 bg-white border border-slate-200 rounded-lg text-sm text-slate-700 hover:border-slate-300 transition-all group"
      >
        <div className="flex items-center gap-2 min-w-0">
          <GitPullRequest size={16} className="text-slate-400 shrink-0" />
          <span className="truncate">
            {currentRepo?.name || 'Select Repository'}
          </span>
        </div>
        <ChevronDown
          size={16}
          className={clsx(
            'text-slate-400 shrink-0 transition-transform',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-slate-200 rounded-lg shadow-xl z-50 overflow-hidden">
          {/* Search Input */}
          <div className="p-2 border-b border-slate-200">
              <input
              type="text"
              placeholder="Search repositories..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-2 py-1.5 bg-slate-50 border border-slate-200 rounded text-xs text-slate-600 placeholder-slate-400 focus:outline-none focus:border-green-700/50"
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
                    normalizeRepo(selectedRepo) === normalizeRepo(repo.fullName)
                      ? 'bg-green-700/10 text-green-700'
                      : 'text-slate-600 hover:bg-slate-50'
                  )}
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <GitPullRequest size={14} className="text-slate-400 shrink-0" />
                    <div className="min-w-0">
                      <p className="font-medium truncate">{repo.name}</p>
                      <p className="text-[10px] text-slate-400 truncate">{repo.fullName}</p>
                      <p className="text-[10px] text-slate-400">
                        {repo.analysisCount && repo.analysisCount > 0
                          ? `${repo.analysisCount} analyses`
                          : 'No analyses yet'}
                      </p>
                    </div>
                  </div>
                  {normalizeRepo(selectedRepo) === normalizeRepo(repo.fullName) && (
                    <Check size={14} className="text-green-700 shrink-0 ml-2" />
                  )}
                </button>
              ))
            )}
          </div>

          {/* Add New Repository */}
          <div className="border-t border-slate-200 p-2">
            <button className="w-full flex items-center gap-2 px-2 py-1.5 text-xs text-slate-500 hover:text-slate-700 hover:bg-slate-100 rounded transition-all">
              <Plus size={14} />
              <span>Add Repository</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
