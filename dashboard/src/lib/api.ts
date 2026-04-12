/**
 * API Integration Utilities
 * Handles communication with Tribunal backend
 * Supports the new flat AnalyzeResponse model structure from Phase 1
 */

// ============= TYPE DEFINITIONS =============

export interface FileAnalysis {
  filename: string;
  riskLevel: string;
  aiScore: number;
  isAIGenerated: boolean;
  summary: string;
  suggestedFix?: string;
}

/**
 * New flat AnalyzeResponse structure (Phase 1 breaking change)
 * Old structure: { summary: {}, results: [] }
 * New structure: flat fields + Files array
 */
export interface AnalyzeResponse {
  recommendation: string;
  totalFiles: number;
  aiGeneratedCount: number;
  criticalCount: number;
  highCount: number;
  mediumCount: number;
  lowCount: number;
  files: FileAnalysis[];
}

export interface AuditSummary {
  repository: string;
  totalPRs: number;
  totalFiles: number;
  aiGeneratedPRs: number;
  criticalRisks: number;
  highRisks: number;
  averageAIScore: number;
}

export interface PRAnalysisRecord {
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
  createdAt?: string;
}

export interface SecurityPolicy {
  id: string;
  repository: string;
  policyName: string;
  policyType: 'SEVERITY_THRESHOLD' | 'AI_DETECTION' | 'CUSTOM';
  rules: Record<string, unknown>;
  isActive: boolean;
  createdBy: string;
  createdAt: string;
}

export interface ApiKeyInfo {
  id: string;
  keyName: string;
  repository: string;
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
  isActive: boolean;
  daysUntilExpiry?: number;
}

export interface PaginationParams {
  offset?: number;
  limit?: number;
}

export interface FilterParams {
  severity?: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  startDate?: string; // ISO 8601
  endDate?: string;   // ISO 8601
  isAIGenerated?: boolean;
}

// ============= API CLIENT =============

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_VERSION = 'v1';

/**
 * Get authorization header with API key
 */
function getAuthHeader(apiKey: string): HeadersInit {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${apiKey}`,
  };
}

/**
 * Fetch audit summary for a repository
 */
export async function fetchAuditSummary(
  repository: string,
  apiKey: string
): Promise<AuditSummary | null> {
  try {
    const response = await fetch(
      `${API_BASE}/api/${API_VERSION}/audit/summary?repository=${encodeURIComponent(repository)}`,
      {
        method: 'GET',
        headers: getAuthHeader(apiKey),
      }
    );

    if (!response.ok) {
      console.warn(`Audit summary fetch failed: ${response.status}`);
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching audit summary:', error);
    return null;
  }
}

/**
 * Fetch PR analysis logs with optional pagination and filtering
 */
export async function fetchAuditLogs(
  repository: string,
  apiKey: string,
  pagination?: PaginationParams,
  filters?: FilterParams
): Promise<PRAnalysisRecord[] | null> {
  try {
    const params = new URLSearchParams({
      repository,
      limit: String(pagination?.limit ?? 10),
      ...(pagination?.offset && { offset: String(pagination.offset) }),
    });

    if (filters?.severity) params.append('severity', filters.severity);
    if (filters?.startDate) params.append('startDate', filters.startDate);
    if (filters?.endDate) params.append('endDate', filters.endDate);
    if (filters?.isAIGenerated !== undefined) {
      params.append('isAIGenerated', String(filters.isAIGenerated));
    }

    const response = await fetch(
      `${API_BASE}/api/${API_VERSION}/audit/logs?${params.toString()}`,
      {
        method: 'GET',
        headers: getAuthHeader(apiKey),
      }
    );

    if (!response.ok) {
      console.warn(`Audit logs fetch failed: ${response.status}`);
      return null;
    }

    const data = await response.json();
    return data.data || [];
  } catch (error) {
    console.error('Error fetching audit logs:', error);
    return null;
  }
}

/**
 * Fetch security policies for a repository
 */
export async function fetchSecurityPolicies(
  repository: string,
  apiKey: string
): Promise<SecurityPolicy[] | null> {
  try {
    const response = await fetch(
      `${API_BASE}/api/${API_VERSION}/policies?repository=${encodeURIComponent(repository)}`,
      {
        method: 'GET',
        headers: getAuthHeader(apiKey),
      }
    );

    if (!response.ok) {
      console.warn(`Security policies fetch failed: ${response.status}`);
      return null;
    }

    const data = await response.json();
    return data.policies || [];
  } catch (error) {
    console.error('Error fetching security policies:', error);
    return null;
  }
}

/**
 * Create a new security policy
 */
export async function createSecurityPolicy(
  repository: string,
  policy: Omit<SecurityPolicy, 'id' | 'createdAt'>,
  apiKey: string
): Promise<SecurityPolicy | null> {
  try {
    const response = await fetch(
      `${API_BASE}/api/${API_VERSION}/policies?repository=${encodeURIComponent(repository)}`,
      {
        method: 'POST',
        headers: getAuthHeader(apiKey),
        body: JSON.stringify(policy),
      }
    );

    if (!response.ok) {
      const error = await response.text();
      console.warn(`Create policy failed: ${response.status} - ${error}`);
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Error creating security policy:', error);
    return null;
  }
}

/**
 * Delete a security policy
 */
export async function deleteSecurityPolicy(
  repository: string,
  policyName: string,
  apiKey: string
): Promise<boolean> {
  try {
    const response = await fetch(
      `${API_BASE}/api/${API_VERSION}/policies?repository=${encodeURIComponent(repository)}&policyName=${encodeURIComponent(policyName)}`,
      {
        method: 'DELETE',
        headers: getAuthHeader(apiKey),
      }
    );

    if (!response.ok) {
      console.warn(`Delete policy failed: ${response.status}`);
      return false;
    }

    return true;
  } catch (error) {
    console.error('Error deleting security policy:', error);
    return false;
  }
}

/**
 * Check API connectivity and authentication
 */
export async function checkHealthStatus(apiKey: string): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE}/health/detailed`, {
      method: 'GET',
      headers: getAuthHeader(apiKey),
    });

    return response.ok;
  } catch (error) {
    console.warn('Health check failed:', error);
    return false;
  }
}

/**
 * Format API error response for display
 */
export function formatApiError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'string') {
    return error;
  }
  return 'An unexpected error occurred';
}

/**
 * Get demo data for offline mode (development/testing)
 */
export function getDemoAuditSummary(repository: string): AuditSummary {
  return {
    repository,
    totalPRs: 24,
    totalFiles: 145,
    aiGeneratedPRs: 8,
    criticalRisks: 0,
    highRisks: 2,
    averageAIScore: 0.12,
  };
}

export function getDemoPRAnalysisRecords(repository: string): PRAnalysisRecord[] {
  return [
    {
      id: 'demo-1',
      repository,
      prNumber: 101,
      recommendation: 'APPROVE',
      totalFiles: 3,
      aiGenerated: 0,
      critical: 0,
      high: 0,
      medium: 0,
      low: 1,
      createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    },
    {
      id: 'demo-2',
      repository,
      prNumber: 99,
      recommendation: 'REVIEW_REQUIRED',
      totalFiles: 12,
      aiGenerated: 6,
      critical: 0,
      high: 2,
      medium: 4,
      low: 2,
      createdAt: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(),
    },
    {
      id: 'demo-3',
      repository,
      prNumber: 97,
      recommendation: 'APPROVE',
      totalFiles: 5,
      aiGenerated: 1,
      critical: 0,
      high: 0,
      medium: 1,
      low: 2,
      createdAt: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString(),
    },
  ];
}
