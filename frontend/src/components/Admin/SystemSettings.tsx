import React, { useEffect, useMemo, useState } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import { GlassCard } from '../../styles/glass';

interface SystemSettingsDTO {
  rate_limit_default_per_minute: number;
  system_name: string;
  maintenance_mode: boolean;
  oauth2_twitter_enabled: boolean;
  oauth2_twitter_client_id: string;
  oauth2_twitter_client_secret: string;
  oauth2_twitter_redirect_url: string;
  email_enabled: boolean;
  email_provider: 'smtp' | 'resend';
  email_smtp_host: string;
  email_smtp_port: number;
  email_smtp_user: string;
  email_smtp_password: string;
  email_from: string;
  email_from_name: string;
  email_resend_api_key: string;
  ai_default_memory_model: string;
  ai_memory_extraction_enabled: boolean;
}

type MessageType = 'success' | 'error';
type SectionKey = 'basic' | 'oauth2' | 'email' | 'ai';

const DEFAULT_SETTINGS: SystemSettingsDTO = {
  rate_limit_default_per_minute: 20,
  system_name: 'AI Chat System',
  maintenance_mode: false,
  oauth2_twitter_enabled: false,
  oauth2_twitter_client_id: '',
  oauth2_twitter_client_secret: '',
  oauth2_twitter_redirect_url: '',
  email_enabled: false,
  email_provider: 'smtp',
  email_smtp_host: '',
  email_smtp_port: 587,
  email_smtp_user: '',
  email_smtp_password: '',
  email_from: 'noreply@example.com',
  email_from_name: 'AI Chat System',
  email_resend_api_key: '',
  ai_default_memory_model: 'gpt-3.5-turbo',
  ai_memory_extraction_enabled: true,
};

const DEFAULT_EXPANDED_STATE: Record<SectionKey, boolean> = {
  basic: true,
  oauth2: true,
  email: true,
  ai: true,
};

const PRESET_MEMORY_MODELS = ['gpt-3.5-turbo', 'gpt-4o-mini', 'gpt-4.1-mini', 'gpt-4o'];



const Card = styled(GlassCard)`
  margin-bottom: 24px;
  overflow: hidden;
`;

const CardToggle = styled.button`
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border: none;
  background: transparent;
  cursor: pointer;
`;

const CardTitle = styled.h2`
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
  text-align: left;
`;

const ToggleIcon = styled.span<{ expanded: boolean }>`
  font-size: 14px;
  color: var(--text-secondary);
  transform: rotate(${props => (props.expanded ? '180deg' : '0deg')});
  transition: transform 0.2s ease;
`;

const CardBody = styled.div`
  padding: 0 24px 24px;
  border-top: 1px solid var(--border-primary);
`;

const FormGroup = styled.div`
  margin-top: 20px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
`;

const SwitchLabel = styled.label`
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
`;

const Checkbox = styled.input`
  accent-color: #667eea;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;
  transition: all 0.2s;

  &:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }
`;

const Select = styled.select`
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;
  transition: all 0.2s;

  &:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }
`;

const Button = styled.button`
  padding: 12px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
`;

const SecondaryButton = styled(Button)`
  background: #4a5568;

  &:hover {
    box-shadow: none;
  }
`;

const Message = styled.div<{ type: MessageType }>`
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  background: ${props => (props.type === 'error' ? 'rgba(252, 129, 129, 0.1)' : 'rgba(72, 187, 120, 0.1)')};
  color: ${props => (props.type === 'error' ? '#fc8181' : '#48bb78')};
  border: 1px solid ${props => (props.type === 'error' ? '#fc8181' : '#48bb78')};
`;

const HelpText = styled.p`
  color: var(--text-muted);
  font-size: 13px;
  margin: 8px 0 0 0;
`;

const InlineActions = styled.div`
  display: flex;
  gap: 12px;
  align-items: flex-end;

  @media (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;
  }
`;

const InlineField = styled.div`
  flex: 1;
`;

const isMaskedValue = (value: string) => value.trim().startsWith('********');

