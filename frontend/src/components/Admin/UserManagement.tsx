import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';

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

const Button = styled.button<{ variant?: 'primary' | 'danger' | 'secondary' }>`
  padding: 8px 16px;
  background: ${props => {
    if (props.variant === 'danger') return '#fc8181';
    if (props.variant === 'secondary') return '#4a5568';
    return 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
  }};
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  margin-right: 8px;
  transition: all 0.2s;

  &:hover {
    transform: translateY(-1px);
    opacity: 0.9;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
`;

const Badge = styled.span<{ type: 'admin' | 'super_admin' | 'user' | 'banned' }>`
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  background: ${props => {
    if (props.type === 'super_admin') return 'rgba(245, 101, 101, 0.2)';
    if (props.type === 'admin') return 'rgba(102, 126, 234, 0.2)';
    if (props.type === 'banned') return 'rgba(203, 166, 247, 0.2)';
    return 'rgba(72, 187, 120, 0.2)';
  }};
  color: ${props => {
    if (props.type === 'super_admin') return '#f56565';
    if (props.type === 'admin') return '#667eea';
    if (props.type === 'banned') return '#cba6f7';
    return '#48bb78';
  }};
`;

const Modal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background: #1a2332;
  border-radius: 12px;
  padding: 32px;
  max-width: 500px;
  width: 90%;
  border: 1px solid #2d3748;
`;

const ModalTitle = styled.h3`
  margin: 0 0 20px 0;
  font-size: 20px;
  color: #e8eaed;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px 16px;
  background: #0f1419;
  border: 1px solid #2d3748;
  border-radius: 8px;
  color: #e8eaed;
  font-size: 14px;
  margin-bottom: 16px;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  color: #a0aec0;
  font-size: 14px;
`;

interface User {
  id: string;
  username: string;
  email?: string;
  role: string;
  is_banned: boolean;
  custom_rate_limit?: number;
  rate_limit_exempt: boolean;
  created_at: string;
}

const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [showLimitModal, setShowLimitModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [customLimit, setCustomLimit] = useState<string>('');
  const [isExempt, setIsExempt] = useState(false);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/admin/users');
      setUsers(response.data.users || []);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleBan = async (userId: string, isBanned: boolean) => {
    try {
      if (isBanned) {
        // Unban
        await apiClient.put(`/admin/users/${userId}/unban`);
      } else {
        // Ban - ask for reason
        const reason = prompt('请输入封禁原因：');
        if (!reason) return; // Cancelled
        await apiClient.put(`/admin/users/${userId}/ban`, { reason });
      }
      await loadUsers();
    } catch (error) {
      console.error('Failed to ban/unban user:', error);
    }
  };

  const handleOpenLimitModal = (user: User) => {
    setSelectedUser(user);
    setCustomLimit(user.custom_rate_limit?.toString() || '');
    setIsExempt(user.rate_limit_exempt);
    setShowLimitModal(true);
  };

  const handleSaveLimit = async () => {
    if (!selectedUser) return;

    try {
      const limit = customLimit ? parseInt(customLimit) : null;
      await apiClient.put(`/admin/users/${selectedUser.id}/rate-limit`, {
        limit: limit,
        exempt: isExempt,
      });
      await loadUsers();
      setShowLimitModal(false);
    } catch (error) {
      console.error('Failed to set rate limit:', error);
    }
  };

  const getRoleLabel = (role: string) => {
    const labels: Record<string, string> = {
      super_admin: '超级管理员',
      admin: '管理员',
      user: '用户',
    };
    return labels[role] || role;
  };

  return (
    <div>
      {loading ? (
        <div>加载中...</div>
      ) : (
        <Table>
          <thead>
            <tr>
              <Th>用户名</Th>
              <Th>邮箱</Th>
              <Th>角色</Th>
              <Th>速率限制</Th>
              <Th>状态</Th>
              <Th>操作</Th>
            </tr>
          </thead>
          <tbody>
            {users.map(user => (
              <tr key={user.id}>
                <Td>{user.username}</Td>
                <Td>{user.email || '-'}</Td>
                <Td>
                  <Badge type={user.role as any}>{getRoleLabel(user.role)}</Badge>
                </Td>
                <Td>
                  {user.rate_limit_exempt ? (
                    <Badge type="admin">无限制</Badge>
                  ) : user.custom_rate_limit ? (
                    `${user.custom_rate_limit}/分钟`
                  ) : (
                    '默认'
                  )}
                </Td>
                <Td>
                  {user.is_banned ? (
                    <Badge type="banned">已封禁</Badge>
                  ) : (
                    <Badge type="user">正常</Badge>
                  )}
                </Td>
                <Td>
                  <Button
                    variant="secondary"
                    onClick={() => handleOpenLimitModal(user)}
                  >
                    设置限制
                  </Button>
                  <Button
                    variant={user.is_banned ? 'primary' : 'danger'}
                    onClick={() => handleBan(user.id, user.is_banned)}
                  >
                    {user.is_banned ? '解封' : '封禁'}
                  </Button>
                </Td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      {showLimitModal && (
        <Modal onClick={() => setShowLimitModal(false)}>
          <ModalContent onClick={(e) => e.stopPropagation()}>
            <ModalTitle>设置用户速率限制</ModalTitle>
            <Label>用户：{selectedUser?.username}</Label>
            <Label>自定义限制（每分钟消息数）</Label>
            <Input
              type="number"
              placeholder="留空使用默认限制"
              value={customLimit}
              onChange={(e) => setCustomLimit(e.target.value)}
            />
            <Label>
              <input
                type="checkbox"
                checked={isExempt}
                onChange={(e) => setIsExempt(e.target.checked)}
                style={{ marginRight: '8px' }}
              />
              免除速率限制（无限制）
            </Label>
            <div style={{ marginTop: '24px', display: 'flex', gap: '12px' }}>
              <Button onClick={handleSaveLimit}>保存</Button>
              <Button variant="secondary" onClick={() => setShowLimitModal(false)}>
                取消
              </Button>
            </div>
          </ModalContent>
        </Modal>
      )}
    </div>
  );
};

export default UserManagement;
