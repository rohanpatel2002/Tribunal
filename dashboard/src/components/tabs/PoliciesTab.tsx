'use client';
import { useState, useEffect } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { fetchSecurityPolicies, createSecurityPolicy, deleteSecurityPolicy, type SecurityPolicy } from '@/lib/api';

const CARD = 'bg-white/60 backdrop-blur-md rounded-[28px] shadow-sm overflow-hidden';
const INPUT = 'w-full px-4 py-2.5 rounded-xl border border-slate-200 bg-white text-sm text-slate-800 focus:outline-none focus:ring-2 focus:ring-[#292b2a]/20 placeholder:text-slate-400';
const TYPES = ['AI_DETECTION', 'VULNERABILITY_SCAN', 'CODE_STYLE', 'COMPLIANCE'] as const;
const THRESHOLDS = ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'] as const;

const TYPE_COLOR: Record<string, string> = {
  AI_DETECTION: 'bg-violet-50 text-violet-700',
  VULNERABILITY_SCAN: 'bg-red-50 text-red-700',
  CODE_STYLE: 'bg-blue-50 text-blue-700',
  COMPLIANCE: 'bg-amber-50 text-amber-700',
};

export default function PoliciesTab({ apiKey, repository }: { apiKey: string; repository: string }) {
  const [policies, setPolicies] = useState<SecurityPolicy[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [form, setForm] = useState({ policyName: '', policyType: 'AI_DETECTION', description: '', severityThreshold: 'HIGH' });

  useEffect(() => {
    fetchSecurityPolicies(repository, apiKey).then(p => { setPolicies(p ?? []); setLoading(false); });
  }, [apiKey, repository]);

  async function handleCreate() {
    if (!form.policyName.trim()) return;
    setSaving(true);
    const created = await createSecurityPolicy(repository, {
      repository, policyName: form.policyName, policyType: form.policyType as SecurityPolicy['policyType'],
      description: form.description, rules: {}, isActive: true, createdBy: 'dashboard',
    }, apiKey);
    if (created) { setPolicies(p => [created, ...p]); setShowForm(false); setForm({ policyName: '', policyType: 'AI_DETECTION', description: '', severityThreshold: 'HIGH' }); }
    setSaving(false);
  }

  async function handleDelete(name: string, id: string) {
    setDeletingId(id);
    await deleteSecurityPolicy(repository, name, apiKey);
    setPolicies(p => p.filter(x => x.id !== id));
    setDeletingId(null);
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-[38px] font-light tracking-tight text-slate-900">Policies</h1>
          <p className="text-sm text-slate-500 mt-1">Security enforcement rules for <span className="font-mono">{repository}</span></p>
        </div>
        <button onClick={() => setShowForm(s => !s)}
          className="flex items-center gap-2 px-5 py-2.5 bg-[#292b2a] text-white rounded-full text-sm font-medium hover:bg-black transition-colors">
          <Plus size={14} /> {showForm ? 'Cancel' : 'New Policy'}
        </button>
      </div>

      {showForm && (
        <div className="bg-white/70 backdrop-blur-md rounded-[28px] p-6 shadow-sm space-y-4">
          <h3 className="font-semibold text-slate-800">Create Policy</h3>
          <div className="grid grid-cols-2 gap-4">
            <input className={INPUT} placeholder="Policy name *" value={form.policyName} onChange={e => setForm(f => ({ ...f, policyName: e.target.value }))} />
            <select className={INPUT} value={form.policyType} onChange={e => setForm(f => ({ ...f, policyType: e.target.value }))}>
              {TYPES.map(t => <option key={t} value={t}>{t.replace(/_/g, ' ')}</option>)}
            </select>
            <select className={INPUT} value={form.severityThreshold} onChange={e => setForm(f => ({ ...f, severityThreshold: e.target.value }))}>
              {THRESHOLDS.map(t => <option key={t}>{t}</option>)}
            </select>
            <input className={INPUT} placeholder="Description (optional)" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} />
          </div>
          <button onClick={handleCreate} disabled={saving || !form.policyName.trim()}
            className="px-6 py-2.5 bg-[#292b2a] text-white rounded-full text-sm font-medium disabled:opacity-40 hover:bg-black transition-colors">
            {saving ? 'Creating…' : 'Create Policy'}
          </button>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20"><div className="w-7 h-7 rounded-full border-2 border-[#292b2a]/20 border-t-[#292b2a] animate-spin" /></div>
      ) : policies.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 bg-white/40 rounded-[28px]">
          <p className="text-slate-400 text-sm">No policies yet. Create your first enforcement rule above.</p>
        </div>
      ) : (
        <div className={CARD}>
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-100">
                {['Policy Name', 'Type', 'Severity', 'Status', ''].map(h => (
                  <th key={h} className="px-6 py-4 text-left text-[10px] font-semibold uppercase tracking-widest text-slate-400">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {policies.map(p => (
                <tr key={p.id} className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60 transition-colors">
                  <td className="px-6 py-4 font-medium text-slate-800 text-sm">{p.policyName}</td>
                  <td className="px-6 py-4">
                    <span className={`text-[11px] font-semibold px-2.5 py-1 rounded-full ${TYPE_COLOR[p.policyType] ?? 'bg-slate-100 text-slate-600'}`}>
                      {p.policyType.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-slate-500">
                    {(p as unknown as { severityThreshold?: string }).severityThreshold ?? '—'}
                  </td>
                  <td className="px-6 py-4">
                    <div className={`flex items-center gap-1.5 text-xs font-medium ${p.isActive ? 'text-emerald-600' : 'text-slate-400'}`}>
                      <div className={`w-1.5 h-1.5 rounded-full ${p.isActive ? 'bg-emerald-400' : 'bg-slate-300'}`} />
                      {p.isActive ? 'Active' : 'Inactive'}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button onClick={() => handleDelete(p.policyName, p.id)} disabled={deletingId === p.id}
                      className="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-40">
                      <Trash2 size={14} />
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
