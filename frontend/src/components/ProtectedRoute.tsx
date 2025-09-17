import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '@/auth/useAuth';
import AuthenticatedLayout from './AuthenticatedLayout';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <AuthenticatedLayout>
      {children}
    </AuthenticatedLayout>
  );
};

export default ProtectedRoute;