import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import useAutoRefresh from '../../hooks/useAutoRefresh';

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
`;

const Card = styled.div`
  background: #1a2332;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #2d3748;
  transition: all 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }
`;

const CardTitle = styled.div`
  font-size: 13px;
  color: #a0aec0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
`;

const CardValue = styled.div`
  font-size: 32px;
  font-weight: 600;
  color: #e8eaed;
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
  color: #e8eaed;
  margin-bottom: 16px;
`;

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #2d3748;
  color: #a0aec0;

  &:last-child {
    border-bottom: none;
  }
`;

const InfoLabel = styled.span`
  font-weight: 500;
`;

const InfoValue = styled.span`
  color: #e8eaed;
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

const SystemOverview: React.FC = () => {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const loadStats = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/admin/statistics/overview');
      setStats(response.data);
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
        <div style={{ color: '#a0aec0', fontSize: '13px', marginBottom: '24px' }}>
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
