import React, { useState, useEffect } from 'react';
import styled, { keyframes } from 'styled-components';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../api/client';

// ─── Animations ──────────────────────────────────────────────────────────────
const fadeIn = keyframes`
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
`;

// ─── Layout ───────────────────────────────────────────────────────────────────
const Page = styled.div`
  min-height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
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

const AddBtn = styled.button`
  padding: 9px 20px;
  border: none;
  border-radius: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;

  &:hover {
    opacity: 0.9;
    box-shadow: 0 4px 14px rgba(102,126,234,0.3);
    transform: translateY(-1px);
  }
`;

const Content = styled.div`
  flex: 1;
  max-width: 800px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 20px;
`;

// ─── Filters ──────────────────────────────────────────────────────────────────
const FilterRow = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
`;

const FilterChip = styled.button<{ $active?: boolean }>`
  padding: 6px 16px;
  border-radius: 20px;
  border: 1.5px solid ${p => p.$active ? '#667eea' : 'var(--border-primary)'};
  background: ${p => p.$active ? 'rgba(102,126,234,0.12)' : 'transparent'};
  color: ${p => p.$active ? '#667eea' : 'var(--text-secondary)'};
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.18s;

  &:hover {
    border-color: #667eea;
    color: #667eea;
  }
`;

// ─── Memory Card ──────────────────────────────────────────────────────────────
const CardGrid = styled.div`
  display: flex;
  flex-direction: column;
  gap: 10px;
`;

const Card = styled.div`
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 16px;
  padding: 16px 18px;
  animation: ${fadeIn} 0.2s ease-out both;
  transition: border-color 0.18s, box-shadow 0.18s;

  &:hover {
    border-color: rgba(102,126,234,0.3);
    box-shadow: 0 2px 16px rgba(0,0,0,0.12);
  }
`;

const CardTop = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 8px;
`;

const CategoryBadge = styled.span<{ $cat: string }>`
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
  ${p => {
    if (p.$cat === 'preference') return 'background: rgba(102,126,234,0.15); color: #667eea;';
    if (p.$cat === 'fact')       return 'background: rgba(52,199,89,0.15); color: #34c759;';
    if (p.$cat === 'context')    return 'background: rgba(255,159,10,0.15); color: #ff9f0a;';
    return 'background: rgba(255,255,255,0.08); color: var(--text-muted);';
  }}
`;

const ImportanceStars = styled.span`
  font-size: 12px;
  color: var(--text-muted);
  margin-left: auto;
  flex-shrink: 0;
`;

const CardContent = styled.div`
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
`;

const CardActions = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 12px;
`;

const ActionBtn = styled.button<{ $danger?: boolean }>`
  padding: 5px 14px;
  border-radius: 10px;
  border: 1px solid ${p => p.$danger ? 'rgba(255,69,58,0.3)' : 'var(--border-primary)'};
  background: transparent;
  color: ${p => p.$danger ? '#ff453a' : 'var(--text-secondary)'};
  font-size: 12px;
  cursor: pointer;
  transition: all 0.18s;

  &:hover {
    background: ${p => p.$danger ? 'rgba(255,69,58,0.1)' : 'rgba(255,255,255,0.05)'};
    border-color: ${p => p.$danger ? '#ff453a' : '#667eea'};
    color: ${p => p.$danger ? '#ff453a' : '#667eea'};
  }
`;

// ─── Modal ────────────────────────────────────────────────────────────────────
const Backdrop = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
`;

const Modal = styled.div`
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 20px;
  padding: 28px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.4);
  animation: ${fadeIn} 0.22s ease-out both;
`;

const ModalTitle = styled.h2`
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 20px;
  color: var(--text-primary);
`;

const Label = styled.label`
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
`;

const Textarea = styled.textarea`
  width: 100%;
  padding: 12px 14px;
  background: var(--bg-primary);
  border: 1.5px solid var(--border-primary);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  min-height: 90px;
  transition: border-color 0.18s;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
  &::placeholder { color: var(--text-muted); }
`;

const Select = styled.select`
  width: 100%;
  padding: 10px 14px;
  background: var(--bg-primary);
  border: 1.5px solid var(--border-primary);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 14px;
  transition: border-color 0.18s;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
  option { background: var(--bg-primary); }
`;

const RangeRow = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`;

const Range = styled.input`
  flex: 1;
  accent-color: #667eea;
`;

const RangeValue = styled.span`
  font-size: 14px;
  font-weight: 700;
  color: #667eea;
  min-width: 20px;
  text-align: center;
`;

const FormGroup = styled.div`
  margin-bottom: 16px;
`;

const ModalActions = styled.div`
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 22px;
`;

const CancelBtn = styled.button`
  padding: 10px 20px;
  border-radius: 12px;
  border: 1px solid var(--border-primary);
  background: transparent;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.18s;

  &:hover { background: rgba(255,255,255,0.06); color: var(--text-primary); }
`;

const SaveBtn = styled.button`
  padding: 10px 24px;
  border-radius: 12px;
  border: none;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;

  &:hover { opacity: 0.9; box-shadow: 0 4px 14px rgba(102,126,234,0.3); }
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

// ─── Empty State ──────────────────────────────────────────────────────────────
const Empty = styled.div`
  text-align: center;
  padding: 60px 20px;
  color: var(--text-muted);
`;

const EmptyIcon = styled.div`
  font-size: 48px;
  margin-bottom: 12px;
