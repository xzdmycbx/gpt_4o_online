import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useAuth } from '../../contexts/AuthContext';
import { media } from '../../styles/responsive';

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--bg-primary);
  padding: 16px;
`;

const Card = styled.div`
  background-color: var(--bg-tertiary);
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
  color: var(--text-primary);
  margin-bottom: 8px;
  text-align: center;
`;

const Subtitle = styled.p`
  font-size: 14px;
  color: var(--text-secondary);
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
  color: var(--text-primary);
`;

const Input = styled.input`
  padding: 12px;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: 16px;
  transition: border-color 150ms ease-in-out;

  &:focus {
    outline: none;
    border-color: #2b5278;
  }

  &::placeholder {
    color: var(--text-muted);
  }
`;

const Button = styled.button`
  padding: 12px;
  border: none;
  border-radius: 8px;
  background-color: #2b5278;
  color: var(--text-primary);
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
  color: var(--text-secondary);

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
  color: var(--text-secondary);
  font-size: 14px;

  &::before,
  &::after {
    content: '';
    flex: 1;
    border-bottom: 1px solid var(--border-primary);
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
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 150ms ease-in-out;
  margin-bottom: 8px;

  &:hover {
    background-color: #2a3542;
    border-color: var(--text-muted);
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

interface OAuthProvider {
  name: string;
  display_name: string;
  auth_url: string;
}

const TwitterIcon = () => (
  <svg viewBox="0 0 24 24" fill="currentColor">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
  </svg>
);

const GenericOAuthIcon = () => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <circle cx="12" cy="12" r="10"/>
    <path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
  </svg>
);

const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);

  const { login } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const apiUrl = import.meta.env.VITE_API_URL || '/api/v1';
    fetch(`${apiUrl}/auth/oauth2/providers`)
      .then(r => r.json())
      .then(data => {
        if (Array.isArray(data?.providers)) {
          setOauthProviders(data.providers);
        }
      })
      .catch(() => {
        // silently ignore — OAuth block will stay hidden
      });
  }, []);

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

  const handleOAuthLogin = (authUrl: string) => {
    const apiUrl = import.meta.env.VITE_API_URL || '/api/v1';
    // authUrl is already an absolute path like /api/v1/auth/oauth2/twitter
    window.location.href = authUrl.startsWith('/api') ? authUrl : `${apiUrl}${authUrl}`;
  };

  return (
    <Container>
      <Card>
        <Title>Welcome Back</Title>
        <Subtitle>Sign in to continue to AI Chat</Subtitle>

        {error && <Error>{error}</Error>}

        {oauthProviders.length > 0 && (
          <>
            {oauthProviders.map(provider => (
              <OAuth2Button
                key={provider.name}
                type="button"
                onClick={() => handleOAuthLogin(provider.auth_url)}
              >
                {provider.name === 'twitter' ? <TwitterIcon /> : <GenericOAuthIcon />}
                Continue with {provider.display_name}
              </OAuth2Button>
            ))}
            <Divider>OR</Divider>
          </>
        )}

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
