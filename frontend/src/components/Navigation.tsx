import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { User, LogOut, ChevronDown, Menu } from 'lucide-react';
import { useAuth } from '@/auth/useAuth';

interface NavigationProps {
  onSidebarToggle: () => void;
}

const Navigation: React.FC<NavigationProps> = ({ onSidebarToggle }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <header className="border-b border-gray-800 bg-gray-900/50">
      <div className="max-w-7xl mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <button
              onClick={onSidebarToggle}
              className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
            >
              <Menu className="w-5 h-5 text-gray-400" />
            </button>
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-sm font-bold">R</span>
            </div>
            <h1 className="text-xl font-semibold text-white">Ripper Dashboard</h1>
          </div>
          
          <div className="relative">
            <button
              onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              className="flex items-center space-x-2 px-3 py-2 rounded-lg hover:bg-gray-800 transition-colors"
            >
              <User className="w-5 h-5 text-gray-400" />
              <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
            
            {isDropdownOpen && (
              <div className="absolute right-0 mt-2 w-64 bg-gray-900 border border-gray-800 rounded-lg shadow-lg z-50">
                <div className="p-4 border-b border-gray-800">
                  <p className="text-white font-medium">{user?.username}</p>
                  <p className="text-gray-400 text-sm">{user?.email}</p>
                </div>
                <div className="p-2">
                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center space-x-2 px-3 py-2 text-red-400 hover:bg-red-900/20 rounded-lg transition-colors"
                  >
                    <LogOut className="w-4 h-4" />
                    <span>Logout</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navigation;