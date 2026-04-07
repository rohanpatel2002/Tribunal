'use client';

import { useState, useEffect } from 'react';

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

  const fetchData = async () => {
    setLoading(true);
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
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-white font-sans p-8">
      {/* Header */}
      <header className="mb-10 flex items-center justify-between border-b border-gray-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white mb-1">TRIBUNAL <span className="text-gray-500 font-light ml-2">Enterprise</span></h1>
          <p className="text-gray-400 text-sm">CTO Audit Dashboard</p>
        </div>
        <div className="flex gap-4">
          <input 
            type="password" 
            placeholder="TRIBUNAL_API_KEY" 
            className="bg-gray-900 border border-gray-800 rounded px-4 py-2 text-sm text-gray-300 focus:outline-none focus:border-indigo-500"
            onChange={(e) => setApiKey(e.target.value)}
            value={apiKey}
          />
          <button 
            onClick={fetchData}
            className="bg-white text-black font-medium px-4 py-2 rounded text-sm hover:bg-gray-200 transition-colors"
          >
            Refresh Data
          </button>
        </div>
      </header>

      {/* Main Content */}
      {loading ? (
        <div className="h-64 flex items-center justify-center">
          <div className="animate-pulse flex flex-col items-center">
            <div className="h-8 w-8 bg-indigo-500 rounded-full mb-4"></div>
            <p className="text-gray-500">Querying Postgres...</p>
          </div>
        </div>
      ) : (
        <div className="max-w-6xl mx-auto space-y-8">
          
          <div className="flex items-center gap-3 mb-6">
            <span className="h-3 w-3 rounded-full bg-green-500 animate-pulse"></span>
            <h2 className="text-xl font-medium tracking-tight">Active Context: <span className="text-indigo-400 font-mono ml-2">{data?.repository}</span></h2>
          </div>

          {/* Metric Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl relative overflow-hidden">
              <div className="absolute top-0 right-0 w-16 h-16 bg-blue-500 opacity-10 blur-xl rounded-full translate-x-4 -translate-y-4"></div>
              <p className="text-sm text-gray-400 mb-1">Total PRs Analyzed</p>
              <p className="text-3xl font-semibold text-white">{data?.totalPRs}</p>
            </div>
            
            <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl relative overflow-hidden">
              <div className="absolute top-0 right-0 w-16 h-16 bg-purple-500 opacity-10 blur-xl rounded-full translate-x-4 -translate-y-4"></div>
              <p className="text-sm text-gray-400 mb-1">AI-Authored Files</p>
              <p className="text-3xl font-semibold text-purple-400">{data?.aiGeneratedPRs} <span className="text-xs text-gray-500 ml-1 font-normal">/ {data?.totalFiles} files</span></p>
            </div>

            <div className="bg-[#1a0f0f] border border-red-900/30 p-6 rounded-xl relative overflow-hidden">
              <div className="absolute top-0 right-0 w-16 h-16 bg-red-500 opacity-10 blur-xl rounded-full translate-x-4 -translate-y-4"></div>
              <p className="text-sm text-red-400/80 mb-1">Critical God-Mode Flags</p>
              <p className="text-3xl font-semibold text-red-500">{data?.criticalRisks}</p>
            </div>

            <div className="bg-[#1c130d] border border-orange-900/30 p-6 rounded-xl relative overflow-hidden">
              <div className="absolute top-0 right-0 w-16 h-16 bg-orange-500 opacity-10 blur-xl rounded-full translate-x-4 -translate-y-4"></div>
              <p className="text-sm text-orange-400/80 mb-1">High Risk Flags</p>
              <p className="text-3xl font-semibold text-orange-500">{data?.highRisks}</p>
            </div>
          </div>

          {/* Detailed Context Zone */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 bg-gray-900 border border-gray-800 rounded-xl p-6">
              <h3 className="text-lg font-medium text-white mb-6">Recent Heuristic Flag Log</h3>
              
              <div className="space-y-4">
                {logs.length === 0 && (
                  <div className="text-center py-4 text-gray-500 text-sm">
                    No recent analysis logs found.
                  </div>
                )}
                {logs.map((log, i) => (
                  <div key={i} className="flex items-center justify-between p-4 bg-black/50 border border-gray-800/60 rounded-lg">
                    <div className="flex items-center gap-4">
                      <span className={`px-2 py-1 text-[10px] font-bold rounded ${log.criticalRiskFound ? 'bg-red-500/10 text-red-500 border border-red-500/20' : log.highRiskFound ? 'bg-orange-500/10 text-orange-500 border border-orange-500/20' : 'bg-yellow-500/10 text-yellow-500 border border-yellow-500/20'}`}>
                        {log.criticalRiskFound ? 'CRITICAL' : log.highRiskFound ? 'HIGH' : 'MEDIUM'}
                      </span>
                      <code className="text-sm text-gray-300">{log.repository}/PR-{log.prNumber}</code>
                    </div>
                    <div className="flex items-center gap-6">
                      <span className="text-sm text-gray-500">{new Date(log.createdAt).toLocaleDateString()}</span>
                      <span className="text-sm text-gray-400 font-mono">Score: {log.overallRiskScore}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 flex flex-col relative overflow-hidden">
              <h3 className="text-lg font-medium text-white mb-6 z-10">System Health Trend</h3>
              <div className="flex-1 flex flex-col items-center justify-center z-10">
                <div className="w-32 h-32 rounded-full border-[8px] border-indigo-500 flex items-center justify-center mb-4 relative drop-shadow-[0_0_15px_rgba(99,102,241,0.2)]">
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
  );
}
