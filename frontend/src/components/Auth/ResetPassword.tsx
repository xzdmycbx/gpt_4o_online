import React, { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import styled from 'styled-components';
import apiClient from '../../api/client';
import { media } from '../../styles/responsive';

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
  padding: 32px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3);

  ${media.mobile} {
    padding: 24px;
  }
`;

const Title = styled.h1`
  font-size: 24px;
  font-weight: 600;
  color: #e8eaed;
  margin-bottom: 8px;
  text-align: center;
`;

const Subtitle = styled.p`
  font-size: 14px;
  color: #9aa0a6;
  margin-bottom: 32px;
  text-align: center;
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 16px;
`;

const InputGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

const Label = styled.label`
  font-size: 14px;
  font-weight: 500;
  color: #e8eaed;
`;

const Input = styled.input`
  padding: 12px;
  border: 1px solid #3c4043;
  border-radius: 8px;
  background-color: #0e1621;
  color: #e8eaed;
  font-size: 16px;
  transition: border-color 150ms ease-in-out;

  &:focus {
    outline: none;
    border-color: #2b5278;
  }

  &::placeholder {
    color: #5f6368;
  }
`;

const Button = styled.button`
  padding: 12px;
  border: none;
  border-radius: 8px;
  background-color: #2b5278;
  color: #e8eaed;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 150ms ease-in-out;

  &:hover {
    background-color: #3a6a95;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const Message = styled.div<{ $isError?: boolean }>`
  padding: 12px;
  border-radius: 8px;
  background-color: ${props => props.$isError ? 'rgba(242, 139, 130, 0.1)' : 'rgba(138, 180, 248, 0.1)'};
  color: ${props => props.$isError ? '#f28b82' : '#8ab4f8'};
  font-size: 14px;
  border: 1px solid ${props => props.$isError ? 'rgba(242, 139, 130, 0.3)' : 'rgba(138, 180, 248, 0.3)'};
`;

const PasswordHint = styled.div`
  font-size: 12px;
  color: #9aa0a6;
  margin-top: -4px;
`;

const Footer = styled.div`
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: #9aa0a6;

  a {
    color: #2b5278;
    text-decoration: none;
    font-weight: 500;

    &:hover {
      text-decoration: underline;
    }
  }
`;

const ResetPassword: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [resetSuccess, setResetSuccess] = useState(false);
  const navigate = useNavigate();

  const token = searchParams.get('token');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setMessage('');

    if (!token) {
      setError('Invalid or missing reset token');
      return;
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters long');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setIsLoading(true);

    try {
      await apiClient.post('/auth/reset-password', {
        token,
        new_password: newPassword,
      });
      setResetSuccess(true);
      setMessage('Password reset successfully! Redirecting to login...');
      setTimeout(() => navigate('/login'), 3000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to reset password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (!token) {
    return (
      <Container>
        <Card>
          <Title>Invalid Link</Title>
          <Subtitle>This password reset link is invalid or has expired</Subtitle>
          <Footer>
            <Link to="/forgot-password">Request a new reset link</Link>
          </Footer>
        </Card>
      </Container>
    );
  }

  return (
    <Container>
      <Card>
        <Title>Reset Password</Title>
        <Subtitle>Enter your new password</Subtitle>

        {error && <Message $isError>{error}</Message>}
        {message && <Message>{message}</Message>}

        {!resetSuccess && (
          <Form onSubmit={handleSubmit}>
            <InputGroup>
              <Label htmlFor="newPassword">New Password</Label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password"
                required
                autoComplete="new-password"
                minLength={8}
              />
              <PasswordHint>At least 8 characters</PasswordHint>
            </InputGroup>

            <InputGroup>
              <Label htmlFor="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm new password"
                required
                autoComplete="new-password"
                minLength={8}
              />
            </InputGroup>

            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Resetting...' : 'Reset Password'}
            </Button>
          </Form>
        )}

        <Footer>
          Remember your password? <Link to="/login">Sign in</Link>
        </Footer>
      </Card>
    </Container>
  );
};

export default ResetPassword;
