import React, { useState } from 'react';
import styled from 'styled-components';
import SystemOverview from './SystemOverview';
import SystemSettings from './SystemSettings';
import UserManagement from './UserManagement';
import ModelManagement from './ModelManagement';
import TokenLeaderboard from './TokenLeaderboard';
import AuditLogs from './AuditLogs';

const Container = styled.div`
  display: flex;
  height: 100vh;
  background: #0e1621;
  color: #e8eaed;
`;

const Sidebar = styled.div<{ isOpen?: boolean }>`
  width: 260px;
  background: linear-gradient(180deg, #1a2332 0%, #0f1419 100%);
  border-right: 1px solid #2d3748;
  display: flex;
  flex-direction: column;
  padding: 24px 0;

  @media (max-width: 768px) {
    width: 100%;
    position: fixed;
    height: 100%;
    z-index: 10;
    transform: translateX(${props => props.isOpen ? '0' : '-100%'});
    transition: transform 0.3s ease-in-out;
  }
`;

const Logo = styled.div`
  font-size: 20px;
  font-weight: 600;
  padding: 0 24px 24px;
  border-bottom: 1px solid #2d3748;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const Nav = styled.nav`
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
`;

const NavItem = styled.div<{ active?: boolean }>`
  padding: 12px 24px;
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid ${props => props.active ? '#667eea' : 'transparent'};
  background: ${props => props.active ? 'rgba(102, 126, 234, 0.1)' : 'transparent'};
  color: ${props => props.active ? '#667eea' : '#a0aec0'};
  font-weight: ${props => props.active ? '500' : '400'};

  &:hover {
    background: rgba(102, 126, 234, 0.05);
    color: #667eea;
  }
`;

const Content = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 32px;
`;

const Header = styled.div`
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 1px solid #2d3748;
`;

const Title = styled.h1`
  font-size: 28px;
  font-weight: 600;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const Subtitle = styled.p`
  color: #718096;
  margin: 0;
  font-size: 14px;
`;

type TabType = 'overview' | 'settings' | 'users' | 'models' | 'tokens' | 'audit';

const Admin: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabType>('overview');
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const tabs = [
    { id: 'overview' as TabType, label: '系统概览', icon: '📈' },
    { id: 'settings' as TabType, label: '系统设置', icon: '⚙️' },
    { id: 'users' as TabType, label: '用户管理', icon: '👥' },
    { id: 'models' as TabType, label: '模型管理', icon: '🤖' },
    { id: 'tokens' as TabType, label: 'Token 排行榜', icon: '📊' },
    { id: 'audit' as TabType, label: '审计日志', icon: '📋' },
  ];

  const renderContent = () => {
    switch (activeTab) {
      case 'overview':
        return <SystemOverview />;
      case 'settings':
        return <SystemSettings />;
      case 'users':
        return <UserManagement />;
      case 'models':
        return <ModelManagement />;
      case 'tokens':
        return <TokenLeaderboard />;
      case 'audit':
        return <AuditLogs />;
      default:
        return null;
    }
  };

  const getTitle = () => {
    const tab = tabs.find(t => t.id === activeTab);
    return tab ? tab.label : '';
  };

  const getSubtitle = () => {
    const subtitles: Record<TabType, string> = {
      overview: '查看系统整体运行状态和统计数据',
      settings: '配置系统全局设置，包括默认速率限制等',
      users: '管理用户账号，设置权限和速率限制',
      models: '配置 AI 模型和 API 密钥',
      tokens: '查看用户 Token 使用统计',
      audit: '查看系统操作审计日志',
    };
    return subtitles[activeTab];
  };

  return (
    <Container>
      <Sidebar isOpen={sidebarOpen}>
        <Logo>AI Chat 管理后台</Logo>
        <Nav>
          {tabs.map(tab => (
            <NavItem
              key={tab.id}
              active={activeTab === tab.id}
              onClick={() => {
                setActiveTab(tab.id);
                setSidebarOpen(false); // Close sidebar on mobile after selection
              }}
            >
              <span style={{ marginRight: '8px' }}>{tab.icon}</span>
              {tab.label}
            </NavItem>
          ))}
        </Nav>
      </Sidebar>
      <Content>
        <Header>
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              style={{
                display: 'none',
                background: 'transparent',
                border: 'none',
                color: '#e8eaed',
                fontSize: '24px',
                cursor: 'pointer',
                padding: '8px',
              }}
              className="mobile-menu-btn"
            >
              ☰
            </button>
            <div>
              <Title>{getTitle()}</Title>
              <Subtitle>{getSubtitle()}</Subtitle>
            </div>
          </div>
        </Header>
        {renderContent()}
      </Content>
      <style>{`
        @media (max-width: 768px) {
          .mobile-menu-btn {
            display: block !important;
          }
        }
      `}</style>
    </Container>
  );
};

export default Admin;
