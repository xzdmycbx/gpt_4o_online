import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import styled from 'styled-components';
import { useAuth } from '../../contexts/AuthContext';
import apiClient from '../../api/client';

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: #0e1621;
  padding: 16px;
`;

const Card = styled.div`
  background-color: #1e2832;
  border-radius: 12px;
  padding: 48px;
  text-align: center;
  max-width: 400px;
`;

const Spinner = styled.div`
  border: 3px solid #3c4043;
  border-top: 3px solid #2b5278;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  animation: spin 1s linear infinite;
  margin: 0 auto 24px;

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const Message = styled.p`
  color: #e8eaed;
  font-size: 16px;
  margin-bottom: 8px;
`;

const ErrorMessage = styled(Message)`
  color: #f28b82;
`;

const SubMessage = styled.p`
  color: #9aa0a6;
  font-size: 14px;
`;

const OAuth2Callback: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { setUser } = useAuth();

  useEffect(() => {
    const handleCallback = async () => {
      const errorParam = searchParams.get('error');

      if (errorParam) {
        const errorMessages: { [key: string]: string } = {
          'oauth_failed': 'OAuth authentication was cancelled or failed',
          'missing_code': 'Authorization code is missing',
          'invalid_state': 'Invalid OAuth state - possible CSRF attack',
          'missing_verifier': 'OAuth verifier is missing',
          'auth_failed': 'Authentication failed - please try again',
        };
        setError(errorMessages[errorParam] || 'Authentication failed');
        setTimeout(() => navigate('/login'), 3000);
        return;
      }

      // Read token from HttpOnly cookie (set by backend)
      // The token is in auth_token cookie, accessible via /api/v1/me endpoint
      try {
        // Fetch user info - this will use the cookie automatically
        const response = await apiClient.auth.getCurrentUser();

        if (response.user) {
          setUser(response.user);
          // Note: We don't store token in localStorage anymore
          // The token is in HttpOnly cookie managed by the browser
          navigate('/');
        } else {
          throw new Error('Failed to fetch user information');
        }
      } catch (err: any) {
        console.error('OAuth2 callback error:', err);
        setError('Failed to complete authentication');
        setTimeout(() => navigate('/login'), 3000);
      }
    };

    handleCallback();
  }, [searchParams, navigate, setUser]);

  return (
    <Container>
      <Card>
        {error ? (
          <>
            <ErrorMessage>{error}</ErrorMessage>
            <SubMessage>Redirecting to login...</SubMessage>
          </>
        ) : (
          <>
            <Spinner />
            <Message>Completing authentication</Message>
            <SubMessage>Please wait...</SubMessage>
          </>
        )}
      </Card>
    </Container>
  );
};

export default OAuth2Callback;
