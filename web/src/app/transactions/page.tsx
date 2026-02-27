'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '../../lib/api';
import type { Account, Transaction, TransactionType } from '../../types';

export default function TransactionsPage() {
  const router = useRouter();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingTxn, setEditingTxn] = useState<Transaction | null>(null);

  // Filters
  const [filterType, setFilterType] = useState<string>('');
  const [filterAccountId, setFilterAccountId] = useState<string>('');
  const [page, setPage] = useState(0);
  const limit = 20;

  // Form state
  const [formType, setFormType] = useState<TransactionType>('expense');
  const [formDesc, setFormDesc] = useState('');
  const [formAmount, setFormAmount] = useState('');
  const [formAccountId, setFormAccountId] = useState('');
  const [formDestAccountId, setFormDestAccountId] = useState('');
  const [formTags, setFormTags] = useState('');
  const [formNote, setFormNote] = useState('');
  const [formDate, setFormDate] = useState(new Date().toISOString().slice(0, 16));
  const [formError, setFormError] = useState('');

  useEffect(() => {
    if (!api.isAuthenticated()) {
      router.replace('/login');
      return;
    }
    loadData();
  }, [router, page, filterType, filterAccountId]);

  const loadData = async () => {
    try {
      const [accs] = await Promise.all([api.listAccounts()]);
      setAccounts(accs);

      const params: Record<string, string> = {
        limit: String(limit),
        offset: String(page * limit),
      };
      if (filterType) params.type = filterType;
      if (filterAccountId) params.account_id = filterAccountId;

      const res = await api.listTransactions(params);
      setTransactions(res.data);
      setTotal(res.total);
    } catch {}
    setLoading(false);
  };

  const resetForm = () => {
    setFormType('expense');
    setFormDesc('');
    setFormAmount('');
    setFormAccountId('');
    setFormDestAccountId('');
    setFormTags('');
    setFormNote('');
    setFormDate(new Date().toISOString().slice(0, 16));
    setFormError('');
    setEditingTxn(null);
    setShowForm(false);
  };

  const openEdit = (txn: Transaction) => {
    setFormType(txn.type);
    setFormDesc(txn.description);
    setFormAmount(txn.amount);
    setFormAccountId(txn.account_id);
    setFormDestAccountId(txn.destination_account_id || '');
    setFormTags(txn.tags.join(', '));
    setFormNote(txn.note || '');
    setFormDate(new Date(txn.transacted_at).toISOString().slice(0, 16));
    setEditingTxn(txn);
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const tags = formTags
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean);

    const body = {
      type: formType,
      description: formDesc,
      amount: formAmount,
      account_id: formAccountId,
      destination_account_id:
        formType === 'transfer' && formDestAccountId ? formDestAccountId : undefined,
      tags,
      note: formNote || undefined,
      transacted_at: new Date(formDate).toISOString(),
    };

    try {
      if (editingTxn) {
        await api.updateTransaction(editingTxn.id, body);
      } else {
        await api.createTransaction(body);
      }
      resetForm();
      loadData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Ошибка');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Удалить транзакцию?')) return;
    try {
      await api.deleteTransaction(id);
      loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Не удалось удалить');
    }
  };

  const handleExport = async () => {
    try {
      await api.exportCSV();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Ошибка экспорта');
    }
  };

  const accountName = (id: string) =>
    accounts.find((a) => a.id === id)?.name || '—';

  const formatAmount = (amount: string, currency = 'KZT') => {
    const num = parseFloat(amount);
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

  const typeLabels: Record<TransactionType, string> = {
    income: 'Доход',
    expense: 'Расход',
    transfer: 'Перевод',
  };

  const typeColors: Record<TransactionType, string> = {
    income: 'text-green-600 bg-green-50',
    expense: 'text-red-600 bg-red-50',
    transfer: 'text-blue-600 bg-blue-50',
  };

  const totalPages = Math.ceil(total / limit);

  if (loading) {
    return <div className="text-center py-20 text-gray-400">Загрузка...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-bold">Транзакции</h1>
        <div className="flex gap-2">
          <button
            onClick={handleExport}
            className="border border-gray-300 rounded-lg px-4 py-2 text-sm font-medium hover:bg-gray-50 transition-colors"
          >
            Экспорт CSV
          </button>
          <button
            onClick={() => {
              resetForm();
              setShowForm(true);
            }}
            className="bg-indigo-600 text-white rounded-lg px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
          >
            + Новая транзакция
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-3 flex-wrap">
        <select
          value={filterType}
          onChange={(e) => {
            setFilterType(e.target.value);
            setPage(0);
          }}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm"
        >
          <option value="">Все типы</option>
          <option value="income">Доход</option>
          <option value="expense">Расход</option>
          <option value="transfer">Перевод</option>
        </select>
        <select
          value={filterAccountId}
          onChange={(e) => {
            setFilterAccountId(e.target.value);
            setPage(0);
          }}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm"
        >
          <option value="">Все счета</option>
          {accounts.map((acc) => (
            <option key={acc.id} value={acc.id}>
              {acc.name}
            </option>
          ))}
        </select>
      </div>

      {/* Form */}
      {showForm && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold mb-4">
            {editingTxn ? 'Редактировать транзакцию' : 'Новая транзакция'}
          </h2>
          {formError && (
            <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">
              {formError}
            </div>
          )}
          <form onSubmit={handleSubmit} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Тип</label>
              <select
                value={formType}
                onChange={(e) => setFormType(e.target.value as TransactionType)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              >
                <option value="expense">Расход</option>
                <option value="income">Доход</option>
                <option value="transfer">Перевод</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Описание
              </label>
              <input
                type="text"
                value={formDesc}
                onChange={(e) => setFormDesc(e.target.value)}
                required
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                placeholder="Продукты в АТБ"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Сумма
              </label>
              <input
                type="number"
                step="0.01"
                min="0.01"
                value={formAmount}
                onChange={(e) => setFormAmount(e.target.value)}
                required
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Счёт {formType === 'transfer' ? '(откуда)' : ''}
              </label>
              <select
                value={formAccountId}
                onChange={(e) => setFormAccountId(e.target.value)}
                required
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              >
                <option value="">Выберите счёт</option>
                {accounts.map((acc) => (
                  <option key={acc.id} value={acc.id}>
                    {acc.name} ({formatAmount(acc.balance, acc.currency)})
                  </option>
                ))}
              </select>
            </div>
            {formType === 'transfer' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Счёт (куда)
                </label>
                <select
                  value={formDestAccountId}
                  onChange={(e) => setFormDestAccountId(e.target.value)}
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                >
                  <option value="">Выберите счёт</option>
                  {accounts
                    .filter((a) => a.id !== formAccountId)
                    .map((acc) => (
                      <option key={acc.id} value={acc.id}>
                        {acc.name} ({formatAmount(acc.balance, acc.currency)})
                      </option>
                    ))}
                </select>
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Дата и время
              </label>
              <input
                type="datetime-local"
                value={formDate}
                onChange={(e) => setFormDate(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Тэги (через запятую)
              </label>
              <input
                type="text"
                value={formTags}
                onChange={(e) => setFormTags(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                placeholder="Еда, Продукты"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Заметка
              </label>
              <textarea
                value={formNote}
                onChange={(e) => setFormNote(e.target.value)}
                rows={2}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none resize-none"
              />
            </div>
            <div className="sm:col-span-2 flex gap-2">
              <button
                type="submit"
                className="bg-indigo-600 text-white rounded-lg px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                {editingTxn ? 'Сохранить' : 'Создать'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="border border-gray-300 rounded-lg px-4 py-2 text-sm font-medium hover:bg-gray-50 transition-colors"
              >
                Отмена
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Transactions table */}
      {transactions.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 p-12 text-center text-gray-400">
          Нет транзакций
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Дата</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Описание</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Тип</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Счёт</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500">Сумма</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Тэги</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {transactions.map((txn) => (
                  <tr key={txn.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap text-gray-600">
                      {new Date(txn.transacted_at).toLocaleDateString('ru-RU')}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium">{txn.description}</div>
                      {txn.note && (
                        <div className="text-xs text-gray-400 mt-0.5">{txn.note}</div>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs font-medium px-2 py-1 rounded ${typeColors[txn.type]}`}
                      >
                        {typeLabels[txn.type]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600">
                      {accountName(txn.account_id)}
                      {txn.type === 'transfer' && txn.destination_account_id && (
                        <span className="text-gray-400">
                          {' → '}
                          {accountName(txn.destination_account_id)}
                        </span>
                      )}
                    </td>
                    <td
                      className={`px-4 py-3 text-right font-semibold ${
                        txn.type === 'expense'
                          ? 'text-red-600'
                          : txn.type === 'income'
                          ? 'text-green-600'
                          : 'text-blue-600'
                      }`}
                    >
                      {txn.type === 'expense' ? '-' : txn.type === 'income' ? '+' : ''}
                      {formatAmount(txn.amount)}
                    </td>
                    <td className="px-4 py-3 text-gray-400 text-xs">
                      {txn.tags.map((t) => `#${t}`).join(' ')}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => openEdit(txn)}
                        className="text-xs text-indigo-600 hover:underline mr-2"
                      >
                        Изм.
                      </button>
                      <button
                        onClick={() => handleDelete(txn.id)}
                        className="text-xs text-red-500 hover:underline"
                      >
                        Уд.
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200">
              <p className="text-sm text-gray-500">
                Показано {page * limit + 1}–{Math.min((page + 1) * limit, total)} из {total}
              </p>
              <div className="flex gap-1">
                <button
                  disabled={page === 0}
                  onClick={() => setPage(page - 1)}
                  className="px-3 py-1 text-sm border rounded disabled:opacity-30 hover:bg-gray-50"
                >
                  ←
                </button>
                <button
                  disabled={page >= totalPages - 1}
                  onClick={() => setPage(page + 1)}
                  className="px-3 py-1 text-sm border rounded disabled:opacity-30 hover:bg-gray-50"
                >
                  →
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
