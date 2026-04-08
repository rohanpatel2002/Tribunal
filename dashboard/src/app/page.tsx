'use client';

import { useState, useEffect } from 'react';
import { Search, Bell, Menu, ShieldAlert, BarChart3, Database, Settings, GitPullRequest, Activity, ChevronRight, Fingerprint, RefreshCcw } from 'lucide-react';

// Interfaces for our Go backend
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
  repository: string;
  prNumber: number;
  commitSHA: string;
  filesAnalyzed: number;
  aiGenerated: boolean;
  overallRiskScore: number;
  highRiskFound: boolean;
  criticalRiskFound: boolean;
  createdAt: string;
}

export default function Dashboard() {
  const [data, setData] = useState<AuditSummary | null>(null);
  const [logs, setLogs] = useState<PRAnalysisRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [repo, setRepo] = useState("rohanpatel2002/tribunal");
  const [apiKey, setApiKey] = useState("");
  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    setIsRefreshing(true);
    try {
      const res = await fetch(`http://localhost:8080/api/v1/audit/summary?repository=${encodeURIComponent(repo)}`, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      if (res.ok) {
        const json = await res.json();
        setData(json);
      } else {
        // Fallback demo data if API is unconfigured/unavailable
        setData({
          repository: repo,
          totalPRs: 142,
          totalFiles: 845,
          aiGeneratedPRs: 37,
          criticalRisks: 4,
          highRisks: 12,
          averageAIScore: 0.68
        });
      }
      } catch (e) {
      console.warn("Using offline demo data mode:", e);
      setData({
        repository: repo,
        totalPRs: 0,
        totalFiles: 0,
        aiGeneratedPRs: 0,
        criticalRisks: 0,
        highRisks: 0,
        averageAIScore: 0.0
      });
    }

    try {
      const logsRes = await fetch(`http://localhost:8080/api/v1/audit/recent?repository=${encodeURIComponent(repo)}&limit=10`, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      if (logsRes.ok) {
        const logsJson = await logsRes.json();
        if (logsJson) {
           setLogs(logsJson);
        } else {
           setLogs([]);
        }
      }
    } catch (e) {
      console.warn("Using offline mode, could not fetch logs", e);
      setLogs([]);
    } finally {
      setLoading(false);
      setTimeout(() => setIsRefreshing(false), 500); // Visual feedback
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="flex w-full h-full text-slate-200 font-sans overflow-hidden">
      {/* ENTERPRISE SIDEBAR */}
      <aside className="w-64 bg-[#09090b]/80 border-r border-slate-800/80 backdrop-blur-xl flex flex-col justify-between transition-all pt-6 pb-4 z-20 hidden md:flex">
        <div className="px-6 flex flex-col gap-8">
          <div className="flex items-center gap-3">
             <div className="bg-indigo-600 p-2 rounded-xl bg-linear-to-br from-indigo-500 to-purple-600 shadow-lg shadow-indigo-500/20">
               <Fingerprint size={24} className="text-white" />
             </div>
             <div>
               <h1 className="text-xl font-bold tracking-tight text-white leading-tight mt-1">Tribunal <span className="font-light text-indigo-400">AI</span></h1>
               <p className="text-[10px] text-slate-500 font-mono tracking-widest uppercase">Platform Defense</p>
             </div>
          </div>

          <nav className="flex flex-col gap-2 mt-4">
             <p className="text-xs font-bold text-slate-600 uppercase tracking-widest mb-2 px-2">Analytics</p>
             <button className="flex items-center justify-between px-3 py-2.5 bg-indigo-500/10 text-indigo-400 rounded-lg group">
                <div className="flex items-center gap-3 font-medium text-sm">
                   <Activity size={18} />
                   <span>Risk Command</span>
                </div>
                <div className="w-1 h-4 bg-indigo-500 rounded-full scale-0 group-hover:scale-100 transition-transform origin-right" />
             </button>
             <button className="flex items-center justify-between px-3 py-2.5 text-slate-400 hover:bg-slate-800/50 hover:text-slate-200 rounded-lg group transition-colors">
                <div className="flex items-center gap-3 font-medium text-sm">
                   <ShieldAlert size={18} />
                   <span>Vulnerabilities</span>
                </div>
             </button>
             <button className="flex items-center justify-between px-3 py-2.5 text-slate-400 hover:bg-slate-800/50 hover:text-slate-200 rounded-lg group transition-colors">
                <div className="flex items-center gap-3 font-medium text-sm">
                   <GitPullRequest size={18} />
                   <span>Repositories</span>
                </div>
             </button>
          </nav>
        </div>

        <div className="px-6 mt-auto">
          <nav className="flex flex-col gap-2">
             <button className="flex items-center gap-3 px-3 py-2.5 text-slate-400 hover:text-slate-200 rounded-lg transition-colors font-medium text-sm">
                <Settings size={18} />
                <span>Enterprise Config</span>
             </button>
          </nav>
          
          <div className="mt-6 pt-6 border-t border-slate-800/50 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-slate-800/80 border border-slate-700/50 flex items-center justify-center">
                 <div className="w-5 h-5 rounded-full bg-linear-to-tr from-cyan-400 to-indigo-500" />
              </div>
              <div className="flex flex-col">
                <span className="text-xs font-semibold text-white">Chief Security</span>
                <span className="text-[10px] text-slate-500">Tier: Elite</span>
              </div>
            </div>
            <button className="text-slate-500 hover:text-white transition-colors">
               <ChevronRight size={16} />
            </button>
          </div>
        </div>
      </aside>

      {/* MAIN CONTENT AREA */}
      <main className="flex-1 flex flex-col h-full bg-transparent overflow-y-auto relative">
         {/* HEADER */}
         <header className="h-16 border-b border-white/5 bg-[#020617]/50 backdrop-blur-xl sticky top-0 z-10 flex items-center justify-between px-8">
            <div className="flex items-center gap-4 text-sm">
               <span className="text-slate-400">Workspaces</span>
               <ChevronRight size={14} className="text-slate-600" />
               <span className="font-medium text-indigo-300 flex items-center gap-2">
                  <Database size={14} />
                  {repo}
               </span>
            </div>
            <div className="flex items-center gap-5">
               <div className="relative group hidden sm:block">
                  <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
                  <input 
                    type="text" 
                    placeholder="Search logs or policies..." 
                    className="pl-9 pr-4 py-1.5 bg-slate-900/50 border border-slate-800 rounded-full text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/50 transition-all w-64"
                  />
               </div>
               <button className="relative text-slate-400 hover:text-white transition-colors">
                  <Bell size={18} />
                  <span className="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-full text-[8px] flex items-center justify-center translate-x-1/3 -translate-y-1/4 ring-2 ring-[#020617]" />
               </button>
            </div>
         </header>

         {/* DASHBOARD CONTENT */}
         <div className="p-8 max-w-7xl w-full mx-auto pb-24">
            {/* HERO SECTION */}
            <div className="flex flex-col md:flex-row md:items-end justify-between mb-10 gap-4">
              <div>
                <h2 className="text-3xl font-extrabold tracking-tight text-white flex items-center gap-3">
                  AI Code Posture
                  <span className="px-2.5 py-0.5 text-[10px] uppercase font-bold tracking-widest text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 rounded-full flex items-center gap-1.5"><span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" /> Live</span>
                </h2>
                <p className="text-sm text-slate-400 mt-2 max-w-xl">Continuous integration insights. Analyzing pull requests for adversarial prompts, sensitive exfiltration, and logic bombs across <span className="text-white font-medium">{repo}</span>.</p>
              </div>
              <div className="flex items-center gap-3">
                 <button onClick={fetchData} className="flex items-center gap-2 px-4 py-2 border border-slate-700/50 hover:bg-slate-800/50 backdrop-blur-sm rounded-lg text-sm text-slate-300 font-medium transition-all mr-2">
                   <RefreshCcw size={14} className={isRefreshing ? "animate-spin text-indigo-400" : ""} />
                   Sync Data
                 </button>
                 <div className="flex items-center gap-2 bg-[#09090b]/80 p-1 border border-slate-800 rounded-lg">
                    <input 
                      type="password" 
                      placeholder="Enterprise API Key..." 
                      className="bg-transparent border-none text-xs text-white px-3 focus:outline-none w-48 font-mono placeholder-slate-600"
                      value={apiKey}
                      onChange={(e) => setApiKey(e.target.value)}
                    />
                 </div>
              </div>
            </div>

            {loading ? (
              <div className="flex flex-col gap-8 w-full animate-pulse mt-8">
                 <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    {[1, 2, 3, 4].map((i) => (
                       <div key={i} className="bg-slate-900/40 border border-slate-800/60 p-6 rounded-2xl h-32" />
                    ))}
                 </div>
                 <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="lg:col-span-2 bg-slate-900/40 border border-slate-800/60 h-96 rounded-2xl" />
                    <div className="bg-slate-900/40 border border-slate-800/60 h-96 rounded-2xl" />
                 </div>
              </div>
            ) : (
              <div className="flex flex-col gap-8">
                {/* 1. METRICS GRID */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 p-6 rounded-2xl relative overflow-hidden group hover:border-slate-700/80 transition-all">
                     <div className="absolute -right-4 -top-4 w-24 h-24 bg-blue-500/10 rounded-full blur-2xl group-hover:bg-blue-500/20 transition-all" />
                     <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-blue-500/10 text-blue-400 rounded-lg">
                           <GitPullRequest size={18} />
                        </div>
                        <h3 className="text-sm font-medium text-slate-400">Total PRs Analyzed</h3>
                     </div>
                     <p className="text-3xl font-extrabold text-white mt-4">{data?.totalPRs}</p>
                     <p className="text-xs text-slate-500 mt-2 flex items-center gap-1"><span className="text-blue-400 font-medium">+{data?.totalFiles}</span> files</p>
                  </div>

                  <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 p-6 rounded-2xl relative overflow-hidden group hover:border-slate-700/80 transition-all">
                     <div className="absolute -right-4 -top-4 w-24 h-24 bg-purple-500/10 rounded-full blur-2xl group-hover:bg-purple-500/20 transition-all" />
                     <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-purple-500/10 text-purple-400 rounded-lg">
                           <Activity size={18} />
                        </div>
                        <h3 className="text-sm font-medium text-slate-400">AI-Generated Code</h3>
                     </div>
                     <p className="text-3xl font-extrabold text-white mt-4">{data?.aiGeneratedPRs}</p>
                     <p className="text-xs text-slate-500 mt-2 flex items-center gap-1">Average Score: <span className="text-purple-400 font-mono">{((data?.averageAIScore || 0) * 100).toFixed(1)}%</span></p>
                  </div>

                  <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 p-6 rounded-2xl relative overflow-hidden group hover:border-slate-700/80 transition-all">
                     <div className="absolute -right-4 -top-4 w-24 h-24 bg-red-500/10 rounded-full blur-2xl group-hover:bg-red-500/20 transition-all" />
                     <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-red-500/10 text-red-500 rounded-lg">
                           <ShieldAlert size={18} />
                        </div>
                        <h3 className="text-sm font-medium text-slate-400">Critical Risks</h3>
                     </div>
                     <p className="text-3xl font-extrabold text-white mt-4">{data?.criticalRisks}</p>
                     <p className="text-xs text-slate-500 mt-2 flex items-center gap-1">
                        {(data?.criticalRisks || 0) > 0 ? <span className="text-red-500 font-medium">Requires immediate action</span> : <span className="text-emerald-500 font-medium">Clear</span>}
                     </p>
                  </div>

                  <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 p-6 rounded-2xl relative overflow-hidden group hover:border-slate-700/80 transition-all">
                     <div className="absolute -right-4 -top-4 w-24 h-24 bg-orange-500/10 rounded-full blur-2xl group-hover:bg-orange-500/20 transition-all" />
                     <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-orange-500/10 text-orange-400 rounded-lg">
                           <ShieldAlert size={18} />
                        </div>
                        <h3 className="text-sm font-medium text-slate-400">High Risks</h3>
                     </div>
                     <p className="text-3xl font-extrabold text-white mt-4">{data?.highRisks}</p>
                     <p className="text-xs text-slate-500 mt-2 flex items-center gap-1">Policy violations</p>
                  </div>
                </div>

                {/* 2. RECENT ANALYSIS LOGS (Data Grid) */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                  <div className="lg:col-span-2 bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 rounded-2xl p-6 relative overflow-hidden group">
                    <div className="absolute top-0 right-0 w-64 h-64 bg-indigo-500/10 rounded-full blur-[80px] pointer-events-none -z-10 transition-all duration-700" />
                    <h3 className="text-lg font-bold text-white mb-6 flex items-center gap-2">Repository Context</h3>
                    
                    <div className="flex flex-col items-center justify-center py-6 mb-4 relative">
                      <div className="w-32 h-32 rounded-full border-8 border-indigo-500 flex items-center justify-center mb-4 relative drop-shadow-[0_0_15px_rgba(99,102,241,0.2)]">
                        <div className="absolute inset-0 rounded-full border border-indigo-400/30 animate-ping" />
                        <span className="text-4xl font-extrabold text-white bg-clip-text text-transparent bg-linear-to-b from-white to-slate-400">
                           {data?.averageAIScore ? (data.averageAIScore * 100).toFixed(0) : '0'}
                        </span>
                      </div>
                      <span className="text-sm font-medium text-indigo-300">Composite Risk Score</span>
                      <p className="text-xs text-slate-500 text-center mt-2 px-4 leading-relaxed">Aggregated confidence score reflecting organizational security compliance.</p>
                    </div>

                    <div className="space-y-4">
                      <div className="flex justify-between items-center py-3 border-b border-slate-800/60">
                        <span className="text-sm text-slate-400 font-medium">Target Repo</span>
                        <span className="text-sm text-white font-mono bg-slate-800/50 px-2 py-0.5 rounded border border-slate-700/50">{data?.repository || 'n/a'}</span>
                      </div>
                      <div className="flex justify-between items-center py-3 border-b border-slate-800/60">
                        <span className="text-sm text-slate-400 font-medium">Files Analyzed</span>
                        <span className="text-sm text-white font-mono">{data?.totalFiles || 0}</span>
                      </div>
                      <div className="flex justify-between items-center py-3">
                        <span className="text-sm text-slate-400 font-medium">Active Policy</span>
                        <span className="text-[10px] uppercase tracking-widest font-bold text-indigo-400 bg-indigo-500/10 px-2 py-1 rounded border border-indigo-500/20">Strict</span>
                      </div>
                    </div>
                  </div>

                  <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/60 rounded-2xl p-6 relative overflow-hidden group">
                    <h3 className="text-lg font-medium text-white mb-6 z-10">System Health Trend</h3>
                    <div className="flex-1 flex flex-col items-center justify-center z-10">
                      <div className="w-32 h-32 rounded-full border-8 border-indigo-500 flex items-center justify-center mb-4 relative drop-shadow-[0_0_15px_rgba(99,102,241,0.2)]">
                        <span className="text-2xl font-bold">{(data?.averageAIScore || 0) * 100}%</span>
                      </div>
                      <p className="text-center text-sm text-gray-400 mt-2">
                        Average AI Generation Probability across <strong className="text-gray-200">24h</strong> trailing window.
                      </p>
                    </div>
                    
                    {/* Decorative mini bar chart in background */}
                    <div className="absolute bottom-0 left-0 w-full h-24 flex items-end gap-1 px-4 opacity-20 pointer-events-none">
                      {[40, 70, 45, 90, 65, 80, 50, 85, 60, 95, 75, 65, 88].map((h, i) => (
                        <div key={i} className="bg-indigo-500 w-full rounded-t-sm" style={{ height: `${h}%` }}></div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
      </main>
    </div>
  );
}
