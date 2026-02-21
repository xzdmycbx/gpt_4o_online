import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import useAutoRefresh from '../../hooks/useAutoRefresh';
import { ensureArray, ensureNumber } from '../../utils/safe';

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  background: var(--bg-secondary);
  border-radius: 12px;
  overflow: hidden;
`;

const Th = styled.th`
  padding: 16px;
  text-align: left;
  background: var(--bg-elevated);
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
`;

const Td = styled.td`
  padding: 16px;
  border-top: 1px solid var(--border-primary);
  color: var(--text-primary);
  font-size: 14px;
`;

const Badge = styled.span<{ action: string }>`
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  background: ${props => {
    if (props.action.includes('delete') || props.action.includes('ban')) {
      return 'rgba(252, 129, 129, 0.2)';
    }
    if (props.action.includes('create') || props.action.includes('add')) {
      return 'rgba(72, 187, 120, 0.2)';
    }
    return 'rgba(102, 126, 234, 0.2)';
  }};
  color: ${props => {
    if (props.action.includes('delete') || props.action.includes('ban')) {
      return '#fc8181';
    }
    if (props.action.includes('create') || props.action.includes('add')) {
      return '#48bb78';
    }
    return '#667eea';
  }};
`;

const Pagination = styled.div`
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
`;

const PageButton = styled.button<{ active?: boolean }>`
  padding: 8px 16px;
  background: ${props => props.active ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : 'var(--bg-secondary)'};
  color: ${props => props.active ? 'white' : 'var(--text-secondary)'};
  border: 1px solid ${props => props.active ? 'transparent' : 'var(--border-primary)'};
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: ${props => props.active ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#2a3441'};
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
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
  color: var(--text-secondary);
  font-size: 13px;
`;

interface AuditLog {
  id: string;
  user_id: string;
  username: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  ip_address?: string;
  created_at: string;
}

const AuditLogs: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const pageSize = 20;

  const loadLogs = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/admin/audit-logs', {
        params: {
          limit: pageSize,
          offset: (page - 1) * pageSize,
        },
      });
      setLogs(ensureArray<AuditLog>(response.data?.logs));
      setTotal(ensureNumber(response.data?.total));
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to load audit logs:', error);
      setLogs([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  // Load data when page changes
  useEffect(() => {
    loadLogs();
  }, [page]);

  // Auto-refresh every 30 seconds
  useAutoRefresh(loadLogs, 30000);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const totalPages = Math.ceil(total / pageSize);

  if (loading && !logs.length) {
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
        <RefreshButton onClick={loadLogs} disabled={loading}>
          {loading ? '刷新中...' : '刷新数据'}
        </RefreshButton>
      </Header>
      <Table>
        <thead>
          <tr>
            <Th>时间</Th>
            <Th>用户</Th>
            <Th>操作</Th>
            <Th>资源类型</Th>
            <Th>IP 地址</Th>
          </tr>
        </thead>
        <tbody>
          {logs.map(log => (
            <tr key={log.id}>
              <Td>{formatDate(log.created_at)}</Td>
              <Td>{log.username}</Td>
              <Td>
                <Badge action={log.action}>{log.action}</Badge>
              </Td>
              <Td>{log.resource_type}</Td>
              <Td>{log.ip_address || '-'}</Td>
            </tr>
          ))}
        </tbody>
      </Table>

      {totalPages > 1 && (
        <Pagination>
          <PageButton
            disabled={page === 1}
            onClick={() => setPage(page - 1)}
          >
            上一页
          </PageButton>
          <PageButton active>
            {page} / {totalPages}
          </PageButton>
          <PageButton
            disabled={page === totalPages}
            onClick={() => setPage(page + 1)}
          >
            下一页
          </PageButton>
        </Pagination>
      )}
    </div>
  );
};

export default AuditLogs;

