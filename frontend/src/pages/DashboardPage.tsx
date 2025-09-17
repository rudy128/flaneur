import React from 'react';
import { useAuth } from '@/auth/useAuth';

const DashboardPage: React.FC = () => {
  const { user } = useAuth();

  return (
    <>
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold mb-2">Welcome back, {user?.username}!</h2>
          <p className="text-gray-400">Here's what's happening with your account today.</p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm">Active Sessions</p>
                <p className="text-2xl font-bold text-green-400">1</p>
              </div>
              <div className="w-12 h-12 bg-green-600/20 rounded-lg flex items-center justify-center">
                <div className="w-6 h-6 bg-green-600 rounded-full"></div>
              </div>
            </div>
          </div>
          
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm">Account Status</p>
                <p className="text-2xl font-bold text-blue-400">Active</p>
              </div>
              <div className="w-12 h-12 bg-blue-600/20 rounded-lg flex items-center justify-center">
                <div className="w-6 h-6 bg-blue-600 rounded-full"></div>
              </div>
            </div>
          </div>
          
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm">Last Login</p>
                <p className="text-2xl font-bold text-purple-400">Now</p>
              </div>
              <div className="w-12 h-12 bg-purple-600/20 rounded-lg flex items-center justify-center">
                <div className="w-6 h-6 bg-purple-600 rounded-full"></div>
              </div>
            </div>
          </div>
        </div>

        {/* User Info Card */}
        <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 mb-8">
          <h3 className="text-lg font-semibold mb-4 flex items-center">
            <div className="w-5 h-5 bg-blue-600 rounded mr-2"></div>
            Account Information
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-3">
              <div>
                <label className="text-gray-400 text-sm">Username</label>
                <p className="text-white font-medium">{user?.username}</p>
              </div>
              <div>
                <label className="text-gray-400 text-sm">Email</label>
                <p className="text-white font-medium">{user?.email}</p>
              </div>
            </div>
            <div className="space-y-3">
              <div>
                <label className="text-gray-400 text-sm">User ID</label>
                <p className="text-white font-medium font-mono text-sm">{user?.user_id}</p>
              </div>
              <div>
                <label className="text-gray-400 text-sm">Connection</label>
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                  <p className="text-green-400 font-medium">WebSocket Active</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Success Message */}
        <div className="bg-green-900/20 border border-green-800 rounded-xl p-6">
          <div className="flex items-start space-x-3">
            <div className="w-6 h-6 bg-green-600 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-white text-sm">✓</span>
            </div>
            <div>
              <h3 className="text-green-400 font-semibold mb-1">Authentication Successful</h3>
              <p className="text-green-300/80 text-sm leading-relaxed">
                You are now logged in and can access all protected content. Your session is securely managed via WebSocket connection with real-time updates.
              </p>
            </div>
          </div>
        </div>
    </>
  );
};

export default DashboardPage;