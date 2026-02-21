import React, { useState } from 'react';
import styled from 'styled-components';
import { useAuth } from '../../contexts/AuthContext';
import { useSettings } from '../../contexts/SettingsContext';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../api/client';
import { media } from '../../styles/responsive';

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;

  ${media.mobile} {
    padding: 16px;
  }
`;

const Title = styled.h1`
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
`;

const Subtitle = styled.p`
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 32px;
`;

const Section = styled.div`
  background-color: var(--bg-tertiary);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
`;

const SectionTitle = styled.h2`
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
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

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const Button = styled.button`
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  background-color: #2b5278;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 150ms ease-in-out;
  width: fit-content;

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

const InfoText = styled.p`
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 8px;
`;

const UserInfo = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
`;

const InfoItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: var(--bg-primary);
  border-radius: 8px;

  ${media.mobile} {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
`;

const InfoLabel = styled.span`
  font-size: 14px;
  color: var(--text-secondary);
`;

const InfoValue = styled.span`
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
`;

const Select = styled.select`
  width: 100%;
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
`;

const ActionRow = styled.div`
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
`;

const Settings: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { settings, updateSettings } = useSettings();
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setMessage('');

    if (newPassword.length < 8) {
      setError('New password must be at least 8 characters long');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match');
      return;
    }

    setIsLoading(true);

    try {
      await apiClient.put('/user/password', {
        current_password: currentPassword,
        new_password: newPassword,
      });

      setMessage('Password changed successfully!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to change password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (!user) {
    return <Container>Loading...</Container>;
  }

  const canAccessAdmin = user.role === 'admin' || user.role === 'super_admin';

  return (
    <Container>
      <Title>Settings</Title>
      <Subtitle>Manage your account settings and preferences</Subtitle>
      <ActionRow>
        <Button type="button" onClick={() => navigate('/chat')}>返回对话</Button>
        {canAccessAdmin && (
          <Button type="button" onClick={() => navigate('/admin')}>管理后台</Button>
        )}
      </ActionRow>

      <Section>
        <SectionTitle>Appearance</SectionTitle>
        <InputGroup>
          <Label htmlFor="themeMode">Theme Mode</Label>
          <Select
            id="themeMode"
            value={settings.theme}
            onChange={(e) => updateSettings({ theme: e.target.value as 'dark' | 'light' | 'auto' })}
          >
            <option value="dark">Dark</option>
            <option value="light">Light</option>
            <option value="auto">Auto (Follow System)</option>
          </Select>
          <InfoText>
            当前主题: {settings.theme === 'auto' ? 'Auto' : settings.theme === 'light' ? 'Light' : 'Dark'}
          </InfoText>
        </InputGroup>
      </Section>

      <Section>
        <SectionTitle>Account Information</SectionTitle>
        <UserInfo>
          <InfoItem>
            <InfoLabel>Username</InfoLabel>
            <InfoValue>{user.username}</InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>Email</InfoLabel>
            <InfoValue>{user.email || 'Not set'}</InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>Role</InfoLabel>
            <InfoValue>{user.role}</InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>Login Method</InfoLabel>
            <InfoValue>{user.oauth2_provider ? `OAuth2 (${user.oauth2_provider})` : 'Password'}</InfoValue>
          </InfoItem>
        </UserInfo>
      </Section>

      <Section>
        <SectionTitle>Change Password</SectionTitle>

        {user.oauth2_provider ? (
          <InfoText>
            You are logged in via OAuth2 ({user.oauth2_provider}). Password change is not available for OAuth2 accounts.
          </InfoText>
        ) : (
          <>
            {error && <Message $isError>{error}</Message>}
            {message && <Message>{message}</Message>}

            <Form onSubmit={handlePasswordChange}>
              <InputGroup>
                <Label htmlFor="currentPassword">Current Password</Label>
                <Input
                  id="currentPassword"
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="Enter current password"
                  required
                  autoComplete="current-password"
                />
              </InputGroup>

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
                <InfoText>At least 8 characters</InfoText>
              </InputGroup>

              <InputGroup>
                <Label htmlFor="confirmPassword">Confirm New Password</Label>
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
                {isLoading ? 'Changing Password...' : 'Change Password'}
              </Button>
            </Form>
          </>
        )}
      </Section>
    </Container>
  );
};

export default Settings;

