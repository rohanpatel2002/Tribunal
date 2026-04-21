'use client';

import { useState, useEffect } from 'react';
import { Shield, Bell, Settings, User, GitPullRequest, Cpu, AlertTriangle, ArrowUpRight, Activity, ChevronDown, MoreVertical, Zap, CheckCircle2, XCircle, RefreshCw } from 'lucide-react';
import {
  fetchAuditSummary, fetchAuditLogs,
  getDemoAuditSummary, getDemoPRAnalysisRecords,
  type AuditSummary, type PRAnalysisRecord,
} from '@/lib/api';
import { BriefingDetail } from '@/components/BriefingDetail';
import OverviewTab from '@/components/tabs/OverviewTab';
import PoliciesTab from '@/components/tabs/PoliciesTab';
import WebhooksTab from '@/components/tabs/WebhooksTab';
import APIKeysTab from '@/components/tabs/APIKeysTab';

const REPOSITORY = 'rohanpatel2002/tribunal';
const TABS = ['Overview', 'Analytics', 'Policies', 'Webhooks', 'API Keys'] as const;
type Tab = typeof TABS[number];

const GRADIENT = { background: 'linear-gradient(135deg, #f8f9f0 0%, #f6ecd8 55%, #f2d896 100%)', fontFamily: "'Geist', 'Inter', Arial, sans-serif" };

function ago(iso?: string) {
  if (!iso) return '—';
  const m = Math.floor((Date.now() - new Date(iso).getTime()) / 60000);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  return h < 24 ? `${h}h ago` : `${Math.floor(h / 24)}d ago`;
}

// ─── Analytics tab content ────────────────────────────────────────────────────

