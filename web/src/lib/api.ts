const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

type FetchOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

class ApiClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private householdId: string | null = null;

  constructor() {
    if (typeof window !== 'undefined') {
      this.accessToken = localStorage.getItem('access_token');
      this.refreshToken = localStorage.getItem('refresh_token');
      this.householdId = localStorage.getItem('household_id');
    }
  }

  setTokens(access: string, refresh?: string) {
    this.accessToken = access;
    localStorage.setItem('access_token', access);
    if (refresh) {
      this.refreshToken = refresh;
      localStorage.setItem('refresh_token', refresh);
    }
  }

  setHouseholdId(id: string) {
    this.householdId = id;
    localStorage.setItem('household_id', id);
  }

  getHouseholdId(): string | null {
    return this.householdId;
  }

  clearTokens() {
    this.accessToken = null;
    this.refreshToken = null;
    this.householdId = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('household_id');
  }

  isAuthenticated(): boolean {
    return !!this.accessToken;
  }

  private async request<T>(path: string, opts: FetchOptions = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...opts.headers,
    };

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }
    if (this.householdId) {
      headers['X-Household-ID'] = this.householdId;
    }

    const res = await fetch(`${API_URL}${path}`, {
      method: opts.method || 'GET',
      headers,
      body: opts.body ? JSON.stringify(opts.body) : undefined,
    });

    // Try refresh on 401
    if (res.status === 401 && this.refreshToken) {
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        // Retry with new token
        headers['Authorization'] = `Bearer ${this.accessToken}`;
        const retry = await fetch(`${API_URL}${path}`, {
          method: opts.method || 'GET',
          headers,
          body: opts.body ? JSON.stringify(opts.body) : undefined,
        });
        if (!retry.ok) {
          const err = await retry.json().catch(() => ({ error: 'Request failed' }));
          throw new Error(err.error || `HTTP ${retry.status}`);
        }
        return retry.json();
      }
      this.clearTokens();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
      throw new Error('Session expired');
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }

    // Handle empty responses
    const text = await res.text();
    return text ? JSON.parse(text) : ({} as T);
  }

  private async tryRefresh(): Promise<boolean> {
    try {
      const res = await fetch(`${API_URL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });
      if (!res.ok) return false;
      const data = await res.json();
      this.setTokens(data.access_token, data.refresh_token);
      return true;
    } catch {
      return false;
    }
  }

  // ----- Auth -----
  register(body: { email: string; password: string; name: string }) {
    return this.request<import('../types').AuthResponse>('/auth/register', {
      method: 'POST',
      body,
    });
  }

  login(body: { email: string; password: string }) {
    return this.request<import('../types').AuthResponse>('/auth/login', {
      method: 'POST',
      body,
    });
  }

  logout() {
    return this.request<void>('/auth/logout', { method: 'POST' });
  }

  // ----- Households -----
  listHouseholds() {
    return this.request<import('../types').Household[]>('/api/households');
  }

  createHousehold(body: { name: string }) {
    return this.request<import('../types').Household>('/api/households', {
      method: 'POST',
      body,
    });
  }

  listMembers(householdId: string) {
    return this.request<import('../types').HouseholdMember[]>(
      `/api/households/${householdId}/members`
    );
  }

  invite(householdId: string, email: string) {
    return this.request<import('../types').Invitation>(
      `/api/households/${householdId}/invite`,
      { method: 'POST', body: { email } }
    );
  }

  acceptInvitation(token: string) {
    return this.request<{ message: string }>(
      `/api/invitations/${token}/accept`,
      { method: 'POST' }
    );
  }

  removeMember(householdId: string, userId: string) {
    return this.request<{ message: string }>(
      `/api/households/${householdId}/members/${userId}`,
      { method: 'DELETE' }
    );
  }

  // ----- Accounts -----
  listAccounts() {
    return this.request<import('../types').Account[]>('/api/accounts');
  }

  createAccount(body: import('../types').CreateAccountRequest) {
    return this.request<import('../types').Account>('/api/accounts', {
      method: 'POST',
      body,
    });
  }

  getAccount(id: string) {
    return this.request<import('../types').Account>(`/api/accounts/${id}`);
  }

  updateAccount(id: string, body: import('../types').UpdateAccountRequest) {
    return this.request<import('../types').Account>(`/api/accounts/${id}`, {
      method: 'PUT',
      body,
    });
  }

  deleteAccount(id: string) {
    return this.request<{ message: string }>(`/api/accounts/${id}`, {
      method: 'DELETE',
    });
  }

  // ----- Transactions -----
  listTransactions(params?: Record<string, string>) {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return this.request<import('../types').PaginatedResponse<import('../types').Transaction>>(
      `/api/transactions${qs}`
    );
  }

  createTransaction(body: import('../types').CreateTransactionRequest) {
    return this.request<import('../types').Transaction>('/api/transactions', {
      method: 'POST',
      body,
    });
  }

  getTransaction(id: string) {
    return this.request<import('../types').Transaction>(`/api/transactions/${id}`);
  }

  updateTransaction(id: string, body: import('../types').UpdateTransactionRequest) {
    return this.request<import('../types').Transaction>(`/api/transactions/${id}`, {
      method: 'PUT',
      body,
    });
  }

  deleteTransaction(id: string) {
    return this.request<{ message: string }>(`/api/transactions/${id}`, {
      method: 'DELETE',
    });
  }

  // ----- Export -----
  async exportCSV(from?: string, to?: string) {
    const params = new URLSearchParams();
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    const qs = params.toString() ? `?${params.toString()}` : '';

    const headers: Record<string, string> = {};
    if (this.accessToken) headers['Authorization'] = `Bearer ${this.accessToken}`;
    if (this.householdId) headers['X-Household-ID'] = this.householdId;

    const res = await fetch(`${API_URL}/api/export/csv${qs}`, { headers });
    if (!res.ok) throw new Error('Export failed');

    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = res.headers.get('Content-Disposition')?.split('filename=')[1]?.replace(/"/g, '') || 'export.csv';
    a.click();
    window.URL.revokeObjectURL(url);
  }
}

export const api = new ApiClient();
