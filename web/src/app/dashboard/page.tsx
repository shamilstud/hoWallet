'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { api } from '../../lib/api';
import type { Account, Transaction } from '../../types';

export default function DashboardPage() {
  const router = useRouter();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [recentTxns, setRecentTxns] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!api.isAuthenticated()) {
      router.replace('/login');
      return;
    }
    loadData();
  }, [router]);

  const loadData = async () => {
    try {
      const [accs, txns] = await Promise.all([
        api.listAccounts(),
        api.listTransactions({ limit: '5' }),
      ]);
      setAccounts(accs);
      setRecentTxns(txns.data);
    } catch {
      // If household not set, data will be empty
    } finally {
      setLoading(false);
    }
  };

  const totalBalance = accounts.reduce(
    (sum, acc) => sum + parseFloat(acc.balance),
    0
  );

  const totalIncome = recentTxns
    .filter((t) => t.type === 'income')
    .reduce((sum, t) => sum + parseFloat(t.amount), 0);

  const totalExpense = recentTxns
    .filter((t) => t.type === 'expense')
    .reduce((sum, t) => sum + parseFloat(t.amount), 0);

  if (loading) {
    return <div className="text-center py-20 text-gray-400">Загрузка...</div>;
  }

  const formatAmount = (amount: string | number, currency = 'KZT') => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    try {
      return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency,
        minimumFractionDigits: 0,
      }).format(num);
    } catch {
      return `${num.toFixed(2)} ${currency}`;
    }
  };

  const txnTypeLabel: Record<string, string> = {
    income: 'Доход',
    expense: 'Расход',
    transfer: 'Перевод',
  };

  const txnTypeColor: Record<string, string> = {
    income: 'text-green-600',
    expense: 'text-red-600',
    transfer: 'text-blue-600',
  };

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">Дашборд</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <p className="text-sm text-gray-500">Общий баланс</p>
          <p className="text-3xl font-bold mt-1">{formatAmount(totalBalance)}</p>
          <p className="text-xs text-gray-400 mt-1">{accounts.length} счетов</p>
        </div>
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <p className="text-sm text-gray-500">Доходы (последние)</p>
          <p className="text-3xl font-bold text-green-600 mt-1">
            +{formatAmount(totalIncome)}
          </p>
        </div>
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <p className="text-sm text-gray-500">Расходы (последние)</p>
          <p className="text-3xl font-bold text-red-600 mt-1">
            -{formatAmount(totalExpense)}
          </p>
        </div>
      </div>

      {/* Accounts */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Счета</h2>
          <Link
            href="/accounts"
            className="text-sm text-indigo-600 hover:underline"
          >
            Все счета →
          </Link>
        </div>
        {accounts.length === 0 ? (
          <div className="bg-white rounded-xl border border-gray-200 p-8 text-center text-gray-400">
            Нет счетов.{' '}
            <Link href="/accounts" className="text-indigo-600 hover:underline">
              Создать первый →
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {accounts.map((acc) => (
              <div
                key={acc.id}
                className="bg-white rounded-xl border border-gray-200 p-4"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">{acc.name}</p>
                    <p className="text-xs text-gray-400 capitalize">{acc.type}</p>
                  </div>
                  <p className="font-semibold">
                    {formatAmount(acc.balance, acc.currency)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Recent Transactions */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Последние транзакции</h2>
          <Link
            href="/transactions"
            className="text-sm text-indigo-600 hover:underline"
          >
            Все транзакции →
          </Link>
        </div>
        {recentTxns.length === 0 ? (
          <div className="bg-white rounded-xl border border-gray-200 p-8 text-center text-gray-400">
            Нет транзакций.{' '}
            <Link href="/transactions" className="text-indigo-600 hover:underline">
              Создать первую →
            </Link>
          </div>
        ) : (
          <div className="bg-white rounded-xl border border-gray-200 divide-y divide-gray-100">
            {recentTxns.map((txn) => (
              <div
                key={txn.id}
                className="flex items-center justify-between px-4 py-3"
              >
                <div>
                  <p className="font-medium text-sm">{txn.description}</p>
                  <p className="text-xs text-gray-400">
                    {txnTypeLabel[txn.type]} ·{' '}
                    {new Date(txn.transacted_at).toLocaleDateString('ru-RU')}
                    {txn.tags.length > 0 && (
                      <span className="ml-2">
                        {txn.tags.map((t) => `#${t}`).join(' ')}
                      </span>
                    )}
                  </p>
                </div>
                <p
                  className={`font-semibold text-sm ${txnTypeColor[txn.type]}`}
                >
                  {txn.type === 'expense' ? '-' : txn.type === 'income' ? '+' : ''}
                  {formatAmount(txn.amount)}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