function AnalyticsContent({ apiKey, onDemo }: { apiKey: string; onDemo: (v: boolean) => void }) {
  const [summary, setSummary] = useState<AuditSummary | null>(null);
  const [logs, setLogs] = useState<PRAnalysisRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedPR, setExpandedPR] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    try {
      const [s, l] = await Promise.all([
        fetchAuditSummary(REPOSITORY, apiKey),
        fetchAuditLogs(REPOSITORY, apiKey, { limit: 8 }),
      ]);
      setSummary(s ?? getDemoAuditSummary(REPOSITORY));
      setLogs(l ?? getDemoPRAnalysisRecords(REPOSITORY));
      onDemo(!s || !l);
    } catch {
      setSummary(getDemoAuditSummary(REPOSITORY));
      setLogs(getDemoPRAnalysisRecords(REPOSITORY));
      onDemo(true);
    } finally { setLoading(false); }
  }

  useEffect(() => { load(); }, []); // eslint-disable-line

  const approve = logs.filter(l => l.recommendation === 'APPROVE').length;
  const block   = logs.filter(l => l.recommendation === 'BLOCK').length;
  const review  = logs.filter(l => l.recommendation === 'REVIEW_REQUIRED').length;
  const total   = Math.max(logs.length, 1);
  const aiPct   = summary ? Math.round((summary.aiGeneratedPRs / Math.max(summary.totalPRs, 1)) * 100) : 0;
  const riskTotal = (summary?.criticalRisks ?? 0) + (summary?.highRisks ?? 0);
  const bars = [
    { day: 'M', h: '35%' }, { day: 'T', h: '55%' }, { day: 'W', h: '42%' },
    { day: 'T', h: '70%' }, { day: 'F', h: '100%', today: true },
    { day: 'S', h: '20%' }, { day: 'S', h: '15%' },
  ];

  if (loading) return (
    <div className="flex items-center justify-center py-24">
      <div className="w-8 h-8 rounded-full border-2 border-[#292b2a]/20 border-t-[#292b2a] animate-spin" />
    </div>
  );

  return (
    <div className="space-y-8">
      {/* Welcome row */}
      <section>
        <h1 className="text-[42px] font-light tracking-tight text-slate-900 mb-8">Security Overview</h1>
        <div className="flex items-end justify-between">
          <div className="flex gap-3.5">
            <Pill label="Scanned" value={`${summary?.totalPRs ?? 0} PRs`} dark />
            <Pill label="AI Code" value={`${aiPct}%`} yellow />
            <div className="space-y-2 w-52">
              <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest">Risk Coverage</p>
              <div className="w-full flex items-center h-[38px] px-4 rounded-full bg-white/40 border border-white/60 overflow-hidden relative">
                <div className="absolute left-0 top-0 bottom-0 bg-[#fad961]/50 rounded-full transition-all"
                     style={{ width: `${Math.min(90, 20 + Math.round((riskTotal / Math.max(summary?.totalFiles ?? 1, 1)) * 100))}%` }} />
                <span className="text-[13px] text-slate-600 relative z-10">{Math.min(90, 20 + Math.round((riskTotal / Math.max(summary?.totalFiles ?? 1, 1)) * 100))}%</span>
                {[25, 50, 75].map(p => <div key={p} className="absolute top-1/2 -translate-y-1/2 w-px h-4 bg-slate-300" style={{ left: `${p}%` }} />)}
              </div>
            </div>
            <Pill label="Blocked" value={String(block)} outline />
          </div>
          <div className="flex items-end gap-9 pr-2">
            {[
              { Icon: GitPullRequest, value: summary?.totalPRs ?? 0, label: 'Total PRs' },
              { Icon: Cpu,           value: summary?.aiGeneratedPRs ?? 0, label: 'AI-Gen' },
              { Icon: AlertTriangle, value: riskTotal, label: 'Risk Events' },
            ].map(({ Icon, value, label }) => (
              <div key={label} className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-full bg-[#e8e9dc] flex items-center justify-center -mr-0.5">
                  <Icon size={13} className="text-slate-600" />
                </div>
                <span className="text-[44px] font-light leading-none -tracking-wide text-slate-900">{value}</span>
                <span className="text-[11px] text-slate-500 font-medium leading-tight mt-2">{label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Cards grid */}
      <div className="grid grid-cols-12 gap-5">
        {/* Left col */}
        <div className="col-span-3 flex flex-col gap-5">
          {/* Latest scan */}
          <div className="relative rounded-[32px] overflow-hidden h-[210px] shadow-[0_8px_30px_rgba(0,0,0,0.13)]"
               style={{ background: 'linear-gradient(145deg, #1a1c2c, #2d1b4e)' }}>
            <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-transparent" />
            <div className="absolute bottom-5 left-5 right-5">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-5 h-5 rounded-full bg-emerald-400 flex items-center justify-center"><CheckCircle2 size={11} className="text-white" strokeWidth={3} /></div>
                <span className="text-white/60 text-[11px] uppercase tracking-wide font-medium">Latest Scan</span>
              </div>
              <h3 className="text-white font-medium text-xl">PR #{logs[0]?.prNumber ?? '—'}</h3>
              <div className="flex items-center justify-between mt-1.5">
                <p className="text-white/45 text-xs">{logs[0]?.totalFiles ?? 0} files · {logs[0]?.aiGenerated ?? 0} AI</p>
                <span className={`px-2.5 py-0.5 rounded-full text-[10px] font-semibold border ${
                  logs[0]?.recommendation === 'APPROVE' ? 'bg-emerald-500/20 border-emerald-400/40 text-emerald-300'
                  : logs[0]?.recommendation === 'BLOCK' ? 'bg-red-500/20 border-red-400/40 text-red-300'
                  : 'bg-amber-500/20 border-amber-400/40 text-amber-300'}`}>
                  {logs[0]?.recommendation ?? 'N/A'}
                </span>
              </div>
            </div>
          </div>
          {/* Info accordion */}
          <div className="bg-white/60 backdrop-blur-md rounded-[32px] p-5 shadow-sm flex-1 flex flex-col justify-between">
            {[{ label: 'Security Policies' }].map(({ label }) => (
              <div key={label} className="flex justify-between items-center pb-3 border-b border-slate-200/60">
                <span className="font-medium text-sm text-slate-800">{label}</span>
                <ChevronDown size={14} className="text-slate-400" />
              </div>
            ))}
            <div className="py-2 border-b border-slate-200/60 space-y-2">
              <div className="flex justify-between items-center text-slate-800">
                <span className="font-medium text-sm">Detection Signals</span>
                <ChevronDown size={14} className="rotate-180" />
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-11 h-9 bg-gradient-to-br from-indigo-500 to-violet-600 rounded-[10px] flex items-center justify-center"><Zap size={16} className="text-white" /></div>
                  <div>
                    <p className="text-[13px] font-medium text-slate-800">3-Signal Engine</p>
                    <p className="text-[11px] text-slate-500">Style · Pattern · Risk</p>
                  </div>
                </div>
                <MoreVertical size={15} className="text-slate-400" />
              </div>
            </div>
            <div className="flex justify-between items-center py-2.5 border-b border-slate-200/60">
              <span className="font-medium text-sm text-slate-800">Webhook Config</span>
              <ChevronDown size={14} className="text-slate-400" />
            </div>
            <div className="flex justify-between items-center pt-2">
              <span className="font-medium text-sm text-slate-800">API Key Management</span>
              <ChevronDown size={14} className="text-slate-400" />
            </div>
          </div>
        </div>

        {/* Center col */}
        <div className="col-span-6 flex flex-col gap-5">
          <div className="flex gap-5 h-[210px]">
            {/* Bar chart */}
            <div className="flex-[0.55] bg-white/70 backdrop-blur-md rounded-[32px] p-6 shadow-sm relative group">
              <div className="absolute top-5 right-5 p-1.5 rounded-full border border-slate-200 group-hover:bg-slate-50 cursor-pointer"><ArrowUpRight size={14} className="text-slate-500" /></div>
              <h3 className="text-[17px] font-medium mb-1 text-slate-800">Risk Activity</h3>
              <div className="flex items-end gap-2">
                <span className="text-[28px] font-light leading-none text-slate-900">{riskTotal}</span>
                <span className="text-[11px] text-slate-500 mb-0.5 font-medium w-16 leading-tight">Risk events this week</span>
              </div>
              <div className="flex items-end justify-between h-[78px] mt-4 px-1 relative">
                <div className="absolute inset-x-0 top-1/2 border-t-[1.5px] border-dashed border-slate-200 -z-0 -translate-y-1/2" />
                {bars.map(({ day, h, today }, i) => (
                  <div key={i} className="flex flex-col items-center h-full justify-end relative">
                    <div className="absolute top-1/2 -mt-[3px] w-1.5 h-1.5 rounded-full bg-slate-200 z-0" />
                    <div className={`w-[3px] rounded-full z-10 ${today ? 'bg-[#fad961]' : (i === 0 || i === 6) ? 'bg-slate-300/50' : 'bg-[#292b2a]'}`} style={{ height: h, marginBottom: '6px' }} />
                    <span className={`text-[10px] font-medium ${today ? 'text-slate-800' : 'text-slate-400'}`}>{day}</span>
                    {today && <div className="absolute -top-6 whitespace-nowrap bg-[#fad961] text-[10px] font-semibold px-2 py-0.5 rounded-md text-slate-900 shadow-sm z-20">{summary?.criticalRisks ?? 0}C · {summary?.highRisks ?? 0}H</div>}
                  </div>
                ))}
              </div>
            </div>
            {/* AI ring */}
            <div className="flex-[0.45] bg-white/70 backdrop-blur-md rounded-[32px] p-6 shadow-sm relative group flex flex-col items-center">
              <div className="absolute top-5 right-5 p-1.5 rounded-full border border-slate-200 group-hover:bg-slate-50"><ArrowUpRight size={14} className="text-slate-500" /></div>
              <h3 className="text-[17px] font-medium self-start w-full text-slate-800">AI Score</h3>
              <div className="relative mt-2 mb-1 w-[88px] h-[88px] flex items-center justify-center">
                <svg className="absolute inset-0 w-full h-full -rotate-90" viewBox="0 0 88 88">
                  <circle cx="44" cy="44" r="38" fill="none" strokeWidth="6" className="stroke-slate-100" />
                  <circle cx="44" cy="44" r="38" fill="none" strokeWidth="6" strokeLinecap="round" stroke="#fad961" strokeDasharray={`${(aiPct / 100) * 2 * Math.PI * 38} ${2 * Math.PI * 38}`} />
                </svg>
                <div className="absolute inset-[-8px] border border-dashed border-slate-300 rounded-full opacity-50" />
                <div className="text-center z-10 mt-1">
                  <span className="text-[24px] font-light text-slate-900 tracking-tight leading-none">{aiPct}%</span>
                  <p className="text-[9px] text-slate-500 font-medium mt-0.5">AI-Gen Code</p>
                </div>
              </div>
              <div className="flex items-center gap-1.5 mt-auto self-start pl-1">
                <Activity size={12} className="text-slate-400" />
                <span className="text-[11px] text-slate-500">{summary?.aiGeneratedPRs ?? 0} PRs flagged</span>
              </div>
            </div>
          </div>

          {/* Audit log */}
          <div className="bg-white/70 backdrop-blur-md rounded-[32px] p-6 pt-5 shadow-sm flex-1 flex flex-col relative overflow-hidden">
            <div className="pointer-events-none absolute -bottom-10 -right-10 w-40 h-40 bg-[#f4d89a]/30 blur-2xl rounded-full" />
            <div className="flex justify-between items-center mb-4 z-10">
              <h3 className="font-medium text-slate-800 text-sm">Recent Audit Log</h3>
              <span className="px-4 py-1 rounded-full border border-slate-200 text-[11px] font-medium bg-white text-slate-600 shadow-sm">{logs.length} scans</span>
            </div>
            <div className="grid grid-cols-[1.2fr_1fr_0.6fr_0.6fr_1fr] text-[10px] text-slate-400 uppercase tracking-widest font-semibold mb-2 pb-2 border-b border-slate-200/60 z-10">
              <div>Pull Request</div><div className="text-center">Verdict</div><div className="text-center">Files</div><div className="text-center">AI</div><div className="text-right">When</div>
            </div>
            <div className="flex-1 overflow-y-auto z-10 space-y-0.5">
              {logs.map(log => {
                const isApprove = log.recommendation === 'APPROVE';
                const isBlock   = log.recommendation === 'BLOCK';
                const cfg = isApprove ? { color: 'text-emerald-700', bg: 'bg-emerald-50', border: 'border-emerald-200', label: 'Approved' }
                  : isBlock ? { color: 'text-red-700', bg: 'bg-red-50', border: 'border-red-200', label: 'Blocked' }
                  : { color: 'text-amber-700', bg: 'bg-amber-50', border: 'border-amber-200', label: 'Review' };
                const isExpanded = expandedPR === log.id;
                return (
                  <div key={log.id}>
                    <div onClick={() => log.contextBriefing && setExpandedPR(isExpanded ? null : log.id)}
                      className={`grid grid-cols-[1.2fr_1fr_0.6fr_0.6fr_1fr] items-center py-2.5 px-3 rounded-xl transition-colors ${log.contextBriefing ? 'cursor-pointer' : ''} hover:bg-slate-50/80`}>
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-lg bg-slate-100 flex items-center justify-center shrink-0"><GitPullRequest size={11} className="text-slate-600" /></div>
                        <span className="text-[13px] font-medium text-slate-800">#{log.prNumber}</span>
                      </div>
                      <div className="flex justify-center">
                        <span className={`text-[10px] font-semibold px-2.5 py-0.5 rounded-full border ${cfg.bg} ${cfg.color} ${cfg.border}`}>{cfg.label}</span>
                      </div>
                      <div className="text-center text-[12px] text-slate-500">{log.totalFiles}</div>
                      <div className="text-center text-[12px] text-slate-500">{log.aiGenerated}</div>
                      <div className="text-right"><span className="text-[11px] text-slate-400">{ago(log.createdAt)}</span></div>
                    </div>
                    {isExpanded && log.contextBriefing && (
                      <div className="mx-3 mb-2 rounded-2xl border border-slate-200/60 bg-white/60 p-4 text-sm">
                        <BriefingDetail briefing={log.contextBriefing} prNumber={log.prNumber} recommendation={log.recommendation} createdAt={log.createdAt} />
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right col */}
        <div className="col-span-3 flex flex-col gap-5">
          {/* Verdicts */}
          <div className="bg-white/60 backdrop-blur-md rounded-[32px] p-6 shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <h3 className="text-[17px] font-medium text-slate-800">Verdicts</h3>
              <span className="text-3xl font-light text-slate-900">{Math.round((approve / total) * 100)}%</span>
            </div>
            <div className="flex gap-1 h-7 w-full rounded-full overflow-hidden">
              <div className="bg-emerald-400 flex items-center justify-center text-[10px] font-semibold text-white" style={{ flex: Math.max(approve, 0.1) }}>{approve > 0 ? 'OK' : ''}</div>
              <div className="bg-[#fad961]" style={{ flex: Math.max(review, 0.1) }} />
              <div className="bg-red-400" style={{ flex: Math.max(block, 0.1) }} />
            </div>
            <div className="flex justify-between mt-3 text-[10px] text-slate-500">
              <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-emerald-400 inline-block" /> Approved</span>
              <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-[#fad961] inline-block" /> Review</span>
              <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-red-400 inline-block" /> Blocked</span>
            </div>
          </div>
          {/* Recent scans */}
          <div className="bg-[#292b2a] rounded-[32px] p-7 pt-6 text-white flex-1 shadow-[0_20px_40px_rgba(0,0,0,0.12)] flex flex-col">
            <div className="flex justify-between items-start mb-5">
              <h3 className="text-[17px] font-medium text-white/90">Recent Scans</h3>
              <span className="text-3xl font-light text-white/90">{logs.length}/{summary?.totalPRs ?? 0}</span>
            </div>
            <div className="space-y-4">
              {logs.slice(0, 5).map((log, i) => {
                const isApprove = log.recommendation === 'APPROVE';
                const isBlock   = log.recommendation === 'BLOCK';
                const Icon = isApprove ? CheckCircle2 : isBlock ? XCircle : AlertTriangle;
                const iconColor = isApprove ? 'text-emerald-400' : isBlock ? 'text-red-400' : 'text-amber-400';
                return (
                  <div key={log.id} className="flex items-center gap-3.5 group cursor-pointer">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${i < 2 ? 'bg-white/10' : 'bg-white/5 group-hover:bg-white/10'} transition-colors`}>
                      <Icon size={14} className={iconColor} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className={`text-[13px] font-medium leading-tight ${i < 2 ? 'text-white/90' : 'text-white/45'}`}>PR #{log.prNumber}</p>
                      <p className="text-[10px] text-white/30 mt-0.5">{log.totalFiles} files · {log.aiGenerated} AI-gen</p>
                    </div>
                    <div className="shrink-0">
                      {isApprove ? (
                        <div className="w-5 h-5 rounded-full bg-emerald-400 flex items-center justify-center shadow-[0_0_8px_rgba(52,211,153,0.4)]"><CheckCircle2 size={12} className="text-black" strokeWidth={3} /></div>
                      ) : (
                        <div className={`w-5 h-5 rounded-full border-[1.5px] ${isBlock ? 'border-red-400' : 'border-white/20 group-hover:border-white/40'} transition-colors`} />
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Pill helpers ─────────────────────────────────────────────────────────────

function Pill({ label, value, dark, yellow, outline }: { label: string; value: string; dark?: boolean; yellow?: boolean; outline?: boolean }) {
  return (
    <div className="space-y-2">
      <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest">{label}</p>
      <div className={`px-7 py-2.5 rounded-full text-[13px] font-medium shadow-md text-center
        ${dark ? 'bg-[#292b2a] text-white' : yellow ? 'bg-[#fad961] text-slate-900 font-semibold' : outline ? 'border-[1.5px] border-slate-400/60 text-slate-700' : 'bg-slate-100 text-slate-700'}`}>
        {value}
      </div>
    </div>
  );
}

// ─── Root page ────────────────────────────────────────────────────────────────

export default function Dashboard() {
  const [tab, setTab]       = useState<Tab>('Analytics');
  const [demo, setDemo]     = useState(false);
  const apiKey = process.env.NEXT_PUBLIC_API_KEY ?? 'dev_enterprise_key_123';

  return (
    <div className="h-full w-full overflow-auto p-8 flex flex-col gap-6 relative" style={GRADIENT}>
      {/* noise */}
      <div className="pointer-events-none absolute inset-0 opacity-[0.018]"
           style={{ backgroundImage: "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.75' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")" }} />

      {/* ── Header ── */}
      <header className="flex items-center justify-between text-sm shrink-0">
        <div className="flex items-center gap-2.5 px-5 py-2.5 border border-slate-300/70 rounded-full bg-white/55 backdrop-blur-sm shadow-sm">
          <Shield size={15} className="text-slate-700" strokeWidth={2.5} />
          <span className="text-[17px] font-light tracking-tight text-slate-900">TRIBUNAL</span>
        </div>

        <nav className="flex items-center gap-0.5 bg-white/55 backdrop-blur-md rounded-full p-1 shadow-sm">
          {TABS.map(t => (
            <button key={t} onClick={() => setTab(t)}
              className={`px-4 py-2 rounded-full text-[13px] font-medium transition-colors ${t === tab ? 'bg-[#292b2a] text-white shadow-md' : 'text-slate-600 hover:bg-white/60'}`}>
              {t}
            </button>
          ))}
        </nav>

        <div className="flex items-center gap-2.5">
          {demo && <span className="px-3 py-1 bg-[#fad961]/60 text-slate-700 rounded-full text-[11px] font-semibold uppercase tracking-wide">Demo</span>}
          <button onClick={() => setTab('Analytics')} className="flex items-center gap-1.5 px-4 py-2.5 bg-white/65 backdrop-blur-sm rounded-full shadow-sm font-medium hover:bg-white transition-colors text-slate-700 text-[13px]">
            <RefreshCw size={14} /> Refresh
          </button>
          <button onClick={() => setTab('Overview')} className="flex items-center gap-1.5 px-4 py-2.5 bg-white/65 backdrop-blur-sm rounded-full shadow-sm font-medium hover:bg-white transition-colors text-slate-700 text-[13px]">
            <Settings size={14} /> Settings
          </button>
          <button className="p-2.5 bg-white/65 backdrop-blur-sm rounded-full shadow-sm hover:bg-white text-slate-700 transition-colors">
            <Bell size={16} />
          </button>
          <button className="p-2.5 bg-white/65 backdrop-blur-sm rounded-full shadow-sm hover:bg-white text-slate-700 transition-colors">
            <User size={16} />
          </button>
        </div>
      </header>

      {/* ── Tab content ── */}
      <div className="flex-1 pb-4">
        {tab === 'Overview'  && <OverviewTab  apiKey={apiKey} />}
        {tab === 'Analytics' && <AnalyticsContent apiKey={apiKey} onDemo={setDemo} />}
        {tab === 'Policies'  && <PoliciesTab  apiKey={apiKey} repository={REPOSITORY} />}
        {tab === 'Webhooks'  && <WebhooksTab  apiKey={apiKey} />}
        {tab === 'API Keys'  && <APIKeysTab   apiKey={apiKey} repository={REPOSITORY} />}
      </div>
    </div>
  );
}
