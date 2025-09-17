import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/auth/useAuth';

const DashboardPage: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    console.log('User logged out, redirecting to login page');
    navigate('/login', { replace: true });
  };

  return (
    <div style={{ padding: '20px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '30px' }}>
        <h1>Dashboard</h1>
        <button onClick={handleLogout} style={{ padding: '8px 16px' }}>
          Logout
        </button>
      </div>
      
      <div style={{ backgroundColor: '#f5f5f5', padding: '20px', borderRadius: '8px', marginBottom: '20px' }}>
        <h3>Welcome back!</h3>
        <p><strong>Name:</strong> {user?.username}</p>
        <p><strong>Email:</strong> {user?.email}</p>
        <p><strong>User ID:</strong> {user?.user_id}</p>
      </div>

      <div style={{ backgroundColor: '#e8f5e8', padding: '20px', borderRadius: '8px' }}>
        <h3>🎉 Authentication Successful</h3>
        <p>You are now logged in and can access protected content.</p>
        <p>Your session is securely managed via WebSocket connection.</p>
      </div>
    </div>
  );
};

export default DashboardPage;