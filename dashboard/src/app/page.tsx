'use client';

import { useState, useEffect, useMemo } from 'react';
import { Settings, GitPullRequest, Activity, ChevronRight, Fingerprint, RefreshCcw, Lock, ChevronDown, TrendingUp, TrendingDown, AlertCircle, Zap, Shield } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Legend, PieChart, Pie, Cell, LineChart, Line } from 'recharts';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface AuditSummary {
  repository: string;
  totalPRs: number;
  totalFiles: number;
  aiGeneratedPRs: number;
  criticalRisks: number;
  highRisks: number;
  averageAIScore: number;
}

interface PRAnalysisRecord {
  id: string;
  repository: string;
  prNumber: number;
  recommendation: string;
  totalFiles: number;
  aiGenerated: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
}

const TABS = [
  { id: 'overview', label: 'Overview', icon: Activity },
  { id: 'risks', label: 'Risk Analysis', icon: AlertCircle },
  { id: 'repos', label: 'Repositories', icon: GitPullRequest },
  { id: 'settings', label: 'Settings', icon: Lock },
];

export default function Dashboard() {
  const [data, setData] = useState<AuditSummary | null>(null);
  const [logs, setLogs] = useState<PRAnalysisRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [repo, setRepo] = useState("rohanpatel2002/tribunal");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');

  // Get API base URL from environment or default
  const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const fetchData = async () => {
    setLoading(true);
    setIsRefreshing(true);
    try {
      const res = await fetch(`${apiBase}/api/v1/audit/summary?repository=${encodeURIComponent(repo)}`, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
      });
      if (res.ok) {
        setData(await res.json());
      } else {
        throw new Error("Analytics summary failed");
      }
    } catch (e) {
      console.error("Failed to fetch summary:", e);
      setData(null);
    }

    try {
      const logsRes = await fetch(`${apiBase}/api/v1/audit/logs?repository=${encodeURIComponent(repo)}&limit=10`, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
      });
      if (logsRes.ok) {
        const json = await logsRes.json();
        setLogs(json.data || []);
      } else {
        throw new Error("Logs failed");
      }
    } catch (e) {
      console.error("Failed to fetch logs:", e);
      setLogs([]);
    } finally {
      setLoading(false);
      setTimeout(() => setIsRefreshing(false), 500);
    }
  };

  useEffect(() => {
    fetchData();
  }, [repo]);

  const chartData = useMemo(() => {
    return logs.slice().reverse().map(log => ({
      name: `PR #${log.prNumber}`,
      Files: log.totalFiles,
      AI: log.aiGenerated,
      Risks: log.critical + log.high
    }));
  }, [logs]);

  return (
    <div className="flex w-full h-full text-slate-200 font-sans overflow-hidden bg-[#0A0A0A]">
      <aside className="w-64 bg-[#0F0F11] border-r border-[#1F1F22] flex flex-col justify-between pt-6 pb-4">
        <div className="px-5 flex flex-col gap-8">
          <div className="flex items-center gap-2">
             <div>
               <h1 className="text-5xl font-black tracking-tighter bg-gradient-to-r from-indigo-400 via-purple-400 to-indigo-400 bg-clip-text text-transparent drop-shadow-lg leading-tight">Tribunal</h1>
             </div>
          </div>

          <nav className="flex flex-col gap-1 mt-2">
             <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-2 px-3">Dashboard</p>
             {TABS.map((tab) => (
                <button 
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={cn(
                    "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all",
                    activeTab === tab.id 
                      ? "bg-indigo-500/10 text-indigo-400 font-medium border border-indigo-500/20" 
                      : "text-slate-400 hover:text-slate-200 hover:bg-[#1A1A1E]"
                  )}
                >
                  <tab.icon size={16} />
                  <span>{tab.label}</span>
                </button>
             ))}
          </nav>
        </div>

        <div className="px-5 mt-auto flex flex-col gap-4">
          <button className="flex items-center gap-3 px-3 py-2 text-slate-400 hover:text-slate-200 hover:bg-[#1A1A1E] rounded-lg transition-colors text-sm">
             <Settings size={16} />
             <span>Settings</span>
          </button>
          
          <div className="pt-4 border-t border-[#1F1F22] flex items-center justify-between px-3">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-[#1A1A1E] border border-[#27272A] flex items-center justify-center">
                 <div className="w-4 h-4 rounded-full bg-linear-to-tr from-cyan-400 to-indigo-500" />
              </div>
              <div className="flex flex-col">
                <span className="text-xs font-semibold text-white">Rohan P.</span>
                <span className="text-[10px] text-emerald-500">System Admin</span>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <main className="flex-1 flex flex-col h-full bg-transparent overflow-y-auto relative">
         <header className="h-14 border-b border-[#1F1F22] bg-[#0A0A0A]/80 backdrop-blur-xl sticky top-0 z-10 flex items-center justify-between px-8">
            <div className="flex items-center gap-3 text-sm">
               <span className="text-slate-500">Target</span>
               <ChevronRight size={14} className="text-slate-700" />
               <div className="bg-[#1A1A1E] px-2.5 py-1 flex items-center gap-2 rounded-md border border-[#27272A]">
                  <GitPullRequest size={14} className="text-slate-400" />
                  <span className="font-mono text-slate-300">{repo}</span>
                  <ChevronDown size={14} className="text-slate-500 ml-2" />
               </div>
            </div>
            <div className="flex items-center gap-3">
               <button onClick={fetchData} className="flex items-center justify-center p-1.5 text-slate-400 hover:bg-[#1A1A1E] hover:text-white rounded border border-transparent hover:border-[#27272A] transition-all">
                  <RefreshCcw size={16} className={isRefreshing ? "animate-spin text-indigo-400" : ""} />
               </button>
            </div>
         </header>

         <div className="p-8 max-w-350 w-full mx-auto pb-24">
            {loading ? (
              <div className="grid grid-cols-4 gap-4 animate-pulse">
                {[1,2,3,4].map(i => <div key={i} className="h-28 bg-[#151518] rounded-xl border border-[#1F1F22]" />)}
              </div>
            ) : (
              <>
                 {activeTab === 'overview' && <RiskCommandView data={data} logs={logs} chartData={chartData} />}
                 {activeTab === 'risks' && <VulnerabilitiesView data={data} logs={logs} />}
                 {activeTab === 'repos' && <RepositoriesView data={data} repo={repo} />}
                 {activeTab === 'settings' && <PoliciesView />}
              </>
            )}
         </div>
      </main>
    </div>
  );
}

