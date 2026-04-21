'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Settings,
  GitPullRequest,
  Activity,
  ChevronRight,
  Fingerprint,
  RefreshCcw,
  Lock,
  AlertCircle,
  Zap,
  Shield,
  Download,
  ChevronDown,
  Sparkles,
  TrendingUp,
  AlertTriangle,
  CheckCircle2,
  type LucideIcon,
} from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  LineChart,
  Line,
  AreaChart,
  Area,
} from 'recharts';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const TABS = [
  { id: 'overview', label: 'Overview', icon: Activity },
  { id: 'analysis', label: 'Risk Analysis', icon: AlertCircle },
  { id: 'policies', label: 'Security Policies', icon: Lock },
  { id: 'settings', label: 'Settings', icon: Settings },
];

export default function PremiumDashboard() {
  const [activeTab, setActiveTab] = useState('overview');

  return (
    <div className="flex w-full h-full font-sans overflow-hidden">
      {/* DARK GRADIENT BACKGROUND */}
      <div className="fixed inset-0 bg-gradient-to-br from-slate-950 via-purple-950 to-slate-950 -z-10" />
      
      {/* ANIMATED BACKGROUND ELEMENTS */}
      <div className="fixed inset-0 -z-10 overflow-hidden">
        <div className="absolute top-0 right-0 w-96 h-96 bg-emerald-500/5 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-blue-500/5 rounded-full blur-3xl animate-pulse animation-delay-2000" />
        <div className="absolute top-1/2 left-1/2 w-96 h-96 bg-purple-500/5 rounded-full blur-3xl animate-pulse animation-delay-4000" />
      </div>

      {/* SIDEBAR - PREMIUM GLASS */}
      <aside className="w-80 bg-gradient-to-b from-slate-900/60 to-slate-950/80 backdrop-blur-2xl border-r border-emerald-500/10 flex flex-col justify-between py-8 px-8 shadow-2xl">
        {/* LOGO SECTION */}
        <div className="flex flex-col gap-10">
          <div className="flex items-center gap-4 group">
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-br from-emerald-400 to-green-600 rounded-3xl blur-lg opacity-75 group-hover:opacity-100 transition-opacity duration-300" />
              <div className="relative bg-gradient-to-br from-emerald-400 to-green-600 p-4 rounded-3xl shadow-xl">
                <Fingerprint size={28} className="text-white" strokeWidth={1.5} />
              </div>
            </div>
            <div>
              <div className="text-3xl font-black tracking-tighter text-white leading-none">
                <span className="bg-gradient-to-r from-emerald-300 via-green-400 to-emerald-300 bg-clip-text text-transparent">
                  TRIBUNAL
                </span>
              </div>
              <p className="text-xs text-emerald-400/60 font-mono tracking-[0.2em] uppercase font-bold mt-2">
                ⚡ Security Audit
              </p>
            </div>
          </div>

          {/* NAVIGATION */}
          <nav className="flex flex-col gap-2">
            <p className="text-[10px] font-bold text-emerald-400/50 uppercase tracking-widest mb-4 px-4">
              Navigation
            </p>
            {TABS.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  'flex items-center gap-4 px-4 py-3.5 rounded-xl text-sm font-semibold transition-all duration-300 group relative overflow-hidden',
                  activeTab === tab.id
                    ? 'bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 shadow-lg shadow-emerald-500/20 border border-emerald-500/30'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-white/5 border border-transparent'
                )}
              >
                <div className={cn(
                  'p-2 rounded-lg transition-all duration-300',
                  activeTab === tab.id
                    ? 'bg-emerald-500/20'
                    : 'bg-slate-800/50 group-hover:bg-slate-700/50'
                )}>
                  <tab.icon size={18} />
                </div>
                <span>{tab.label}</span>
                {activeTab === tab.id && (
                  <Sparkles size={16} className="ml-auto text-emerald-400 animate-pulse" />
                )}
              </button>
            ))}
          </nav>
        </div>

        {/* FOOTER STATS */}
        <div className="flex flex-col gap-4 pt-8 border-t border-emerald-500/10">
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-white/5 backdrop-blur border border-emerald-500/10 rounded-xl p-3">
              <p className="text-[10px] text-slate-400 uppercase font-bold tracking-wide">Status</p>
              <p className="text-lg font-black text-emerald-400 mt-1 flex items-center gap-1">
                <span className="w-2 h-2 bg-emerald-400 rounded-full animate-pulse" />
                ACTIVE
              </p>
            </div>
            <div className="bg-white/5 backdrop-blur border border-emerald-500/10 rounded-xl p-3">
              <p className="text-[10px] text-slate-400 uppercase font-bold tracking-wide">Version</p>
              <p className="text-lg font-black text-blue-400 mt-1">v2.1.0</p>
            </div>
          </div>
          
          <button className="w-full bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-400 hover:to-green-500 text-white font-bold py-3 rounded-xl transition-all duration-300 shadow-lg shadow-emerald-500/30 hover:shadow-emerald-500/50 uppercase text-xs tracking-wider">
            Update Available
          </button>
        </div>
      </aside>

      {/* MAIN CONTENT */}
      <main className="flex-1 flex flex-col h-full overflow-y-auto">
        {/* HEADER */}
        <header className="h-20 border-b border-emerald-500/10 bg-gradient-to-r from-slate-900/50 to-slate-950/50 backdrop-blur-xl sticky top-0 z-10 flex items-center justify-between px-12 gap-6 shadow-xl">
          <div className="flex items-center gap-4 flex-1">
            <div className="flex items-center gap-3">
              <GitPullRequest className="text-emerald-400" size={20} />
              <div>
                <p className="text-xs text-slate-500 uppercase font-bold tracking-wider">Target Repository</p>
                <p className="text-lg font-bold text-white">rohanpatel2002/tribunal</p>
              </div>
            </div>
            <ChevronRight className="text-slate-600" size={20} />
          </div>

          <div className="flex items-center gap-4 ml-auto">
            <div className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/5 border border-emerald-500/10 backdrop-blur">
              <span className="text-xs text-slate-400">GitHub Connected</span>
              <span className="w-2 h-2 bg-emerald-400 rounded-full animate-pulse" />
            </div>
            <button className="p-2.5 rounded-xl bg-white/5 border border-emerald-500/10 hover:bg-white/10 transition-all text-slate-300 hover:text-emerald-300">
              <RefreshCcw size={18} />
            </button>
            <button className="p-2.5 rounded-xl bg-white/5 border border-emerald-500/10 hover:bg-white/10 transition-all text-slate-300 hover:text-emerald-300">
              <Download size={18} />
            </button>
          </div>
        </header>

        {/* CONTENT AREA */}
        <div className="flex-1 p-12 overflow-y-auto">
          {/* HERO SECTION */}
          <div className="mb-12">
            <div className="flex items-center gap-3 mb-2">
              <Sparkles className="text-emerald-400" size={24} />
              <h2 className="text-4xl font-black text-white tracking-tight">
                Risk Command
                <span className="bg-gradient-to-r from-emerald-400 to-green-400 bg-clip-text text-transparent"> Center</span>
              </h2>
            </div>
            <p className="text-slate-400 text-lg">Real-time security analysis and threat detection across your pull requests.</p>
          </div>

          {/* KPI CARDS - PREMIUM STYLE */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
            {[
              { title: 'Pull Requests Analyzed', value: '2,847', icon: GitPullRequest, trend: '+12%', color: 'from-blue-500 to-blue-600' },
              { title: 'AI-Generated Detection', value: '342', icon: Zap, trend: '+8%', color: 'from-yellow-500 to-orange-600' },
              { title: 'Critical Risks', value: '18', icon: AlertTriangle, trend: '-4%', color: 'from-red-500 to-red-600' },
              { title: 'Security Score', value: '94%', icon: Shield, trend: '+2%', color: 'from-emerald-500 to-green-600' },
            ].map((kpi, i) => (
              <div key={i} className="group relative">
                <div className="absolute inset-0 bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300 blur" />
                <div className="relative bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-xl border border-emerald-500/10 rounded-2xl p-6 overflow-hidden hover:border-emerald-500/30 transition-all duration-300 shadow-xl hover:shadow-2xl hover:shadow-emerald-500/20">
                  {/* Gradient accent */}
                  <div className={`absolute top-0 right-0 w-32 h-32 bg-gradient-to-br ${kpi.color} opacity-5 rounded-full blur-2xl`} />
                  
                  <div className="relative">
                    <div className="flex items-start justify-between mb-4">
                      <div className={`p-3 rounded-xl bg-gradient-to-br ${kpi.color} bg-opacity-10`}>
                        <kpi.icon className="text-white" size={20} />
                      </div>
                      <span className="text-xs font-bold text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-full">
                        {kpi.trend}
                      </span>
                    </div>
                    <p className="text-slate-400 text-sm font-semibold mb-2">{kpi.title}</p>
                    <p className="text-4xl font-black text-white mb-1">{kpi.value}</p>
                    <div className="h-1 w-full bg-slate-700 rounded-full overflow-hidden">
                      <div className={`h-full bg-gradient-to-r ${kpi.color} w-3/4 rounded-full`} />
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* CHARTS SECTION */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* MAIN CHART */}
            <div className="lg:col-span-2 bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-xl border border-emerald-500/10 rounded-2xl p-8 shadow-xl hover:border-emerald-500/30 transition-all duration-300">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h3 className="text-lg font-bold text-white flex items-center gap-2">
                    <TrendingUp className="text-emerald-400" size={20} />
                    Pipeline Activity
                  </h3>
                  <p className="text-xs text-slate-400 mt-1">Last 12 months performance</p>
                </div>
                <ChevronDown className="text-slate-500" size={20} />
              </div>
              <div className="h-64 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={[
                    { month: 'Jan', analyzed: 240, ai: 24, risk: 12 },
                    { month: 'Feb', analyzed: 340, ai: 34, risk: 15 },
                    { month: 'Mar', analyzed: 440, ai: 44, risk: 18 },
                    { month: 'Apr', analyzed: 380, ai: 38, risk: 14 },
                    { month: 'May', analyzed: 520, ai: 52, risk: 22 },
                    { month: 'Jun', analyzed: 620, ai: 62, risk: 28 },
                  ]}>
                    <defs>
                      <linearGradient id="colorAnalyzed" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                    <XAxis dataKey="month" stroke="#94a3b8" fontSize={12} />
                    <YAxis stroke="#94a3b8" fontSize={12} />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #10b981', borderRadius: '12px' }}
                      labelStyle={{ color: '#f1f5f9' }}
                    />
                    <Area type="monotone" dataKey="analyzed" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorAnalyzed)" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* STATS WIDGET */}
            <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-xl border border-emerald-500/10 rounded-2xl p-8 shadow-xl hover:border-emerald-500/30 transition-all duration-300">
              <h3 className="text-lg font-bold text-white mb-6 flex items-center gap-2">
                <AlertCircle className="text-emerald-400" size={20} />
                Quick Stats
              </h3>
              
              <div className="space-y-4">
                {[
                  { label: 'Approval Rate', value: '94%', color: 'from-emerald-500 to-green-600' },
                  { label: 'Avg. Scan Time', value: '2.3s', color: 'from-blue-500 to-blue-600' },
                  { label: 'Threat Level', value: 'LOW', color: 'from-yellow-500 to-orange-600' },
                  { label: 'Policy Violations', value: '3', color: 'from-red-500 to-red-600' },
                ].map((stat, i) => (
                  <div key={i} className="group">
                    <div className="flex items-center justify-between mb-2">
                      <p className="text-sm text-slate-400 font-semibold">{stat.label}</p>
                      <p className="text-xl font-black text-white group-hover:text-emerald-300 transition-colors">{stat.value}</p>
                    </div>
                    <div className="h-1.5 bg-slate-700 rounded-full overflow-hidden">
                      <div className={`h-full bg-gradient-to-r ${stat.color} w-4/5 rounded-full`} />
                    </div>
                  </div>
                ))}
              </div>

              <button className="w-full mt-8 bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-400 hover:to-green-500 text-white font-bold py-3 rounded-xl transition-all duration-300 shadow-lg shadow-emerald-500/30 hover:shadow-emerald-500/50 uppercase text-xs tracking-wider">
                View Detailed Report
              </button>
            </div>
          </div>

          {/* FOOTER */}
          <div className="mt-12 pt-8 border-t border-emerald-500/10 flex items-center justify-between text-xs text-slate-400">
            <p>© 2026 Tribunal. All security rights reserved.</p>
            <p>Last updated: Just now</p>
          </div>
        </div>
      </main>
    </div>
  );
}
