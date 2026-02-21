import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import useAutoRefresh from '../../hooks/useAutoRefresh';
import { ensureNumber } from '../../utils/safe';

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
`;

const Card = styled.div`
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 24px;
  border: 1px solid var(--border-primary);
  transition: all 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }
`;

const CardTitle = styled.div`
  font-size: 13px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
`;

const CardValue = styled.div`
  font-size: 32px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
`;

const CardChange = styled.div<{ positive?: boolean }>`
  font-size: 13px;
  color: ${props => props.positive ? '#48bb78' : '#fc8181'};
  display: flex;
  align-items: center;
  gap: 4px;
`;

const Section = styled.div`
  margin-bottom: 32px;
`;

const SectionTitle = styled.h2`
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
`;

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-primary);
  color: var(--text-secondary);

  &:last-child {
    border-bottom: none;
  }
`;

const InfoLabel = styled.span`
  font-weight: 500;
`;

const InfoValue = styled.span`
  color: var(--text-primary);
`;

const RefreshButton = styled.button`
  padding: 10px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 24px;

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

interface SystemStats {
  total_users: number;
  total_conversations: number;
  total_messages: number;
  total_tokens_used: number;
  active_users_today: number;
  active_users_week: number;
  messages_today: number;
  messages_week: number;
  average_tokens_per_message: number;
  system_uptime: string;
}

const EMPTY_STATS: SystemStats = {
  total_users: 0,
  total_conversations: 0,
  total_messages: 0,
  total_tokens_used: 0,
  active_users_today: 0,
  active_users_week: 0,
  messages_today: 0,
  messages_week: 0,
  average_tokens_per_message: 0,
  system_uptime: '0分钟',
};

const normalizeStats = (value: unknown): SystemStats => {
  const raw = typeof value === 'object' && value !== null ? (value as Record<string, unknown>) : {};
  return {
    total_users: ensureNumber(raw.total_users),
    total_conversations: ensureNumber(raw.total_conversations),
    total_messages: ensureNumber(raw.total_messages),
    total_tokens_used: ensureNumber(raw.total_tokens_used),
    active_users_today: ensureNumber(raw.active_users_today),
    active_users_week: ensureNumber(raw.active_users_week),
    messages_today: ensureNumber(raw.messages_today),
    messages_week: ensureNumber(raw.messages_week),
    average_tokens_per_message: ensureNumber(raw.average_tokens_per_message),
    system_uptime: typeof raw.system_uptime === 'string' && raw.system_uptime ? raw.system_uptime : EMPTY_STATS.system_uptime,
  };
};

const SystemOverview: React.FC = () => {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const loadStats = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/admin/statistics/overview');
      setStats(normalizeStats(response.data));
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to load statistics:', error);
    } finally {
      setLoading(false);
    }
  };

  // Load initial data
  useEffect(() => {
    loadStats();
  }, []);

  // Auto-refresh every 30 seconds
  useAutoRefresh(loadStats, 30000);

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  if (!stats && loading) {
    return <div>加载中...</div>;
  }

  if (!stats) {
    return <div>无法加载系统统计数据</div>;
  }

  return (
    <div>
      <RefreshButton onClick={loadStats} disabled={loading}>
        {loading ? '刷新中...' : '刷新数据'}
      </RefreshButton>
      {lastUpdated && (
        <div style={{ color: 'var(--text-secondary)', fontSize: '13px', marginBottom: '24px' }}>
          最后更新: {lastUpdated.toLocaleString('zh-CN')}
        </div>
      )}

      <Grid>
        <Card>
          <CardTitle>总用户数</CardTitle>
          <CardValue>{formatNumber(stats.total_users)}</CardValue>
          <CardChange positive>
            今日活跃: {stats.active_users_today}
          </CardChange>
        </Card>

        <Card>
          <CardTitle>总对话数</CardTitle>
          <CardValue>{formatNumber(stats.total_conversations)}</CardValue>
        </Card>

        <Card>
          <CardTitle>总消息数</CardTitle>
          <CardValue>{formatNumber(stats.total_messages)}</CardValue>
          <CardChange positive>
            今日: {stats.messages_today}
          </CardChange>
        </Card>

        <Card>
          <CardTitle>总Token使用</CardTitle>
          <CardValue>{formatNumber(stats.total_tokens_used)}</CardValue>
          <CardChange>
            平均: {stats.average_tokens_per_message.toFixed(0)}/消息
          </CardChange>
        </Card>
      </Grid>

      <Section>
        <SectionTitle>活跃度统计</SectionTitle>
        <Card>
          <InfoRow>
            <InfoLabel>今日活跃用户</InfoLabel>
            <InfoValue>{stats.active_users_today}</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>本周活跃用户</InfoLabel>
            <InfoValue>{stats.active_users_week}</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>今日消息数</InfoLabel>
            <InfoValue>{stats.messages_today}</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>本周消息数</InfoLabel>
            <InfoValue>{stats.messages_week}</InfoValue>
          </InfoRow>
        </Card>
      </Section>

      <Section>
        <SectionTitle>系统信息</SectionTitle>
        <Card>
          <InfoRow>
            <InfoLabel>系统运行时间</InfoLabel>
            <InfoValue>{stats.system_uptime}</InfoValue>
          </InfoRow>
          <InfoRow>
            <InfoLabel>平均Token/消息</InfoLabel>
            <InfoValue>{stats.average_tokens_per_message.toFixed(2)}</InfoValue>
          </InfoRow>
        </Card>
      </Section>
    </div>
  );
};

export default SystemOverview;

