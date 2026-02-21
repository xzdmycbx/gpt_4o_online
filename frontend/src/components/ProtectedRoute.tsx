import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import styled from 'styled-components';

const ErrorContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--bg-primary);
  padding: 16px;
`;

const ErrorCard = styled.div`
  background-color: var(--bg-tertiary);
  border-radius: 12px;
  padding: 48px;
  text-align: center;
  max-width: 500px;
`;

const ErrorTitle = styled.h1`
  font-size: 48px;
  font-weight: 700;
  color: #f28b82;
  margin-bottom: 16px;
`;

const ErrorMessage = styled.p`
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 8px;
`;

const ErrorSubMessage = styled.p`
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 24px;
`;

const BackButton = styled.a`
  display: inline-block;
  padding: 12px 24px;
  border-radius: 8px;
  background-color: #2b5278;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 500;
  text-decoration: none;
  transition: background-color 150ms ease-in-out;

  &:hover {
    background-color: #3a6a95;
  }
`;

interface ProtectedRouteProps {
  allowedRoles?: string[];
  requireAuth?: boolean;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  allowedRoles = [],
  requireAuth = true,
}) => {
  const { user, loading } = useAuth();

  // Show loading state while checking authentication
  if (loading) {
    return (
      <ErrorContainer>
        <ErrorCard>
          <ErrorMessage>Loading...</ErrorMessage>
        </ErrorCard>
      </ErrorContainer>
    );
  }

  // Redirect to login if authentication is required but user is not logged in
  if (requireAuth && !user) {
    return <Navigate to="/login" replace />;
  }

  // Check if user role is allowed
  if (allowedRoles.length > 0 && user && !allowedRoles.includes(user.role)) {
    return (
      <ErrorContainer>
        <ErrorCard>
          <ErrorTitle>403</ErrorTitle>
          <ErrorMessage>Access Denied</ErrorMessage>
          <ErrorSubMessage>
            You don't have permission to access this page.
            {user.role === 'user' && ' This page is restricted to administrators.'}
          </ErrorSubMessage>
          <BackButton href="/">Go to Home</BackButton>
        </ErrorCard>
      </ErrorContainer>
    );
  }

  // User is authenticated and authorized, render the protected content
  return <Outlet />;
};

export default ProtectedRoute;

