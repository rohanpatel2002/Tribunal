'use client';
import { useState, useEffect } from 'react';
import { RotateCcw, Check, Copy, Key } from 'lucide-react';
import { fetchAPIKeys, rotateAPIKey, type ApiKeyInfo } from '@/lib/api';

const CARD = 'bg-white/60 backdrop-blur-md rounded-[28px] shadow-sm overflow-hidden';

function ago(iso?: string) {
  if (!iso) return '—';
  const m = Math.floor((Date.now() - new Date(iso).getTime()) / 60000);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  return h < 24 ? `${h}h ago` : `${Math.floor(h / 24)}d ago`;
}

function CopyBtn({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  function copy() { navigator.clipboard.writeText(text).then(() => { setCopied(true); setTimeout(() => setCopied(false), 2000); }); }
  return (
    <button onClick={copy} className="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors">
      {copied ? <Check size={13} className="text-emerald-500" /> : <Copy size={13} />}
    </button>
  );
}

export default function APIKeysTab({ apiKey, repository }: { apiKey: string; repository: string }) {
  const [keys, setKeys] = useState<ApiKeyInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [rotatingId, setRotatingId] = useState<string | null>(null);
  const [newKeyNotice, setNewKeyNotice] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    const k = await fetchAPIKeys(repository, apiKey);
    setKeys(k ?? []); setLoading(false);
  }

  useEffect(() => { load(); }, [apiKey, repository]); // eslint-disable-line

  async function handleRotate(k: ApiKeyInfo) {
    setRotatingId(k.id);
    const res = await rotateAPIKey(k.id, k.keyName, apiKey);
    if (res?.newKey) setNewKeyNotice(res.newKey);
    await load();
    setRotatingId(null);
  }

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-[38px] font-light tracking-tight text-slate-900">API Keys</h1>
        <p className="text-sm text-slate-500 mt-1">Manage credentials for <span className="font-mono">{repository}</span></p>
      </div>

      {/* Active key banner */}
      <div className="bg-[#292b2a] rounded-[28px] p-5 flex items-center justify-between">
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-widest text-white/40 mb-1.5">Current API Key (env)</p>
          <div className="flex items-center gap-2">
            <Key size={14} className="text-white/50" />
            <code className="text-white font-mono text-sm">{apiKey.slice(0, 10)}{'•'.repeat(Math.max(0, apiKey.length - 10))}</code>
            <CopyBtn text={apiKey} />
          </div>
        </div>
        <div className="flex items-center gap-1.5 text-emerald-400 text-xs font-medium">
          <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
          Active
        </div>
      </div>

      {newKeyNotice && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-[20px] p-4 flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-semibold text-emerald-800 mb-1">✅ New key generated — copy it now, it won't be shown again</p>
            <div className="flex items-center gap-2">
              <code className="text-xs font-mono text-emerald-700">{newKeyNotice}</code>
              <CopyBtn text={newKeyNotice} />
            </div>
          </div>
          <button onClick={() => setNewKeyNotice(null)} className="text-emerald-500 text-xs shrink-0 hover:text-emerald-700">Dismiss</button>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-7 h-7 rounded-full border-2 border-[#292b2a]/20 border-t-[#292b2a] animate-spin" />
        </div>
      ) : keys.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 bg-white/40 rounded-[28px]">
          <Key className="h-8 w-8 text-slate-300 mb-3" />
          <p className="text-slate-400 text-sm">No API keys found for this repository.</p>
          <p className="text-slate-400 text-xs mt-1">Keys are created server-side via environment variable.</p>
        </div>
      ) : (
        <div className={CARD}>
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-100">
                {['Key Name', 'Created', 'Last Used', 'Expires', 'Status', ''].map(h => (
                  <th key={h} className="px-6 py-4 text-left text-[10px] font-semibold uppercase tracking-widest text-slate-400">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {keys.map(k => (
                <tr key={k.id} className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60 transition-colors">
                  <td className="px-6 py-4 font-medium text-slate-800 text-sm">{k.keyName}</td>
                  <td className="px-6 py-4 text-sm text-slate-500">{ago(k.createdAt)}</td>
                  <td className="px-6 py-4 text-sm text-slate-500">{ago(k.lastUsedAt)}</td>
                  <td className="px-6 py-4 text-sm text-slate-500">{k.expiresAt ? ago(k.expiresAt) : 'Never'}</td>
                  <td className="px-6 py-4">
                    <div className={`flex items-center gap-1.5 text-xs font-medium ${k.isActive ? 'text-emerald-600' : 'text-slate-400'}`}>
                      <div className={`w-1.5 h-1.5 rounded-full ${k.isActive ? 'bg-emerald-400' : 'bg-slate-300'}`} />
                      {k.isActive ? 'Active' : 'Inactive'}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button onClick={() => handleRotate(k)} disabled={rotatingId === k.id}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-slate-200 bg-white text-xs font-medium text-slate-600 hover:border-slate-300 transition-colors disabled:opacity-50">
                      <RotateCcw size={11} className={rotatingId === k.id ? 'animate-spin' : ''} />
                      Rotate
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