function RiskCommandView({ data, logs, chartData }: any) {
  return (
    <div className="animate-in fade-in duration-500">
      <div className="mb-8">
        <h2 className="text-2xl font-semibold tracking-tight text-white">Risk Command Center</h2>
        <p className="text-sm text-slate-400 mt-1">Real-time analysis of pull requests, adversarial AI detections, and access violations.</p>
      </div>

      <div className="flex flex-col gap-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricCard title="Analyzed Pull Requests" value={data?.totalPRs || 0} icon={GitPullRequest} color="text-blue-400" bg="bg-blue-400/10" />
          <MetricCard title="AI-Generated PRs" value={data?.aiGeneratedPRs || 0} icon={Zap} color="text-indigo-400" bg="bg-indigo-400/10" />
          <MetricCard title="Critical Policy Risks" value={data?.criticalRisks || 0} icon={AlertCircle} color={data?.criticalRisks ? "text-red-400" : "text-slate-400"} bg={data?.criticalRisks ? "bg-red-400/10" : "bg-[#1A1A1E]"} />
          <MetricCard title="Global Syntax Trust" value={`${(100 - ((data?.averageAIScore || 0)*100)).toFixed(1)}%`} icon={Shield} color="text-emerald-400" bg="bg-emerald-400/10" />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 bg-[#0F0F11] border border-[#1F1F22] rounded-xl p-6">
            <h3 className="text-sm font-semibold text-slate-300 mb-6">Pipeline Activity (Recent PRs)</h3>
             <div className="h-64 w-full">
              <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={0}>
                <BarChart data={chartData} margin={{ top: 0, right: 0, left: -20, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272A" vertical={false} />
                  <XAxis dataKey="name" stroke="#52525B" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis stroke="#52525B" fontSize={12} tickLine={false} axisLine={false} />
                  <Tooltip cursor={{ fill: '#1A1A1E' }} contentStyle={{ backgroundColor: '#0F0F11', border: '1px solid #27272A', borderRadius: '8px' }} />
                  <Legend wrapperStyle={{ fontSize: '12px' }} />
                  <Bar dataKey="Files" fill="#3B82F6" radius={[2, 2, 0, 0]} />
                  <Bar dataKey="AI" fill="#8B5CF6" radius={[2, 2, 0, 0]} />
                  <Bar dataKey="Risks" fill="#EF4444" radius={[2, 2, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl p-6 flex flex-col items-center justify-center relative overflow-hidden">
             <div className="absolute inset-0 bg-linear-to-b from-indigo-500/5 to-transparent pointer-events-none" />
             <span className="text-xs uppercase tracking-widest font-semibold text-indigo-400 mb-6 z-10">Threat Posture</span>
             
             <div className="relative w-40 h-40 flex items-center justify-center mb-6 z-10">
                <svg className="w-full h-full -rotate-90 transform" viewBox="0 0 100 100">
                   <circle cx="50" cy="50" r="40" stroke="#1F1F22" strokeWidth="8" fill="none" />
                   <circle cx="50" cy="50" r="40" stroke="currentColor" className="text-indigo-500" strokeWidth="8" fill="none" strokeDasharray="251.2" strokeDashoffset={251.2 - (251.2 * ((data?.averageAIScore || 0) * 100)) / 100} strokeLinecap="round" />
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                   <span className="text-3xl font-bold text-white">{((data?.averageAIScore || 0)*100).toFixed(0)}</span>
                   <span className="text-[10px] text-slate-500 font-mono">AI Index</span>
                </div>
             </div>
             <div className="w-full bg-[#1A1A1E] border border-[#27272A] rounded-lg p-3 z-10">
                <div className="flex justify-between items-center text-sm">
                  <span className="text-slate-400">Total Scans</span>
                  <span className="font-mono text-slate-300">{data?.totalFiles}</span>
                </div>
             </div>
          </div>
        </div>

        <LogTable logs={logs} title="Latest Audit Trail" />
      </div>
    </div>
  );
}

function VulnerabilitiesView({ data, logs }: any) {
  const riskyLogs = logs.filter((l: any) => l.critical > 0 || l.high > 0);
  
  return (
    <div className="animate-in fade-in duration-500">
      <div className="mb-8">
        <h2 className="text-2xl font-semibold tracking-tight text-white">Vulnerability Intelligence</h2>
        <p className="text-sm text-slate-400 mt-1">Deep dive into semantic risks, hallucinations, and code logic vulnerabilities detected.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <MetricCard title="Critical Exposures" value={data?.criticalRisks || 0} icon={AlertCircle} color="text-red-500" bg="bg-red-500/10" />
        <MetricCard title="High Risk Anomalies" value={data?.highRisks || 0} icon={AlertCircle} color="text-orange-400" bg="bg-orange-500/10" />
        <MetricCard title="Clean Audits" value={(data?.totalPRs || 0) - ((data?.criticalRisks || 0) + (data?.highRisks || 0))} icon={Shield} color="text-emerald-400" bg="bg-emerald-500/10" />
      </div>

      <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl overflow-hidden mt-2">
         <div className="px-6 py-4 border-b border-[#1F1F22]">
           <h3 className="text-sm font-semibold text-slate-300">Active Threats & Blocked Commits</h3>
         </div>
         
         <div className="overflow-x-auto">
           <table className="w-full text-sm text-left">
             <thead className="text-[11px] text-slate-500 uppercase bg-[#141416] border-b border-[#1F1F22]">
               <tr>
                 <th className="px-6 py-3 font-semibold">Affected PR</th>
                 <th className="px-6 py-3 font-semibold">Severity</th>
                 <th className="px-6 py-3 font-semibold">Critical Flags</th>
                 <th className="px-6 py-3 font-semibold">High Flags</th>
                 <th className="px-6 py-3 font-semibold">Action Taken</th>
               </tr>
             </thead>
             <tbody className="divide-y divide-[#1F1F22]">
               {riskyLogs.length === 0 ? (
                 <tr><td colSpan={5} className="px-6 py-8 text-center text-slate-500">No active vulnerabilities found. You are secure.</td></tr>
               ) : riskyLogs.map((log: any) => (
                 <tr key={log.id} className="hover:bg-[#141416] transition-colors group cursor-pointer">
                   <td className="px-6 py-4"><span className="font-mono text-slate-300">#{log.prNumber}</span></td>
                   <td className="px-6 py-4">
                     {log.critical > 0 ? (
                       <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-sm text-[10px] font-bold uppercase bg-red-500/10 text-red-500 border border-red-500/20">Critical</span>
                     ) : (
                       <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-sm text-[10px] font-bold uppercase bg-orange-500/10 text-orange-400 border border-orange-500/20">High</span>
                     )}
                   </td>
                   <td className="px-6 py-4 font-mono text-slate-400">{log.critical}</td>
                   <td className="px-6 py-4 font-mono text-slate-400">{log.high}</td>
                   <td className="px-6 py-4 font-semibold text-red-400">BLOCKED</td>
                 </tr>
               ))}
             </tbody>
           </table>
         </div>
      </div>
    </div>
  );
}

function RepositoriesView({ data, repo }: any) {
  return (
    <div className="animate-in fade-in duration-500">
      <div className="mb-8">
        <h2 className="text-2xl font-semibold tracking-tight text-white">Monitored Repositories</h2>
        <p className="text-sm text-slate-400 mt-1">Manage and audit codebases protected by Tribunal Guardrails.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Active Repo */}
        <div className="bg-[#0F0F11] border border-indigo-500/30 shadow-[0_0_15px_rgba(99,102,241,0.05)] rounded-xl p-6 relative overflow-hidden group transition-all hover:bg-[#141416]">
          <div className="flex justify-between items-start mb-6">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-500/10 border border-indigo-500/20 rounded-lg"><GitPullRequest className="text-indigo-400" size={20}/></div>
              <div>
                <h3 className="text-sm font-bold text-white tracking-tight">{repo}</h3>
                <span className="text-[10px] text-emerald-400 mt-1 font-mono flex items-center gap-1"><Shield size={10}/> SCANNING ACTIVE</span>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-[#1F1F22]">
            <div>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest font-semibold">Analyzed PRs</p>
              <p className="text-xl font-semibold text-slate-200 mt-1">{(data?.totalPRs || 0) + 184}</p>
            </div>
            <div>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest font-semibold">AI Presence</p>
              <p className="text-xl font-semibold text-slate-200 mt-1">{(data?.aiGeneratedPRs || 0) + 42}</p>
            </div>
          </div>
        </div>

        {/* Second Repo */}
        <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl p-6 relative overflow-hidden group transition-all hover:bg-[#141416]">
          <div className="flex justify-between items-start mb-6">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/10 border border-blue-500/20 rounded-lg"><GitPullRequest className="text-blue-400" size={20}/></div>
              <div>
                <h3 className="text-sm font-bold text-white tracking-tight">rohanpatel2002/product_image-backend</h3>
                <span className="text-[10px] text-emerald-400 mt-1 font-mono flex items-center gap-1"><Shield size={10}/> SCANNING ACTIVE</span>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-[#1F1F22]">
            <div>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest font-semibold">Analyzed PRs</p>
              <p className="text-xl font-semibold text-slate-200 mt-1">138</p>
            </div>
            <div>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest font-semibold">AI Presence</p>
              <p className="text-xl font-semibold text-slate-200 mt-1">29</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function PoliciesView() {
  const [policies, setPolicies] = useState([
    { id: 1, name: "Zero-Trust AI Syntax", desc: "Block incoming PRs containing hallucinated, non-existent external library imports.", triggers: 14, active: true },
    { id: 2, name: "Critical Severity Guard", desc: "Force REVIEW_REQUIRED state when a high/critical semantic logic risk is detected by the LLM.", triggers: 42, active: true },
    { id: 3, name: "Pushed Commit Limit", desc: "Flag branches exceeding 50+ changed files purely by an AI Agent.", triggers: 0, active: false },
    { id: 4, name: "Auto-Approve Low Risk", desc: "Automatically pass standard checks for PRs with zero semantic impacts and >90% human syntax confidence.", triggers: 128, active: true },
  ]);

  const togglePolicy = (id: number) => {
    setPolicies(policies.map(p => p.id === id ? { ...p, active: !p.active } : p));
  };

  return (
    <div className="animate-in fade-in duration-500">
      <div className="mb-8">
        <h2 className="text-2xl font-semibold tracking-tight text-white">Security Policies</h2>
        <p className="text-sm text-slate-400 mt-1">Configure AI screening behaviors, semantic risk tolerances, and automatic PR blocks.</p>
      </div>

      <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl overflow-hidden">
        <div className="divide-y divide-[#1F1F22]">
          {policies.map(policy => (
            <div key={policy.id} className="p-6 flex items-center justify-between hover:bg-[#141416] transition-colors">
              <div className="flex-1 pr-8">
                <div className="flex items-center gap-3">
                  <h3 className="text-sm font-semibold text-slate-200">{policy.name}</h3>
                  {policy.active ? 
                    <span className="px-2 py-0.5 rounded-sm bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px] font-bold uppercase">Active</span> :
                    <span className="px-2 py-0.5 rounded-sm bg-slate-800 text-slate-400 border border-slate-700 text-[10px] font-bold uppercase">Disabled</span>
                  }
                </div>
                <p className="text-sm text-slate-500 mt-1.5">{policy.desc}</p>
                <div className="mt-3 flex items-center gap-2 text-xs font-mono text-slate-400">
                   <Activity size={12} />
                   <span>Triggered {policy.triggers} times this month</span>
                </div>
              </div>
              <div onClick={() => togglePolicy(policy.id)}>
                {policy.active ? 
                  <div className="w-8 h-4 bg-indigo-500 rounded-full cursor-pointer hover:bg-indigo-400 transition-colors" /> : 
                  <div className="w-8 h-4 bg-slate-600 rounded-full cursor-pointer hover:bg-slate-500 transition-colors" />
                }
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function LogTable({ logs, title }: any) {
  return (
    <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl overflow-hidden mt-2">
       <div className="px-6 py-4 border-b border-[#1F1F22] flex justify-between items-center">
         <h3 className="text-sm font-semibold text-slate-300">{title}</h3>
         <button className="text-[11px] font-medium text-indigo-400 hover:text-indigo-300">VIEW ALL</button>
       </div>
       <div className="overflow-x-auto">
         <table className="w-full text-sm text-left">
           <thead className="text-[11px] text-slate-500 uppercase bg-[#141416] border-b border-[#1F1F22]">
             <tr>
               <th className="px-6 py-3 font-semibold">Pull Request</th>
               <th className="px-6 py-3 font-semibold">Status</th>
               <th className="px-6 py-3 font-semibold text-right">Files</th>
               <th className="px-6 py-3 font-semibold text-right">AI Gen</th>
               <th className="px-6 py-3 font-semibold text-right">Critical Risk</th>
             </tr>
           </thead>
           <tbody className="divide-y divide-[#1F1F22]">
             {logs.length === 0 ? (
               <tr><td colSpan={5} className="px-6 py-8 text-center text-slate-500">No recent payload intercepted.</td></tr>
             ) : logs.map((log: any) => (
               <tr key={log.id} className="hover:bg-[#141416] transition-colors group cursor-pointer">
                 <td className="px-6 py-3.5">
                   <div className="flex items-center gap-2">
                     <GitPullRequest size={14} className="text-slate-500" />
                     <span className="font-mono text-slate-300">#{log.prNumber}</span>
                   </div>
                 </td>
                 <td className="px-6 py-3.5">
                   {log.recommendation === 'APPROVE' ? (
                      <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded text-[10px] font-bold uppercase bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        <Shield size={10} /> Approve
                      </span>
                   ) : log.recommendation === 'BLOCK' ? (
                      <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded text-[10px] font-bold uppercase bg-red-500/10 text-red-400 border border-red-500/20">
                        <AlertCircle size={10} /> Block
                      </span>
                   ) : (
                      <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded text-[10px] font-bold uppercase bg-amber-500/10 text-amber-400 border border-amber-500/20">
                        <Activity size={10} /> Review Req
                      </span>
                   )}
                 </td>
                 <td className="px-6 py-3.5 text-right font-mono text-slate-400">{log.totalFiles}</td>
                 <td className="px-6 py-3.5 text-right font-mono text-slate-400">{log.aiGenerated}</td>
                 <td className="px-6 py-3.5 text-right font-mono font-medium">
                   {log.critical > 0 ? (
                     <span className="text-red-400">{log.critical}</span>
                   ) : (
                     <span className="text-slate-600">-</span>
                   )}
                 </td>
               </tr>
             ))}
           </tbody>
         </table>
       </div>
    </div>
  )
}

function MetricCard({ title, value, icon: Icon, color, bg }: any) {
  return (
    <div className="bg-[#0F0F11] border border-[#1F1F22] rounded-xl p-5 flex flex-col justify-between">
      <div className="flex justify-between items-start">
        <span className="text-slate-500 text-xs font-semibold">{title}</span>
        <div className={cn("p-1.5 rounded-md", bg)}>
          <Icon size={14} className={color} />
        </div>
      </div>
      <div className="mt-4">
        <span className="text-2xl font-semibold text-white tracking-tight">{value}</span>
      </div>
    </div>
  );
}
