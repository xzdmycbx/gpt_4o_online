import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import useAutoRefresh from '../../hooks/useAutoRefresh';

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  background: #1a2332;
  border-radius: 12px;
  overflow: hidden;
`;

const Th = styled.th`
  padding: 16px;
  text-align: left;
  background: #0f1419;
  color: #a0aec0;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
`;

const Td = styled.td`
  padding: 16px;
  border-top: 1px solid #2d3748;
  color: #e8eaed;
`;

const Rank = styled.div<{ rank: number }>`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  background: ${props => {
    if (props.rank === 1) return 'linear-gradient(135deg, #f6d365 0%, #fda085 100%)';
    if (props.rank === 2) return 'linear-gradient(135deg, #c2e9fb 0%, #a1c4fd 100%)';
    if (props.rank === 3) return 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)';
    return '#4a5568';
  }};
  color: ${props => props.rank <= 3 ? '#1a202c' : '#e8eaed'};
`;

const ProgressBar = styled.div`
  width: 100%;
  height: 8px;
  background: #0f1419;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
`;

const Progress = styled.div<{ width: number }>`
  height: 100%;
  width: ${props => props.width}%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s;
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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

const LastUpdated = styled.div`
  color: #a0aec0;
  font-size: 13px;
`;

interface TokenStat {
  user_id: string;
  username: string;
  total_tokens: number;
  total_requests: number;
}

const TokenLeaderboard: React.FC = () => {
  const [stats, setStats] = useState<TokenStat[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const loadStats = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/admin/statistics/tokens');
      // Backend returns {leaderboard: [...]} but handle both formats
      setStats(response.data.leaderboard || response.data || []);
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to load token stats:', error);
    } finally {
      setLoading(false);
    }
  };

  // Load initial data
  useEffect(() => {
    loadStats();
  }, []);

  // Auto-refresh every 60 seconds
  useAutoRefresh(loadStats, 60000);

  const maxTokens = Math.max(...stats.map(s => s.total_tokens), 1);

  const formatNumber = (num: number) => {
    return num.toLocaleString();
  };

  if (loading && !stats.length) {
    return <div>加载中...</div>;
  }

  return (
    <div>
      <Header>
        <div>
          {lastUpdated && (
            <LastUpdated>
              最后更新: {lastUpdated.toLocaleString('zh-CN')}
            </LastUpdated>
          )}
        </div>
        <RefreshButton onClick={loadStats} disabled={loading}>
          {loading ? '刷新中...' : '刷新数据'}
        </RefreshButton>
      </Header>
      <Table>
      <thead>
        <tr>
          <Th style={{ width: '60px' }}>排名</Th>
          <Th>用户名</Th>
          <Th>总 Tokens</Th>
          <Th>请求次数</Th>
          <Th style={{ width: '200px' }}>使用率</Th>
        </tr>
      </thead>
      <tbody>
        {stats.map((stat, index) => (
          <tr key={stat.user_id}>
            <Td>
              <Rank rank={index + 1}>{index + 1}</Rank>
            </Td>
            <Td>{stat.username}</Td>
            <Td>{formatNumber(stat.total_tokens)}</Td>
            <Td>{formatNumber(stat.total_requests)}</Td>
            <Td>
              <ProgressBar>
                <Progress width={(stat.total_tokens / maxTokens) * 100} />
              </ProgressBar>
            </Td>
          </tr>
        ))}
      </tbody>
    </Table>
    </div>
  );
};

export default TokenLeaderboard;
