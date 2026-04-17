"use client";

import { useEffect, useState } from "react";
import { Activity, ShieldAlert, FileText, CheckCircle, TrendingUp, AlertTriangle } from "lucide-react";
import {
  fetchAuditSummary,
  fetchAuditLogs,
  getDemoAuditSummary,
  getDemoPRAnalysisRecords,
  type AuditSummary,
  type PRAnalysisRecord,
} from "@/lib/api";

export default function AnalyticsDashboard() {
  const [summary, setSummary] = useState<AuditSummary | null>(null);
  const [logs, setLogs] = useState<PRAnalysisRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [usingDemo, setUsingDemo] = useState(false);

  const REPOSITORY = "rohanpatel2002/tribunal";
  const apiKey = process.env.NEXT_PUBLIC_API_KEY ?? "dev_enterprise_key_123";

  useEffect(() => {
    let isActive = true;

    async function fetchAnalytics() {
      setLoading(true);
      setUsingDemo(false);

      try {
        const [summaryRes, logsRes] = await Promise.all([
          fetchAuditSummary(REPOSITORY, apiKey),
          fetchAuditLogs(REPOSITORY, apiKey, { limit: 10 }),
        ]);

        if (!isActive) return;

        if (summaryRes) {
          setSummary(summaryRes);
        } else {
          setSummary(getDemoAuditSummary(REPOSITORY));
          setUsingDemo(true);
        }

        if (logsRes) {
          setLogs(logsRes);
        } else {
          setLogs(getDemoPRAnalysisRecords(REPOSITORY));
          setUsingDemo(true);
        }
      } catch (err) {
        console.error("Failed to fetch analytics:", err);
        if (!isActive) return;
        setSummary(getDemoAuditSummary(REPOSITORY));
        setLogs(getDemoPRAnalysisRecords(REPOSITORY));
        setUsingDemo(true);
      } finally {
        if (isActive) setLoading(false);
      }
    }

    fetchAnalytics();

    return () => {
      isActive = false;
    };
  }, [apiKey]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-zinc-950 text-white">
        <Activity className="h-8 w-8 animate-spin text-emerald-500" />
        <span className="ml-3 text-lg font-medium">Loading Security Analytics...</span>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-white mb-1">Enterprise Analytics</h1>
            <p className="text-zinc-400">Security overview for <span className="text-emerald-400 font-mono">{REPOSITORY}</span></p>
          </div>
          <div className="flex items-center gap-3">
            {usingDemo && (
              <span className="text-xs font-semibold uppercase tracking-wide px-3 py-1 rounded-full bg-amber-500/10 text-amber-300 border border-amber-500/30">
                Demo data
              </span>
            )}
            <button className="bg-zinc-900 border border-zinc-800 hover:border-zinc-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center">
              <CheckCircle className="h-4 w-4 mr-2 text-emerald-500" />
              System Healthy
            </button>
          </div>
        </header>

        {/* KPI Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-zinc-900/50 border border-zinc-800/50 rounded-xl p-6 shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <p className="text-sm font-medium text-zinc-400">Total Scanned PRs</p>
              <div className="p-2 bg-blue-500/10 rounded-lg"><FileText className="h-4 w-4 text-blue-500" /></div>
            </div>
            <h3 className="text-3xl font-bold">{summary?.totalPRs || 0}</h3>
          </div>
          
          <div className="bg-zinc-900/50 border border-zinc-800/50 rounded-xl p-6 shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <p className="text-sm font-medium text-zinc-400">AI-Generated Code</p>
              <div className="p-2 bg-emerald-500/10 rounded-lg"><Activity className="h-4 w-4 text-emerald-500" /></div>
            </div>
            <h3 className="text-3xl font-bold">{summary?.aiGeneratedPRs || 0}</h3>
            <p className="text-xs tracking-wide text-zinc-500 mt-2">PRs containing raw AI output</p>
          </div>

          <div className="bg-zinc-900/50 border border-zinc-800/50 rounded-xl p-6 shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <p className="text-sm font-medium text-zinc-400">Critical Threats</p>
              <div className="p-2 bg-red-500/10 rounded-lg"><ShieldAlert className="h-4 w-4 text-red-500" /></div>
            </div>
            <h3 className="text-3xl font-bold text-red-400">{summary?.criticalRisks || 0}</h3>
            <p className="text-xs tracking-wide text-red-500/70 mt-2">Requires immediate attention</p>
          </div>

          <div className="bg-zinc-900/50 border border-zinc-800/50 rounded-xl p-6 shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <p className="text-sm font-medium text-zinc-400">Average AI Score</p>
              <div className="p-2 bg-amber-500/10 rounded-lg"><TrendingUp className="h-4 w-4 text-amber-500" /></div>
            </div>
            <h3 className="text-3xl font-bold">{((summary?.averageAIScore || 0) * 100).toFixed(1)}%</h3>
            <p className="text-xs tracking-wide text-zinc-500 mt-2">Likelihood of syntethic patches</p>
          </div>
        </div>

        {/* Audit Log Table */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg">
          <div className="border-b border-zinc-800 p-6 flex justify-between items-center">
            <h2 className="text-lg font-semibold flex items-center">
              <AlertTriangle className="h-5 w-5 mr-3 text-zinc-400" />
              Recent Audit Logs
            </h2>
            <span className="text-xs font-medium px-2 py-1 bg-zinc-800 text-zinc-300 rounded-md">Last 10 Scans</span>
          </div>
          
          {logs.length === 0 ? (
            <div className="p-12 text-center text-zinc-500 flex flex-col items-center justify-center">
              <ShieldAlert className="h-10 w-10 mb-4 opacity-50" />
              <p>No recent pull request scans found.</p>
              <p className="text-sm mt-1">Submit a webhook payload to generate analytics.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-zinc-400 uppercase bg-zinc-900/50 border-b border-zinc-800">
                  <tr>
                    <th className="px-6 py-4 font-medium">PR ID</th>
                    <th className="px-6 py-4 font-medium">Recommendation</th>
                    <th className="px-6 py-4 font-medium">Files</th>
                    <th className="px-6 py-4 font-medium">AI Generated</th>
                    <th className="px-6 py-4 font-medium text-red-400">Critical</th>
                    <th className="px-6 py-4 font-medium text-amber-400">High</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800">
                  {logs.map((log) => (
                    <tr key={log.id} className="hover:bg-zinc-800/20 transition-colors">
                      <td className="px-6 py-4 font-medium text-emerald-400">#{" "}{log.prNumber}</td>
                      <td className="px-6 py-4">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-semibold
                          ${log.recommendation === 'APPROVE' ? 'bg-emerald-500/10 text-emerald-400' : 
                            log.recommendation === 'BLOCK' ? 'bg-red-500/10 text-red-400' : 
                            'bg-amber-500/10 text-amber-400'}`}>
                          {log.recommendation}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-zinc-300">{log.totalFiles}</td>
                      <td className="px-6 py-4 text-zinc-300">{log.aiGenerated}</td>
                      <td className="px-6 py-4 font-medium text-red-500">{log.critical}</td>
                      <td className="px-6 py-4 font-medium text-amber-500">{log.high}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
        
      </div>
    </div>
  );
}
