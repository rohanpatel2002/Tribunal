'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  Search,
  Bell,
  ShieldAlert,
  BarChart3,
  Settings,
  GitPullRequest,
  Activity,
  ChevronRight,
  Fingerprint,
  RefreshCcw,
  Lock,
  Box,
  CheckCircle2,
  ChevronDown,
  CheckCircle,
  Calendar,
  Filter,
  AlertCircle,
  Loader2,
  Zap,
  Shield,
  Download,
  Clock,
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
} from 'recharts';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import ErrorBoundary from '@/components/ErrorBoundary';
import { DateRangeFilter } from '@/components/DateRangeFilter';
import { RepositorySelector } from '@/components/RepositorySelector';
import {
  fetchAuditSummary,
  fetchAuditLogs,
  getDemoAuditSummary,
  getDemoPRAnalysisRecords,
  type AuditSummary,
  type PRAnalysisRecord,
  type FilterParams,
} from '@/lib/api';
import {
  exportToCSV,
  exportToJSON,
  exportToTSV,
  generateHTMLReport,
  getTimestampedFilename,
} from '@/lib/exportUtils';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const TABS = [
  { id: 'overview', label: 'Overview', icon: Activity },
  { id: 'analysis', label: 'Risk Analysis', icon: AlertCircle },
  { id: 'policies', label: 'Security Policies', icon: Lock },
  { id: 'settings', label: 'Settings', icon: Settings },
];

