import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';

const Card = styled.div`
  background: #1a2332;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid #2d3748;
`;

const CardTitle = styled.h2`
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #e8eaed;
`;

const FormGroup = styled.div`
  margin-bottom: 20px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  color: #a0aec0;
  font-size: 14px;
  font-weight: 500;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px 16px;
  background: #0f1419;
  border: 1px solid #2d3748;
  border-radius: 8px;
  color: #e8eaed;
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

const Message = styled.div<{ type?: 'success' | 'error' }>`
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  background: ${props => props.type === 'error' ? 'rgba(252, 129, 129, 0.1)' : 'rgba(72, 187, 120, 0.1)'};
  color: ${props => props.type === 'error' ? '#fc8181' : '#48bb78'};
  border: 1px solid ${props => props.type === 'error' ? '#fc8181' : '#48bb78'};
`;

const HelpText = styled.p`
  color: #718096;
  font-size: 13px;
  margin: 8px 0 0 0;
`;

interface SystemSettings {
  rate_limit_default_per_minute: number;
  system_name: string;
  maintenance_mode: boolean;
}

const SystemSettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<SystemSettings>({
    rate_limit_default_per_minute: 20,
    system_name: 'AI Chat System',
    maintenance_mode: false,
  });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const response = await apiClient.get('/admin/system/settings');
      setSettings(response.data);
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    setMessage(null);

    try {
      await apiClient.put('/admin/system/settings', settings);
      setMessage({ text: '设置保存成功！', type: 'success' });
      setTimeout(() => setMessage(null), 3000);
    } catch (error: any) {
      setMessage({
        text: error.response?.data?.error || '保存失败，请重试',
        type: 'error'
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      {message && <Message type={message.type}>{message.text}</Message>}

      <Card>
        <CardTitle>速率限制设置</CardTitle>
        <FormGroup>
          <Label>默认速率限制（每分钟消息数）</Label>
          <Input
            type="number"
            min="1"
            max="1000"
            value={settings.rate_limit_default_per_minute}
            onChange={(e) => setSettings({
              ...settings,
              rate_limit_default_per_minute: parseInt(e.target.value) || 20
            })}
          />
          <HelpText>
            此设置应用于所有用户的默认速率限制。您可以在"用户管理"中为特定用户设置自定义限制。
          </HelpText>
        </FormGroup>
      </Card>

      <Card>
        <CardTitle>系统信息</CardTitle>
        <FormGroup>
          <Label>系统名称</Label>
          <Input
            type="text"
            value={settings.system_name}
            onChange={(e) => setSettings({
              ...settings,
              system_name: e.target.value
            })}
          />
        </FormGroup>

        <FormGroup>
          <Label>
            <input
              type="checkbox"
              checked={settings.maintenance_mode}
              onChange={(e) => setSettings({
                ...settings,
                maintenance_mode: e.target.checked
              })}
              style={{ marginRight: '8px' }}
            />
            维护模式
          </Label>
          <HelpText>
            启用后，非管理员用户将无法访问系统
          </HelpText>
        </FormGroup>
      </Card>

      <Button onClick={handleSave} disabled={loading}>
        {loading ? '保存中...' : '保存设置'}
      </Button>
    </div>
  );
};

export default SystemSettingsPage;