`;

// ─── Types ────────────────────────────────────────────────────────────────────
type Category = 'preference' | 'fact' | 'context';

interface MemoryItem {
  id: string;
  content: string;
  category: Category;
  importance: number;
  created_at?: string;
}

const CATEGORY_LABELS: Record<Category | 'all', string> = {
  all:        '全部',
  preference: '偏好',
  fact:       '事实',
  context:    '上下文',
};

const stars = (n: number) => '★'.repeat(n) + '☆'.repeat(10 - n);

// ─── Component ────────────────────────────────────────────────────────────────
const Memory: React.FC = () => {
  const navigate = useNavigate();
  const [memories, setMemories] = useState<MemoryItem[]>([]);
  const [filter, setFilter] = useState<Category | 'all'>('all');
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<MemoryItem | null>(null);
  const [loading, setLoading] = useState(false);

  // Form state
  const [formContent, setFormContent] = useState('');
  const [formCategory, setFormCategory] = useState<Category>('preference');
  const [formImportance, setFormImportance] = useState(5);

  useEffect(() => { fetchMemories(); }, []);

  const fetchMemories = async () => {
    try {
      const r = await apiClient.memories.list();
      const list: MemoryItem[] = Array.isArray(r?.memories) ? r.memories : [];
      setMemories(list);
    } catch {
      setMemories([]);
    }
  };

  const openCreate = () => {
    setEditTarget(null);
    setFormContent('');
    setFormCategory('preference');
    setFormImportance(5);
    setShowModal(true);
  };

  const openEdit = (m: MemoryItem) => {
    setEditTarget(m);
    setFormContent(m.content);
    setFormCategory(m.category);
    setFormImportance(m.importance);
    setShowModal(true);
  };

  const handleSave = async () => {
    if (!formContent.trim()) return;
    setLoading(true);
    try {
      if (editTarget) {
        await apiClient.memories.update(editTarget.id, formContent.trim(), formCategory, formImportance);
      } else {
        await apiClient.memories.create(formContent.trim(), formCategory, formImportance);
      }
      setShowModal(false);
      await fetchMemories();
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('确认删除这条记忆？')) return;
    try {
      await apiClient.memories.delete(id);
      setMemories(prev => prev.filter(m => m.id !== id));
    } catch (e) {
      console.error(e);
    }
  };

  const filtered = filter === 'all' ? memories : memories.filter(m => m.category === filter);

  return (
    <Page>
      <Header>
        <BackBtn onClick={() => navigate('/chat')}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </BackBtn>
        <PageTitle>🧠 记忆管理</PageTitle>
        <AddBtn onClick={openCreate}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          添加记忆
        </AddBtn>
      </Header>

      <Content>
        {/* Category filter */}
        <FilterRow>
          {(['all', 'preference', 'fact', 'context'] as const).map(cat => (
            <FilterChip
              key={cat}
              $active={filter === cat}
              onClick={() => setFilter(cat)}
            >
              {CATEGORY_LABELS[cat]}
              {cat !== 'all' && (
                <> ({memories.filter(m => m.category === cat).length})</>
              )}
            </FilterChip>
          ))}
        </FilterRow>

        {/* Memory cards */}
        {filtered.length === 0 ? (
          <Empty>
            <EmptyIcon>🧠</EmptyIcon>
            <div style={{ fontSize: 15, marginBottom: 6 }}>
              {filter === 'all' ? '还没有记忆' : `没有「${CATEGORY_LABELS[filter]}」类型的记忆`}
            </div>
            <div style={{ fontSize: 13 }}>AI 会在对话中自动提取记忆，也可手动添加</div>
          </Empty>
        ) : (
          <CardGrid>
            {filtered.map(m => (
              <Card key={m.id}>
                <CardTop>
                  <CategoryBadge $cat={m.category}>{CATEGORY_LABELS[m.category]}</CategoryBadge>
                  <ImportanceStars title={`重要程度 ${m.importance}/10`}>
                    {stars(Math.min(Math.max(m.importance, 0), 10))}
                  </ImportanceStars>
                </CardTop>
                <CardContent>{m.content}</CardContent>
                <CardActions>
                  <ActionBtn onClick={() => openEdit(m)}>编辑</ActionBtn>
                  <ActionBtn $danger onClick={() => handleDelete(m.id)}>删除</ActionBtn>
                </CardActions>
              </Card>
            ))}
          </CardGrid>
        )}
      </Content>

      {/* Create / Edit modal */}
      {showModal && (
        <Backdrop onClick={e => { if (e.target === e.currentTarget) setShowModal(false); }}>
          <Modal>
            <ModalTitle>{editTarget ? '编辑记忆' : '添加记忆'}</ModalTitle>

            <FormGroup>
              <Label>内容</Label>
              <Textarea
                value={formContent}
                onChange={e => setFormContent(e.target.value)}
                placeholder="描述需要记住的信息…"
                autoFocus
              />
            </FormGroup>

            <FormGroup>
              <Label>类型</Label>
              <Select
                value={formCategory}
                onChange={e => setFormCategory(e.target.value as Category)}
              >
                <option value="preference">偏好 — 用户的喜好与习惯</option>
                <option value="fact">事实 — 关于用户的客观信息</option>
                <option value="context">上下文 — 对话背景与临时信息</option>
              </Select>
            </FormGroup>

            <FormGroup>
              <Label>重要程度 (1–10)</Label>
              <RangeRow>
                <Range
                  type="range"
                  min={1}
                  max={10}
                  value={formImportance}
                  onChange={e => setFormImportance(Number(e.target.value))}
                />
                <RangeValue>{formImportance}</RangeValue>
              </RangeRow>
            </FormGroup>

            <ModalActions>
              <CancelBtn onClick={() => setShowModal(false)}>取消</CancelBtn>
              <SaveBtn onClick={handleSave} disabled={loading || !formContent.trim()}>
                {loading ? '保存中…' : '保存'}
              </SaveBtn>
            </ModalActions>
          </Modal>
        </Backdrop>
      )}
    </Page>
  );
};

export default Memory;
