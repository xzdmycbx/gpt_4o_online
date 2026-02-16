import React, { useState } from 'react';
import { Link } from 'react-router-dom';
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

const ForgotPassword: React.FC = () => {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [emailSent, setEmailSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      await apiClient.post('/auth/forgot-password', { email });
      setEmailSent(true);
      setMessage('Password reset link has been sent to your email. Please check your inbox.');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send reset link. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Container>
      <Card>
        <Title>Forgot Password</Title>
        <Subtitle>
          {emailSent
            ? 'Check your email for reset instructions'
            : 'Enter your email to receive a password reset link'}
        </Subtitle>

        {error && <Message $isError>{error}</Message>}
        {message && <Message>{message}</Message>}

        {!emailSent && (
          <Form onSubmit={handleSubmit}>
            <InputGroup>
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="Enter your email"
                required
                autoComplete="email"
              />
            </InputGroup>

            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Sending...' : 'Send Reset Link'}
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

export default ForgotPassword;
