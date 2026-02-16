import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useAuth } from '../../contexts/AuthContext';
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

const Error = styled.div`
  padding: 12px;
  border-radius: 8px;
  background-color: rgba(242, 139, 130, 0.1);
  color: #f28b82;
  font-size: 14px;
  border: 1px solid rgba(242, 139, 130, 0.3);
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

const Divider = styled.div`
  display: flex;
  align-items: center;
  margin: 24px 0;
  color: #9aa0a6;
  font-size: 14px;

  &::before,
  &::after {
    content: '';
    flex: 1;
    border-bottom: 1px solid #3c4043;
  }

  &::before {
    margin-right: 16px;
  }

  &::after {
    margin-left: 16px;
  }
`;

const OAuth2Button = styled.button`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  border: 1px solid #3c4043;
  border-radius: 8px;
  background-color: #1e2832;
  color: #e8eaed;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 150ms ease-in-out;

  &:hover {
    background-color: #2a3542;
    border-color: #5f6368;
  }

  svg {
    width: 20px;
    height: 20px;
  }
`;

const ForgotPassword = styled(Link)`
  display: block;
  margin-top: 8px;
  text-align: right;
  font-size: 14px;
  color: #2b5278;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }
`;

const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login(username, password);
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleTwitterLogin = () => {
    const apiUrl = import.meta.env.VITE_API_URL || '/api/v1';
    window.location.href = `${apiUrl}/auth/oauth2/twitter`;
  };

  return (
    <Container>
      <Card>
        <Title>Welcome Back</Title>
        <Subtitle>Sign in to continue to AI Chat</Subtitle>

        {error && <Error>{error}</Error>}

        <OAuth2Button type="button" onClick={handleTwitterLogin}>
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
          </svg>
          Continue with X (Twitter)
        </OAuth2Button>

        <Divider>OR</Divider>

        <Form onSubmit={handleSubmit}>
          <InputGroup>
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              required
              autoComplete="username"
            />
          </InputGroup>

          <InputGroup>
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              autoComplete="current-password"
            />
            <ForgotPassword to="/forgot-password">Forgot password?</ForgotPassword>
          </InputGroup>

          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Signing in...' : 'Sign In'}
          </Button>
        </Form>

        <Footer>
          Don't have an account? <Link to="/register">Sign up</Link>
        </Footer>
      </Card>
    </Container>
  );
};

export default Login;
