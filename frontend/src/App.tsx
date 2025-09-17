import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import AuthProvider from '@/auth/AuthContext';
import ProtectedRoute from '@/components/ProtectedRoute';
import AuthRedirect from '@/components/AuthRedirect';
import LoginPage from '@/pages/LoginPage';
import SignupPage from '@/pages/SignupPage';
import DashboardPage from '@/pages/DashboardPage';
function App() {
  console.log('App component loaded with React Router');
  
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/login" element={
            <AuthRedirect>
              <LoginPage />
            </AuthRedirect>
          } />
          <Route path="/signup" element={
            <AuthRedirect>
              <SignupPage />
            </AuthRedirect>
          } />
          <Route path="/dashboard" element={
            <ProtectedRoute>
              <DashboardPage />
            </ProtectedRoute>
          } />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
