import { PRAnalysisRecord } from './api';

/**
 * Export audit logs to CSV format
 */
export function exportToCSV(logs: PRAnalysisRecord[], filename: string = 'audit-logs.csv'): void {
  if (!logs || logs.length === 0) {
    console.warn('No data to export');
    return;
  }

  // CSV Header
  const headers = [
    'PR Number',
    'Repository',
    'Recommendation',
    'Total Files',
    'AI Generated',
    'Critical',
    'High',
    'Medium',
    'Low',
  ];

  // CSV Rows
  const rows = logs.map((log) => [
    log.prNumber,
    log.repository,
    log.recommendation,
    log.totalFiles,
    log.aiGenerated,
    log.critical,
    log.high,
    log.medium,
    log.low,
  ]);

  // Build CSV string
  const csv = [
    headers.join(','),
    ...rows.map((row) => row.map((cell) => `"${cell}"`).join(',')),
  ].join('\n');

  // Create blob and download
  downloadFile(csv, filename, 'text/csv;charset=utf-8;');
}

/**
 * Export audit logs to JSON format
 */
export function exportToJSON(
  logs: PRAnalysisRecord[],
  metadata?: { repository?: string; exportDate?: string },
  filename: string = 'audit-logs.json'
): void {
  const data = {
    metadata: {
      exportDate: metadata?.exportDate || new Date().toISOString(),
      repository: metadata?.repository || 'All Repositories',
      totalRecords: logs.length,
    },
    data: logs,
  };

  const json = JSON.stringify(data, null, 2);
  downloadFile(json, filename, 'application/json;charset=utf-8;');
}

/**
 * Export audit logs to TSV (Tab-Separated Values)
 */
export function exportToTSV(logs: PRAnalysisRecord[], filename: string = 'audit-logs.tsv'): void {
  if (!logs || logs.length === 0) {
    console.warn('No data to export');
    return;
  }

  // TSV Header
  const headers = [
    'PR Number',
    'Repository',
    'Recommendation',
    'Total Files',
    'AI Generated',
    'Critical',
    'High',
    'Medium',
    'Low',
  ];

  // TSV Rows
  const rows = logs.map((log) => [
    log.prNumber,
    log.repository,
    log.recommendation,
    log.totalFiles,
    log.aiGenerated,
    log.critical,
    log.high,
    log.medium,
    log.low,
  ]);

  // Build TSV string
  const tsv = [headers.join('\t'), ...rows.map((row) => row.join('\t'))].join('\n');

  downloadFile(tsv, filename, 'text/tab-separated-values;charset=utf-8;');
}

/**
 * Generate a summary report in HTML format
 */
export function generateHTMLReport(
  logs: PRAnalysisRecord[],
  summary?: { totalPRs?: number; criticalRisks?: number; highRisks?: number; aiGeneratedPRs?: number }
): string {
  const totalRecords = summary?.totalPRs ?? logs.length;
  const criticalCount = summary?.criticalRisks ?? logs.reduce((sum, log) => sum + log.critical, 0);
  const highCount = summary?.highRisks ?? logs.reduce((sum, log) => sum + log.high, 0);
  const aiGeneratedCount = summary?.aiGeneratedPRs ?? logs.filter((log) => log.aiGenerated > 0).length;

  const html = `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>Tribunal Audit Report</title>
      <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #0A0A0A; color: #E2E8F0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #FFFFFF; border-bottom: 2px solid #4F46E5; padding-bottom: 10px; }
        .summary { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin: 20px 0; }
        .metric { background: #0F0F11; border: 1px solid #1F1F22; padding: 15px; border-radius: 8px; }
        .metric-value { font-size: 24px; font-weight: bold; color: #4F46E5; }
        .metric-label { font-size: 12px; color: #64748B; text-transform: uppercase; margin-top: 5px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th { background: #1A1A1E; text-align: left; padding: 10px; font-weight: 600; border-bottom: 1px solid #27272A; }
        td { padding: 10px; border-bottom: 1px solid #27272A; }
        tr:hover { background: #141416; }
        .critical { color: #EF4444; }
        .high { color: #F97316; }
        .medium { color: #EAB308; }
        .low { color: #22C55E; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #27272A; color: #64748B; font-size: 12px; }
      </style>
    </head>
    <body>
      <div class="container">
        <h1>🛡️ Tribunal Security Audit Report</h1>
        
        <div class="summary">
          <div class="metric">
            <div class="metric-value">${totalRecords}</div>
            <div class="metric-label">Total PRs Analyzed</div>
          </div>
          <div class="metric">
            <div class="metric-value critical">${criticalCount}</div>
            <div class="metric-label">Critical Risks</div>
          </div>
          <div class="metric">
            <div class="metric-value high">${highCount}</div>
            <div class="metric-label">High Risks</div>
          </div>
          <div class="metric">
            <div class="metric-value">${aiGeneratedCount}</div>
            <div class="metric-label">AI-Generated PRs</div>
          </div>
        </div>

        <h2>Audit Trail</h2>
        <table>
          <thead>
            <tr>
              <th>PR #</th>
              <th>Repository</th>
              <th>Status</th>
              <th>Files</th>
              <th>AI Gen</th>
              <th style="text-align: center;">Risk Breakdown</th>
            </tr>
          </thead>
          <tbody>
            ${logs
              .map(
                (log) => `
              <tr>
                <td>#${log.prNumber}</td>
                <td>${log.repository}</td>
                <td>${log.recommendation}</td>
                <td>${log.totalFiles}</td>
                <td>${log.aiGenerated}</td>
                <td>
                  <span class="critical">🔴 ${log.critical}</span> |
                  <span class="high">🟠 ${log.high}</span> |
                  <span class="medium">🟡 ${log.medium}</span> |
                  <span class="low">🟢 ${log.low}</span>
                </td>
              </tr>
            `
              )
              .join('')}
          </tbody>
        </table>

        <div class="footer">
          <p>Generated: ${new Date().toLocaleString()}</p>
          <p>Tribunal Security Audit System © 2026</p>
        </div>
      </div>
    </body>
    </html>
  `;

  return html;
}

/**
 * Download generated file to user's computer
 */
function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

/**
 * Generate timestamped filename
 */
export function getTimestampedFilename(baseFilename: string): string {
  const timestamp = new Date()
    .toISOString()
    .split('T')[0]; // YYYY-MM-DD
  const [name, ext] = baseFilename.split('.');
  return `${name}-${timestamp}.${ext}`;
}
