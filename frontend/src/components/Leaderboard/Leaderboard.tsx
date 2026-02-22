import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../api/client';
import { ensureArray, ensureNumber } from '../../utils/safe';

const Page = styled.div`
  min-height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
`;

const Header = styled.div`
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  gap: 14px;
`;

const BackBtn = styled.button`
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: rgba(255,255,255,0.06);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.18s;
  flex-shrink: 0;
  &:hover { background: rgba(255,255,255,0.1); }
`;

const PageTitle = styled.h1`
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  flex: 1;
`;

const RefreshBtn = styled.button`
  padding: 8px 18px;
  border: none;
  border-radius: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  &:hover { opacity: 0.88; }
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

const Content = styled.div`
  max-width: 760px;
  margin: 0 auto;
  padding: 28px 20px;
`;

const MetaRow = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
`;

const LastUpdated = styled.span`
  font-size: 12px;
  color: var(--text-muted);
`;

const TableWrap = styled.div`
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 18px;
  overflow: hidden;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
`;

const Th = styled.th`
  padding: 14px 18px;
  text-align: left;
  background: var(--bg-primary);
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
`;

const Td = styled.td`
  padding: 14px 18px;
  border-top: 1px solid var(--border-primary);
  color: var(--text-primary);
  font-size: 14px;
`;

const Rank = styled.div<{ $rank: number }>`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
  background: ${p => {
    if (p.$rank === 1) return 'linear-gradient(135deg, #f6d365 0%, #fda085 100%)';
    if (p.$rank === 2) return 'linear-gradient(135deg, #c2e9fb 0%, #a1c4fd 100%)';
    if (p.$rank === 3) return 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)';
    return 'rgba(255,255,255,0.07)';
  }};
  color: ${p => p.$rank <= 3 ? '#1a202c' : 'var(--text-secondary)'};
`;

const BarWrap = styled.div`
  width: 100%;
  height: 6px;
  background: var(--border-primary);
  border-radius: 4px;
  overflow: hidden;
`;

const Bar = styled.div<{ $pct: number }>`
  height: 100%;
  width: ${p => p.$pct}%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.4s;
`;

interface TokenStat {
  user_id: string;
  username: string;
  total_tokens: number;
  total_requests: number;
}

const Leaderboard: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<TokenStat[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const r = await apiClient.get('/statistics/tokens');
      setStats(ensureArray<TokenStat>(r.data?.leaderboard));
      setLastUpdated(new Date());
    } catch {
      setStats([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const max = Math.max(...stats.map(s => ensureNumber(s.total_tokens)), 1);
  const fmt = (n: number) => ensureNumber(n).toLocaleString();

  return (
    <Page>
      <Header>
        <BackBtn onClick={() => navigate('/chat')}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </BackBtn>
        <PageTitle>🏆 Token 排行榜</PageTitle>
        <RefreshBtn onClick={load} disabled={loading}>
          {loading ? '加载中…' : '刷新'}
        </RefreshBtn>
      </Header>

      <Content>
        {lastUpdated && (
          <MetaRow>
            <LastUpdated>更新于 {lastUpdated.toLocaleString('zh-CN')}</LastUpdated>
          </MetaRow>
        )}

        <TableWrap>
          <Table>
            <thead>
              <tr>
                <Th style={{ width: 56 }}>排名</Th>
                <Th>用户名</Th>
                <Th>总 Tokens</Th>
                <Th>请求次数</Th>
                <Th style={{ width: 160 }}>用量</Th>
              </tr>
            </thead>
            <tbody>
              {stats.map((s, i) => (
                <tr key={s.user_id}>
                  <Td><Rank $rank={i + 1}>{i + 1}</Rank></Td>
                  <Td style={{ fontWeight: 500 }}>{s.username}</Td>
                  <Td>{fmt(s.total_tokens)}</Td>
                  <Td>{fmt(s.total_requests)}</Td>
                  <Td>
                    <BarWrap>
                      <Bar $pct={(ensureNumber(s.total_tokens) / max) * 100} />
                    </BarWrap>
                  </Td>
                </tr>
              ))}
              {!loading && stats.length === 0 && (
                <tr>
                  <Td colSpan={5} style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 40 }}>
                    暂无数据
                  </Td>
                </tr>
              )}
            </tbody>
          </Table>
        </TableWrap>
      </Content>
    </Page>
  );
};

export default Leaderboard;
