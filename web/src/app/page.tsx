'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '../lib/api';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    if (api.isAuthenticated()) {
      router.replace('/dashboard');
    } else {
      router.replace('/login');
    }
  }, [router]);

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div className="animate-pulse text-lg text-gray-500">Загрузка...</div>
    </div>
  );
}
