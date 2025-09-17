import React, { useState } from 'react';
import Navigation from './Navigation';
import Sidebar from './Sidebar';

interface AuthenticatedLayoutProps {
  children: React.ReactNode;
}

const AuthenticatedLayout: React.FC<AuthenticatedLayoutProps> = ({ children }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <Navigation onSidebarToggle={() => setIsSidebarOpen(true)} />
      <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />
      <main className="max-w-7xl mx-auto px-6 py-8">
        {children}
      </main>
    </div>
  );
};

export default AuthenticatedLayout;