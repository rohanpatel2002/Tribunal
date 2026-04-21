'use client';
import { useState, useEffect } from 'react';
import { CheckCircle2, XCircle, RefreshCw } from 'lucide-react';
import {
  checkHealthStatus, fetchGitHubConnectionStatus, fetchRepositoryAnalysisCounts,
  startGitHubConnection, disconnectGitHub,
  type GitHubConnectionStatus, type RepositoryAnalysisCount,
} from '@/lib/api';

const CARD = 'bg-white/60 backdrop-blur-md rounded-[28px] p-6 shadow-sm';

function ago(iso?: string) {
  if (!iso) return '—';
  const m = Math.floor((Date.now() - new Date(iso).getTime()) / 60000);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  return h < 24 ? `${h}h ago` : `${Math.floor(h / 24)}d ago`;
}

export default function OverviewTab({ apiKey }: { apiKey: string }) {
  const [health, setHealth] = useState<boolean | null>(null);
  const [gh, setGh] = useState<GitHubConnectionStatus | null>(null);
  const [repos, setRepos] = useState<RepositoryAnalysisCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  async function load() {
    setLoading(true);
    const [h, g, r] = await Promise.all([
      checkHealthStatus(apiKey),
      fetchGitHubConnectionStatus(apiKey),
      fetchRepositoryAnalysisCounts(apiKey),
    ]);
    setHealth(h); setGh(g); setRepos(r ?? []); setLoading(false);
  }

  useEffect(() => { load(); }, [apiKey]); // eslint-disable-line

  async function connect() {
    setBusy(true);
    const url = await startGitHubConnection(apiKey);
    setBusy(false);
    if (url) window.open(url, '_blank');
  }

  async function disconnect() {
    setBusy(true);
    await disconnectGitHub(apiKey);
    setGh(null); setBusy(false);
  }

  if (loading) return (
    <div className="flex items-center justify-center py-24">
      <div className="w-8 h-8 rounded-full border-2 border-[#292b2a]/20 border-t-[#292b2a] animate-spin" />
    </div>
  );

  const rows = [
    { label: 'Go Interceptor', ok: health !== false },
    { label: 'Backend API', ok: health !== false },
    { label: 'Enterprise Key', ok: !!apiKey },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-[38px] font-light tracking-tight text-slate-900">Overview</h1>
        <p className="text-sm text-slate-500 mt-1">System status and integration health</p>
      </div>

      <div className="grid grid-cols-12 gap-5">
        {/* GitHub */}
        <div className={`col-span-4 ${CARD}`}>
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-800">GitHub Connection</h2>
            <button onClick={load} className="text-slate-400 hover:text-slate-700"><RefreshCw size={13} /></button>
          </div>
          {gh?.connected ? (
            <>
              <div className="flex items-center gap-3 mb-4">
                {gh.avatarUrl && <img src={gh.avatarUrl} alt="" className="w-10 h-10 rounded-full border border-white shadow-sm" />}
                <div>
                  <p className="font-medium text-slate-800 text-sm">{gh.name || gh.login}</p>
                  <p className="text-[11px] text-slate-500">@{gh.login} · {ago(gh.connectedAt)}</p>
                </div>
              </div>
              {gh.repos?.length > 0 && (
                <div className="mb-4 space-y-1 max-h-32 overflow-y-auto">
                  {gh.repos.slice(0, 5).map(r => (
                    <a key={r.id} href={r.htmlUrl} target="_blank" rel="noopener noreferrer"
                      className="flex items-center justify-between py-1 text-xs text-slate-600 hover:text-slate-900 transition-colors">
                      <span className="font-mono">{r.fullName}</span>
                      {r.private && <span className="text-[10px] bg-slate-100 px-1.5 rounded text-slate-500">private</span>}
                    </a>
                  ))}
                </div>
              )}
              <button onClick={disconnect} disabled={busy}
                className="w-full py-2 rounded-full border border-slate-300 text-sm text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50">
                {busy ? 'Disconnecting…' : 'Disconnect'}
              </button>
            </>
          ) : (
            <>
              <p className="text-sm text-slate-500 mb-5">Connect GitHub to enable Check Runs and repository context fetching.</p>
              <button onClick={connect} disabled={busy}
                className="w-full py-2.5 bg-[#292b2a] text-white rounded-full text-sm font-medium hover:bg-black transition-colors disabled:opacity-50">
                {busy ? 'Opening…' : 'Connect GitHub'}
              </button>
            </>
          )}
        </div>

        {/* Health */}
        <div className={`col-span-4 ${CARD}`}>
          <h2 className="font-semibold text-slate-800 mb-4">System Health</h2>
          <div className="space-y-4">
            {rows.map(({ label, ok }) => (
              <div key={label} className="flex items-center justify-between">
                <span className="text-sm text-slate-600">{label}</span>
                <div className={`flex items-center gap-1.5 text-xs font-medium ${ok ? 'text-emerald-600' : 'text-red-500'}`}>
                  {ok ? <CheckCircle2 size={14} /> : <XCircle size={14} />}
                  {ok ? 'Healthy' : 'Error'}
                </div>
              </div>
            ))}
            <div className="pt-2 border-t border-slate-100">
              <p className="text-[11px] text-slate-400">Backend: {process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}</p>
            </div>
          </div>
        </div>

        {/* Repos */}
        <div className={`col-span-4 ${CARD}`}>
          <h2 className="font-semibold text-slate-800 mb-4">Tracked Repositories</h2>
          {repos.length === 0 ? (
            <p className="text-sm text-slate-400">No repositories with analyses yet. Submit a webhook to get started.</p>
          ) : (
            <div className="space-y-2.5">
              {repos.map(r => (
                <div key={r.repository} className="flex items-center justify-between py-1.5 border-b border-slate-100 last:border-0">
                  <span className="text-sm font-mono text-slate-700">{r.repository}</span>
                  <span className="text-xs font-medium text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full">{r.totalPRs} PRs</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
