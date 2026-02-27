// Domain types matching Go backend models

export type AccountType = 'card' | 'deposit' | 'cash';
export type TransactionType = 'income' | 'expense' | 'transfer';
export type HouseholdRole = 'owner' | 'member';
export type InvitationStatus = 'pending' | 'accepted' | 'expired';

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface Household {
  id: string;
  name: string;
  owner_id: string;
  created_at: string;
}

export interface HouseholdMember {
  household_id: string;
  user_id: string;
  role: HouseholdRole;
  joined_at: string;
  email: string;
  user_name: string;
}

export interface Invitation {
  id: string;
  household_id: string;
  email: string;
  invited_by: string;
  token: string;
  status: InvitationStatus;
  expires_at: string;
  created_at: string;
}

export interface Account {
  id: string;
  household_id: string;
  name: string;
  type: AccountType;
  balance: string;
  currency: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  household_id: string;
  type: TransactionType;
  description: string;
  amount: string;
  account_id: string;
  destination_account_id?: string;
  tags: string[];
  note?: string;
  transacted_at: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token?: string;
  user: User;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// Request types
export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface CreateAccountRequest {
  name: string;
  type: AccountType;
  balance: string;
  currency: string;
}

export interface UpdateAccountRequest {
  name?: string;
  type?: AccountType;
  currency?: string;
}

export interface CreateTransactionRequest {
  type: TransactionType;
  description: string;
  amount: string;
  account_id: string;
  destination_account_id?: string;
  tags: string[];
  note?: string;
  transacted_at: string;
}

export interface UpdateTransactionRequest extends CreateTransactionRequest {}

export interface CreateHouseholdRequest {
  name: string;
}

export interface InviteRequest {
  email: string;
}