const normalizeSensitiveValue = (value: string) => {
  const trimmed = value.trim();
  if (!trimmed || isMaskedValue(trimmed)) {
    return '';
  }
  return trimmed;
};

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const getErrorMessage = (error: unknown, fallback: string) => {
  if (typeof error !== 'object' || error === null) {
    return fallback;
  }

  const maybeAxiosError = error as {
    response?: { data?: { error?: string } };
    message?: string;
  };

  if (maybeAxiosError.response?.data?.error) {
    return maybeAxiosError.response.data.error;
  }

  if (maybeAxiosError.message) {
    return maybeAxiosError.message;
  }

  return fallback;
};

const SystemSettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<SystemSettingsDTO>(DEFAULT_SETTINGS);
  const [loading, setLoading] = useState(false);
  const [testingEmail, setTestingEmail] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: MessageType } | null>(null);
  const [testEmail, setTestEmail] = useState('');
  const [expanded, setExpanded] = useState<Record<SectionKey, boolean>>(DEFAULT_EXPANDED_STATE);
  const [hasStoredSensitive, setHasStoredSensitive] = useState({
    oauth2Secret: false,
    smtpPassword: false,
    resendApiKey: false,
  });
  const [availableModels, setAvailableModels] = useState<{ id: string; display_name: string; model_identifier: string; name: string }[]>([]);
  const [customModelIdentifier, setCustomModelIdentifier] = useState('');

  const memoryModelOptions = useMemo(() => {
    const current = settings.ai_default_memory_model.trim();
    // Build options from configured models + ensure current value is included
    const modelIdentifiers = availableModels.map(m => m.model_identifier);
    if (current && !modelIdentifiers.includes(current)) {
      return [current, ...modelIdentifiers];
    }
    return modelIdentifiers.length > 0 ? modelIdentifiers : PRESET_MEMORY_MODELS;
  }, [settings.ai_default_memory_model, availableModels]);

  useEffect(() => {
    void loadSettings();
    // Load available AI models for memory model selector
    apiClient.get('/admin/models').then(res => {
      const models = Array.isArray(res.data?.models) ? res.data.models : [];
      setAvailableModels(models);
    }).catch(() => {});
  }, []);

  const showMessage = (text: string, type: MessageType) => {
    setMessage({ text, type });
    window.setTimeout(() => setMessage(null), 3000);
  };

  const loadSettings = async () => {
    try {
      const response = await apiClient.get('/admin/system/settings');
      const data = response.data as Partial<SystemSettingsDTO>;

      setHasStoredSensitive({
        oauth2Secret: isMaskedValue(data.oauth2_twitter_client_secret ?? ''),
        smtpPassword: isMaskedValue(data.email_smtp_password ?? ''),
        resendApiKey: isMaskedValue(data.email_resend_api_key ?? ''),
      });

      setSettings({
        ...DEFAULT_SETTINGS,
        ...data,
        email_provider: data.email_provider === 'resend' ? 'resend' : 'smtp',
        email_smtp_port:
          typeof data.email_smtp_port === 'number' && data.email_smtp_port > 0
            ? data.email_smtp_port
            : DEFAULT_SETTINGS.email_smtp_port,
        oauth2_twitter_client_secret: '',
        email_smtp_password: '',
        email_resend_api_key: '',
      });
    } catch (error) {
      showMessage(getErrorMessage(error, 'Failed to load settings'), 'error');
    }
  };

  const toggleSection = (section: SectionKey) => {
    setExpanded(prev => ({ ...prev, [section]: !prev[section] }));
  };

  const validateSettings = (): string | null => {
    if (
      !Number.isInteger(settings.rate_limit_default_per_minute) ||
      settings.rate_limit_default_per_minute < 1 ||
      settings.rate_limit_default_per_minute > 1000
    ) {
      return 'Rate limit must be between 1 and 1000';
    }

    if (!settings.system_name.trim()) {
      return 'System name is required';
    }

    if (settings.oauth2_twitter_enabled) {
      if (!settings.oauth2_twitter_client_id.trim()) {
        return 'Twitter Client ID is required when OAuth2 is enabled';
      }

      if (!settings.oauth2_twitter_redirect_url.trim()) {
        return 'Twitter Redirect URL is required when OAuth2 is enabled';
      }

      try {
        new URL(settings.oauth2_twitter_redirect_url.trim());
      } catch {
        return 'Twitter Redirect URL is invalid';
      }

      if (!settings.oauth2_twitter_client_secret.trim() && !hasStoredSensitive.oauth2Secret) {
        return 'Twitter Client Secret is required when OAuth2 is enabled';
      }
    }

    if (settings.email_enabled) {
      if (!['smtp', 'resend'].includes(settings.email_provider)) {
        return 'Email provider must be smtp or resend';
      }

      if (!settings.email_from.trim()) {
        return 'Sender email is required when email service is enabled';
      }

      if (!emailRegex.test(settings.email_from.trim())) {
        return 'Sender email format is invalid';
      }

      if (!settings.email_from_name.trim()) {
        return 'Sender name is required when email service is enabled';
      }

      if (settings.email_provider === 'smtp') {
        if (!settings.email_smtp_host.trim()) {
          return 'SMTP host is required for SMTP provider';
        }

        if (!Number.isInteger(settings.email_smtp_port) || settings.email_smtp_port <= 0) {
          return 'SMTP port must be greater than 0';
        }
      }

      if (
        settings.email_provider === 'resend' &&
        !settings.email_resend_api_key.trim() &&
        !hasStoredSensitive.resendApiKey
      ) {
        return 'Resend API key is required for resend provider';
      }
    }

    if (!settings.ai_default_memory_model.trim()) {
      return 'Default memory model is required';
    }

    return null;
  };

  const handleSave = async () => {
    const validationError = validateSettings();
    if (validationError) {
      setMessage({ text: validationError, type: 'error' });
      return;
    }

    setLoading(true);
    setMessage(null);

    const payload: SystemSettingsDTO = {
      ...settings,
      system_name: settings.system_name.trim(),
      oauth2_twitter_client_id: settings.oauth2_twitter_client_id.trim(),
      oauth2_twitter_client_secret: normalizeSensitiveValue(settings.oauth2_twitter_client_secret),
      oauth2_twitter_redirect_url: settings.oauth2_twitter_redirect_url.trim(),
      email_smtp_host: settings.email_smtp_host.trim(),
      email_smtp_user: settings.email_smtp_user.trim(),
      email_smtp_password: normalizeSensitiveValue(settings.email_smtp_password),
      email_from: settings.email_from.trim(),
      email_from_name: settings.email_from_name.trim(),
      email_resend_api_key: normalizeSensitiveValue(settings.email_resend_api_key),
      ai_default_memory_model: settings.ai_default_memory_model.trim(),
    };

    try {
      await apiClient.put('/admin/system/settings', payload);
      await loadSettings();
      showMessage('Settings saved successfully', 'success');
    } catch (error) {
      setMessage({ text: getErrorMessage(error, 'Failed to save settings'), type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleSendTestEmail = async () => {
    if (!settings.email_enabled) {
      setMessage({ text: 'Email service is disabled', type: 'error' });
      return;
    }

    if (!testEmail.trim()) {
      setMessage({ text: 'Please input a test email address', type: 'error' });
      return;
    }

    if (!emailRegex.test(testEmail.trim())) {
      setMessage({ text: 'Test email format is invalid', type: 'error' });
      return;
    }

    setTestingEmail(true);
    setMessage(null);

    try {
      const response = await apiClient.post('/admin/system/test-email', {
        test_email: testEmail.trim(),
      });
      showMessage(response.data?.message || 'Test email sent successfully', 'success');
    } catch (error) {
      setMessage({ text: getErrorMessage(error, 'Failed to send test email'), type: 'error' });
    } finally {
      setTestingEmail(false);
    }
  };

  return (
    <div>
      {message && <Message type={message.type}>{message.text}</Message>}

      <Card>
        <CardToggle type="button" onClick={() => toggleSection('basic')}>
          <CardTitle>Basic Settings</CardTitle>
          <ToggleIcon expanded={expanded.basic}>▼</ToggleIcon>
        </CardToggle>
        {expanded.basic && (
          <CardBody>
            <FormGroup>
              <Label>Default Rate Limit (messages per minute)</Label>
              <Input
                type="number"
                min="1"
                max="1000"
                value={settings.rate_limit_default_per_minute}
                onChange={e =>
                  setSettings(prev => ({
                    ...prev,
                    rate_limit_default_per_minute: Number.parseInt(e.target.value, 10) || 0,
                  }))
                }
              />
              <HelpText>This value applies to users without custom rate-limit rules.</HelpText>
            </FormGroup>

            <FormGroup>
              <Label>System Name</Label>
              <Input
                type="text"
                value={settings.system_name}
                onChange={e => setSettings(prev => ({ ...prev, system_name: e.target.value }))}
              />
            </FormGroup>

            <FormGroup>
              <SwitchLabel>
                <Checkbox
                  type="checkbox"
                  checked={settings.maintenance_mode}
                  onChange={e => setSettings(prev => ({ ...prev, maintenance_mode: e.target.checked }))}
                />
                Maintenance Mode
              </SwitchLabel>
              <HelpText>When enabled, non-admin users cannot access the system.</HelpText>
            </FormGroup>
          </CardBody>
        )}
      </Card>

      <Card>
        <CardToggle type="button" onClick={() => toggleSection('oauth2')}>
          <CardTitle>OAuth2 Settings</CardTitle>
          <ToggleIcon expanded={expanded.oauth2}>▼</ToggleIcon>
        </CardToggle>
        {expanded.oauth2 && (
          <CardBody>
            <FormGroup>
              <SwitchLabel>
                <Checkbox
                  type="checkbox"
                  checked={settings.oauth2_twitter_enabled}
                  onChange={e => setSettings(prev => ({ ...prev, oauth2_twitter_enabled: e.target.checked }))}
                />
                Enable Twitter OAuth2
              </SwitchLabel>
            </FormGroup>

            {settings.oauth2_twitter_enabled && (
              <>
                <FormGroup>
                  <Label>Twitter Client ID</Label>
                  <Input
                    type="text"
                    value={settings.oauth2_twitter_client_id}
                    onChange={e =>
                      setSettings(prev => ({
                        ...prev,
                        oauth2_twitter_client_id: e.target.value,
                      }))
                    }
                  />
                </FormGroup>

                <FormGroup>
                  <Label>Twitter Client Secret</Label>
                  <Input
                    type="password"
                    value={settings.oauth2_twitter_client_secret}
                    placeholder={hasStoredSensitive.oauth2Secret ? 'Configured' : ''}
                    onChange={e =>
                      setSettings(prev => ({
                        ...prev,
                        oauth2_twitter_client_secret: e.target.value,
                      }))
                    }
                  />
                  <HelpText>留空表示不修改现有值</HelpText>
                </FormGroup>

                <FormGroup>
                  <Label>Redirect URL</Label>
                  <Input
                    type="url"
                    value={settings.oauth2_twitter_redirect_url}
                    onChange={e =>
                      setSettings(prev => ({
                        ...prev,
                        oauth2_twitter_redirect_url: e.target.value,
                      }))
                    }
                  />
                </FormGroup>
              </>
            )}
          </CardBody>
        )}
      </Card>

      <Card>
        <CardToggle type="button" onClick={() => toggleSection('email')}>
          <CardTitle>Email Settings</CardTitle>
          <ToggleIcon expanded={expanded.email}>▼</ToggleIcon>
        </CardToggle>
        {expanded.email && (
          <CardBody>
            <FormGroup>
              <SwitchLabel>
                <Checkbox
                  type="checkbox"
                  checked={settings.email_enabled}
                  onChange={e => setSettings(prev => ({ ...prev, email_enabled: e.target.checked }))}
                />
                Enable Email Service
              </SwitchLabel>
            </FormGroup>

            {settings.email_enabled && (
              <>
                <FormGroup>
                  <Label>Email Provider</Label>
                  <Select
                    value={settings.email_provider}
                    onChange={e =>
                      setSettings(prev => ({
                        ...prev,
                        email_provider: e.target.value === 'resend' ? 'resend' : 'smtp',
                      }))
                    }
                  >
                    <option value="smtp">smtp</option>
                    <option value="resend">resend</option>
                  </Select>
                </FormGroup>

                {settings.email_provider === 'smtp' && (
                  <>
                    <FormGroup>
                      <Label>SMTP Host</Label>
                      <Input
                        type="text"
                        value={settings.email_smtp_host}
                        onChange={e => setSettings(prev => ({ ...prev, email_smtp_host: e.target.value }))}
                      />
                    </FormGroup>

                    <FormGroup>
                      <Label>SMTP Port</Label>
                      <Input
                        type="number"
                        min="1"
                        value={settings.email_smtp_port}
                        onChange={e =>
                          setSettings(prev => ({
                            ...prev,
                            email_smtp_port: Number.parseInt(e.target.value, 10) || 0,
                          }))
                        }
                      />
                    </FormGroup>

                    <FormGroup>
                      <Label>SMTP Username</Label>
                      <Input
                        type="text"
                        value={settings.email_smtp_user}
                        onChange={e => setSettings(prev => ({ ...prev, email_smtp_user: e.target.value }))}
                      />
                    </FormGroup>

                    <FormGroup>
                      <Label>SMTP Password</Label>
                      <Input
                        type="password"
                        value={settings.email_smtp_password}
                        placeholder={hasStoredSensitive.smtpPassword ? 'Configured' : ''}
                        onChange={e => setSettings(prev => ({ ...prev, email_smtp_password: e.target.value }))}
                      />
                      <HelpText>留空表示不修改现有值</HelpText>
                    </FormGroup>
                  </>
                )}

                {settings.email_provider === 'resend' && (
                  <FormGroup>
                    <Label>Resend API Key</Label>
                    <Input
                      type="password"
                      value={settings.email_resend_api_key}
                      placeholder={hasStoredSensitive.resendApiKey ? 'Configured' : ''}
                      onChange={e => setSettings(prev => ({ ...prev, email_resend_api_key: e.target.value }))}
                    />
                    <HelpText>留空表示不修改现有值</HelpText>
                  </FormGroup>
                )}

                <FormGroup>
                  <Label>Sender Email</Label>
                  <Input
                    type="email"
                    value={settings.email_from}
                    onChange={e => setSettings(prev => ({ ...prev, email_from: e.target.value }))}
                  />
                </FormGroup>

                <FormGroup>
                  <Label>Sender Name</Label>
                  <Input
                    type="text"
                    value={settings.email_from_name}
                    onChange={e => setSettings(prev => ({ ...prev, email_from_name: e.target.value }))}
                  />
                </FormGroup>

                <FormGroup>
                  <Label>Test Email</Label>
                  <InlineActions>
                    <InlineField>
                      <Input
                        type="email"
                        value={testEmail}
                        placeholder="you@example.com"
                        onChange={e => setTestEmail(e.target.value)}
                      />
                    </InlineField>
                    <SecondaryButton type="button" onClick={handleSendTestEmail} disabled={testingEmail}>
                      {testingEmail ? 'Sending...' : 'Send Test Email'}
                    </SecondaryButton>
                  </InlineActions>
                  <HelpText>Save settings first, then test delivery with the email above.</HelpText>
                </FormGroup>
              </>
            )}
          </CardBody>
        )}
      </Card>

      <Card>
        <CardToggle type="button" onClick={() => toggleSection('ai')}>
          <CardTitle>AI Settings</CardTitle>
          <ToggleIcon expanded={expanded.ai}>▼</ToggleIcon>
        </CardToggle>
        {expanded.ai && (
          <CardBody>
            <FormGroup>
              <Label>Default Memory Model</Label>
              <Select
                value={settings.ai_default_memory_model}
                onChange={e => {
                  setSettings(prev => ({ ...prev, ai_default_memory_model: e.target.value }));
                  setCustomModelIdentifier('');
                }}
              >
                {memoryModelOptions.map(m => (
                  <option key={m} value={m}>
                    {availableModels.find(am => am.model_identifier === m)?.display_name || m}
                  </option>
                ))}
              </Select>
            </FormGroup>
            <FormGroup>
              <Label>Custom Model Identifier (override)</Label>
              <Input
                type="text"
                value={customModelIdentifier}
                placeholder="Leave blank to use selection above"
                onChange={e => {
                  setCustomModelIdentifier(e.target.value);
                  if (e.target.value.trim()) {
                    setSettings(prev => ({ ...prev, ai_default_memory_model: e.target.value.trim() }));
                  }
                }}
              />
              <HelpText>Enter a model identifier manually to override the dropdown selection.</HelpText>
            </FormGroup>

            <FormGroup>
              <SwitchLabel>
                <Checkbox
                  type="checkbox"
                  checked={settings.ai_memory_extraction_enabled}
                  onChange={e =>
                    setSettings(prev => ({
                      ...prev,
                      ai_memory_extraction_enabled: e.target.checked,
                    }))
                  }
                />
                Enable Memory Extraction
              </SwitchLabel>
            </FormGroup>
          </CardBody>
        )}
      </Card>

      <Button type="button" onClick={handleSave} disabled={loading}>
        {loading ? 'Saving...' : 'Save Settings'}
      </Button>
    </div>
  );
};

export default SystemSettingsPage;

