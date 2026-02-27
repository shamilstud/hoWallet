'use client';

import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import { api } from '../../lib/api';
import type { Household } from '../../types';

export default function AppShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [households, setHouseholds] = useState<Household[]>([]);
  const [currentHH, setCurrentHH] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);
  const [isAuth, setIsAuth] = useState(false);

  useEffect(() => {
    setMounted(true);
    setIsAuth(api.isAuthenticated());
  }, []);

  useEffect(() => {
    if (!isAuth) return;
    api.listHouseholds().then((hh) => {
      setHouseholds(hh);
      const saved = api.getHouseholdId();
      if (saved && hh.find((h) => h.id === saved)) {
        setCurrentHH(saved);
      } else if (hh.length > 0) {
        api.setHouseholdId(hh[0].id);
        setCurrentHH(hh[0].id);
      }
    }).catch(() => {});
  }, [isAuth]);

  const handleHHChange = (id: string) => {
    api.setHouseholdId(id);
    setCurrentHH(id);
    router.refresh();
  };

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch {}
    api.clearTokens();
    router.push('/login');
  };

  if (!mounted) {
    return <div className="min-h-screen bg-gray-50" />;
  }

  if (!isAuth) return <>{children}</>;

  const navItems = [
    { href: '/dashboard', label: 'Дашборд' },
    { href: '/accounts', label: 'Счета' },
    { href: '/transactions', label: 'Транзакции' },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-8">
              <Link href="/dashboard" className="text-xl font-bold text-indigo-600">
                hoWallet
              </Link>
              <nav className="hidden md:flex gap-1">
                {navItems.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                      pathname === item.href
                        ? 'bg-indigo-50 text-indigo-700'
                        : 'text-gray-600 hover:bg-gray-100'
                    }`}
                  >
                    {item.label}
                  </Link>
                ))}
              </nav>
            </div>
            <div className="flex items-center gap-4">
              {households.length > 0 && (
                <select
                  value={currentHH || ''}
                  onChange={(e) => handleHHChange(e.target.value)}
                  className="text-sm border border-gray-300 rounded-md px-2 py-1.5"
                >
                  {households.map((hh) => (
                    <option key={hh.id} value={hh.id}>
                      {hh.name}
                    </option>
                  ))}
                </select>
              )}
              <button
                onClick={handleLogout}
                className="text-sm text-gray-600 hover:text-red-600 transition-colors"
              >
                Выйти
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
}