function DashboardContent() {
  const [data, setData] = useState<AuditSummary | null>(null);
  const [logs, setLogs] = useState<PRAnalysisRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [repo, setRepo] = useState('rohanpatel2002/tribunal');
  const [apiKey, setApiKey] = useState('dev_enterprise_key_123');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');

  // ============= PHASE 2: FILTERING =============
  const [filters, setFilters] = useState<FilterParams>({});
  const [showFilters, setShowFilters] = useState(false);

  // ============= PHASE 2: PAGINATION =============
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const itemsPerPage = pageSize;

  // ============= PHASE 3: DATE RANGE FILTERING =============
  const [dateRangeOpen, setDateRangeOpen] = useState(false);
  const [dateFilter, setDateFilter] = useState<{ startDate: Date | null; endDate: Date | null }>({
    startDate: null,
    endDate: null,
  });

  // ============= PHASE 3: REPOSITORY SELECTION =============
  const [repositories, setRepositories] = useState<any[]>([
    { id: '1', name: 'tribunal', fullName: 'rohanpatel2002/tribunal', url: 'https://github.com/rohanpatel2002/tribunal', isMonitored: true },
    { id: '2', name: 'another-repo', fullName: 'rohanpatel2002/another-repo', url: 'https://github.com/rohanpatel2002/another-repo', isMonitored: true },
  ]);

  // ============= PHASE 3: EXPORT STATE =============
  const [showExportMenu, setShowExportMenu] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    setIsRefreshing(true);

    try {
      // Fetch summary
      const summary = await fetchAuditSummary(repo, apiKey);
      if (summary) {
        setData(summary);
      } else {
        // Use demo data as fallback
        console.info('Using demo data (API unavailable)');
        setData(getDemoAuditSummary(repo));
      }

      // Fetch logs with pagination
      const offset = (currentPage - 1) * itemsPerPage;
      const logsData = await fetchAuditLogs(
        repo,
        apiKey,
        { offset, limit: itemsPerPage },
        filters
      );

      if (logsData) {
        setLogs(logsData);
      } else {
        // Use demo data as fallback
        setLogs(getDemoPRAnalysisRecords(repo));
      }
    } catch (error) {
      console.error('Error fetching data:', error);
      // Gracefully fall back to demo data
      setData(getDemoAuditSummary(repo));
      setLogs(getDemoPRAnalysisRecords(repo));
    } finally {
      setLoading(false);
      setTimeout(() => setIsRefreshing(false), 500);
    }
  };

  useEffect(() => {
    fetchData();
  }, [repo, currentPage, pageSize, filters]);

  // ============= CHART DATA =============
  const chartData = useMemo(() => {
    return logs
      .slice()
      .reverse()
      .map((log) => ({
        name: `PR #${log.prNumber}`,
        Files: log.totalFiles,
        AI: log.aiGenerated,
        Risks: log.critical + log.high,
      }));
  }, [logs]);

  // ============= PAGINATION HELPERS =============
  const totalPages = Math.ceil((logs.length > 0 ? logs.length * 3 : 0) / itemsPerPage) || 1;
  const canPreviousPage = currentPage > 1;
  const canNextPage = currentPage < totalPages;

  const handleApplySeverityFilter = (severity: string) => {
    setFilters((prev) => ({
      ...prev,
      severity: severity as any,
    }));
    setCurrentPage(1);
  };

  const handleClearFilters = () => {
    setFilters({});
    setCurrentPage(1);
  };

  // ============= PHASE 3: DATE RANGE HANDLERS =============
  const handleApplyDateFilter = (startDate: Date | null, endDate: Date | null) => {
    setDateFilter({ startDate, endDate });
    setCurrentPage(1);
    setDateRangeOpen(false);
  };

  const handleClearDateFilter = () => {
    setDateFilter({ startDate: null, endDate: null });
    setCurrentPage(1);
  };

  // ============= PHASE 3: REPOSITORY HANDLER =============
  const handleSelectRepository = (name: string, fullName: string) => {
    setRepo(fullName);
    setCurrentPage(1);
  };

  // ============= PHASE 3: EXPORT HANDLERS =============
  const handleExportCSV = () => {
    exportToCSV(logs, getTimestampedFilename('audit-logs.csv'));
    setShowExportMenu(false);
  };

  const handleExportJSON = () => {
    exportToJSON(logs, { repository: repo, exportDate: new Date().toISOString() }, getTimestampedFilename('audit-logs.json'));
    setShowExportMenu(false);
  };

  const handleExportTSV = () => {
    exportToTSV(logs, getTimestampedFilename('audit-logs.tsv'));
    setShowExportMenu(false);
  };

  const handleExportHTML = () => {
    const html = generateHTMLReport(logs, data);
    const blob = new Blob([html], { type: 'text/html;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = window.URL.createObjectURL(blob);
    link.download = getTimestampedFilename('audit-report.html');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(link.href);
    setShowExportMenu(false);
  };

  return (
    <div className="flex w-full h-full text-slate-200 font-sans overflow-hidden bg-[#0A0A0A]">
      {/* SIDEBAR */}
      <aside className="w-64 bg-[#0F0F11] border-r border-[#1F1F22] flex flex-col justify-between pt-6 pb-4">
        <div className="px-5 flex flex-col gap-8">
          <div className="flex items-center gap-3">
            <div className="bg-indigo-600/10 border border-indigo-500/20 p-2 rounded-xl">
              <Fingerprint size={20} className="text-indigo-400" />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight text-white leading-tight">
                Tribunal
              </h1>
              <p className="text-[10px] text-slate-500 font-mono tracking-widest uppercase">
                Guardrails
              </p>
            </div>
          </div>

          <nav className="flex flex-col gap-1 mt-2">
            <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-2 px-3">
              Main Menu
            </p>
            {TABS.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  'flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-all group',
                  activeTab === tab.id
                    ? 'bg-indigo-500/10 text-indigo-400 font-medium'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-[#1A1A1E]'
                )}
              >
                <div className="flex items-center gap-3">
                  <tab.icon
                    size={16}
                    className={
                      activeTab === tab.id
                        ? 'text-indigo-400'
                        : 'text-slate-500 group-hover:text-slate-300'
                    }
                  />
                  <span>{tab.label}</span>
                </div>
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
                <div className="w-4 h-4 rounded-full bg-gradient-to-tr from-cyan-400 to-indigo-500" />
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
         <header className="h-16 border-b border-[#1F1F22] bg-[#0A0A0A]/80 backdrop-blur-xl sticky top-0 z-10 flex items-center justify-between px-8 gap-4">
            <div className="flex items-center gap-3 text-sm min-w-0">
               <span className="text-slate-500 whitespace-nowrap">Target</span>
               <ChevronRight size={14} className="text-slate-700 flex-shrink-0" />
               <div className="min-w-0 flex-1 max-w-xs">
                  <RepositorySelector
                    selectedRepo={repo}
                    onSelectRepo={handleSelectRepository}
                    repositories={repositories}
                  />
               </div>
            </div>

            <div className="flex items-center gap-2 ml-auto">
               {/* Phase 3: Date Range Filter */}
               <DateRangeFilter
                 isOpen={dateRangeOpen}
                 onToggle={() => setDateRangeOpen(!dateRangeOpen)}
                 onApplyFilter={handleApplyDateFilter}
                 onClearFilter={handleClearDateFilter}
               />

               {/* Phase 3: Export Menu */}
               <div className="relative">
                  <button
                    onClick={() => setShowExportMenu(!showExportMenu)}
                    className="inline-flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium bg-slate-500/10 text-slate-400 border border-slate-500/20 hover:border-slate-500/40 transition-all"
                  >
                    <Download size={14} />
                    <span>Export</span>
                  </button>

                  {showExportMenu && (
                    <div className="absolute right-0 mt-2 w-32 bg-[#0F0F11] border border-[#1F1F22] rounded-lg shadow-lg z-50">
                      <button
                        onClick={handleExportCSV}
                        className="w-full px-3 py-2 text-left text-xs text-slate-300 hover:bg-[#1A1A1E] transition-colors border-b border-[#1F1F22]"
                      >
                        CSV
                      </button>
                      <button
                        onClick={handleExportJSON}
                        className="w-full px-3 py-2 text-left text-xs text-slate-300 hover:bg-[#1A1A1E] transition-colors border-b border-[#1F1F22]"
                      >
                        JSON
                      </button>
                      <button
                        onClick={handleExportTSV}
                        className="w-full px-3 py-2 text-left text-xs text-slate-300 hover:bg-[#1A1A1E] transition-colors border-b border-[#1F1F22]"
                      >
                        TSV
                      </button>
                      <button
                        onClick={handleExportHTML}
                        className="w-full px-3 py-2 text-left text-xs text-slate-300 hover:bg-[#1A1A1E] transition-colors"
                      >
                        HTML Report
                      </button>
                    </div>
                  )}
               </div>

               <button
                 onClick={fetchData}
                 className="flex items-center justify-center p-1.5 text-slate-400 hover:bg-[#1A1A1E] hover:text-white rounded border border-transparent hover:border-[#27272A] transition-all"
               >
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
               ) : riskyLogs.map((log: any, index: number) => (
                 <tr key={`${log.id ?? 'log'}-${log.prNumber ?? 'pr'}-${index}`} className="hover:bg-[#141416] transition-colors group cursor-pointer">
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
             ) : logs.map((log: any, index: number) => (
               <tr key={`${log.id ?? 'log'}-${log.prNumber ?? 'pr'}-${index}`} className="hover:bg-[#141416] transition-colors group cursor-pointer">
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

export default function Dashboard() {
  return (
    <ErrorBoundary>
      <DashboardContent />
    </ErrorBoundary>
  );
}
