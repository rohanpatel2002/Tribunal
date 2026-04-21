'use client';

import React from 'react';
import { AlertCircle, ChevronDown, ChevronUp, FileText, Clock, Tag } from 'lucide-react';
import { useState } from 'react';

interface BriefingDetailProps {
  briefing: string;
  prNumber: number;
  recommendation: string;
  createdAt?: string;
}

export function BriefingDetail({ briefing, prNumber, recommendation, createdAt }: BriefingDetailProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  if (!briefing) {
    return null;
  }

  // Render markdown-like briefing as HTML
  const renderBriefing = (text: string) => {
    const lines = text.split('\n');
    const elements: React.ReactNode[] = [];
    let i = 0;

    while (i < lines.length) {
      const line = lines[i];

      // Headers
      if (line.startsWith('# ')) {
        elements.push(
          <h1 key={i} className="text-2xl font-bold text-white mt-4 mb-2">
            {line.replace('# ', '')}
          </h1>
        );
      } else if (line.startsWith('## ')) {
        elements.push(
          <h2 key={i} className="text-xl font-bold text-emerald-400 mt-3 mb-2">
            {line.replace('## ', '')}
          </h2>
        );
      } else if (line.startsWith('### ')) {
        elements.push(
          <h3 key={i} className="text-lg font-semibold text-zinc-200 mt-2 mb-1">
            {line.replace('### ', '')}
          </h3>
        );
      }
      // Code blocks
      else if (line.startsWith('```')) {
        const codeLines = [];
        i++;
        while (i < lines.length && !lines[i].startsWith('```')) {
          codeLines.push(lines[i]);
          i++;
        }
        elements.push(
          <pre key={i} className="bg-zinc-900 border border-zinc-800 rounded p-3 text-sm overflow-x-auto my-2">
            <code className="text-zinc-200 font-mono">{codeLines.join('\n')}</code>
          </pre>
        );
      }
      // Bold text
      else if (line.includes('**')) {
        const parts = line.split(/\*\*([^*]+)\*\*/);
        elements.push(
          <p key={i} className="text-zinc-300 text-sm mb-2">
            {parts.map((part, idx) => (
              <span key={idx} className={idx % 2 === 1 ? 'font-bold text-white' : ''}>
                {part}
              </span>
            ))}
          </p>
        );
      }
      // Lists
      else if (line.startsWith('- ')) {
        elements.push(
          <li key={i} className="text-zinc-300 text-sm ml-4 mb-1">
            {line.replace('- ', '')}
          </li>
        );
      }
      // Horizontal rule
      else if (line.trim() === '---') {
        elements.push(
          <hr key={i} className="border-t border-zinc-700 my-3" />
        );
      }
      // Regular paragraphs
      else if (line.trim()) {
        elements.push(
          <p key={i} className="text-zinc-300 text-sm mb-2">
            {line}
          </p>
        );
      } else {
        elements.push(<div key={i} className="h-1" />);
      }

      i++;
    }

    return elements;
  };

  return (
    <div className="bg-zinc-900/50 border border-zinc-800/50 rounded-lg p-4 my-4">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between text-left hover:bg-zinc-800/30 p-2 rounded transition"
      >
        <div className="flex items-center gap-3">
          <FileText className="h-5 w-5 text-emerald-500 shrink-0" />
          <div>
            <h3 className="font-semibold text-white">
              {recommendation === 'BLOCK' ? '🔴 PR #' + prNumber : '📋 PR #' + prNumber} - Context Briefing
            </h3>
            {createdAt && (
              <p className="text-xs text-zinc-500 mt-1 flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {new Date(createdAt).toLocaleDateString()}
              </p>
            )}
          </div>
        </div>
        {isExpanded ? (
          <ChevronUp className="h-5 w-5 text-zinc-500 shrink-0" />
        ) : (
          <ChevronDown className="h-5 w-5 text-zinc-500 shrink-0" />
        )}
      </button>

      {isExpanded && (
        <div className="mt-4 pl-8 pr-2 max-h-96 overflow-y-auto text-zinc-300 space-y-2 pb-2">
          {renderBriefing(briefing)}
        </div>
      )}
    </div>
  );
}
