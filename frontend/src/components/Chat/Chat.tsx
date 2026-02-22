import React, { useState, useEffect, useRef } from 'react';
import styled, { keyframes } from 'styled-components';
import apiClient from '../../api/client';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useSettings } from '../../contexts/SettingsContext';

// ─── Animations ──────────────────────────────────────────────────────────────
const fadeIn = keyframes`
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
`;

const pulse = keyframes`
  0%, 80%, 100% { transform: scale(0); opacity: 0.4; }
  40%           { transform: scale(1); opacity: 1; }
`;

// ─── Layout ───────────────────────────────────────────────────────────────────
const Container = styled.div`
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  position: relative;
  overflow: hidden;
`;

// ─── Sidebar ──────────────────────────────────────────────────────────────────
const Sidebar = styled.aside<{ $open: boolean }>`
  width: 300px;
  min-width: 300px;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-primary);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(12px);

  @media (max-width: 768px) {
    position: absolute;
    inset: 0 auto 0 0;
    width: 280px;
    z-index: 30;
    transform: ${p => p.$open ? 'translateX(0)' : 'translateX(-100%)'};
    box-shadow: ${p => p.$open ? '4px 0 24px rgba(0,0,0,0.4)' : 'none'};
  }
`;

const SidebarOverlay = styled.div<{ $open: boolean }>`
  display: none;
  @media (max-width: 768px) {
    display: ${p => p.$open ? 'block' : 'none'};
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.5);
    z-index: 29;
    backdrop-filter: blur(2px);
  }
`;

// ─── Sidebar Header ───────────────────────────────────────────────────────────
const SidebarTop = styled.div`
  padding: 20px 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  border-bottom: 1px solid var(--border-primary);
`;

const AppBrand = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
`;

const BrandAvatar = styled.div`
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  font-weight: 700;
  flex-shrink: 0;
`;

const BrandName = styled.span`
  font-size: 17px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const NewChatBtn = styled.button`
  width: 100%;
  padding: 10px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;

  &:hover {
    opacity: 0.9;
    box-shadow: 0 4px 16px rgba(102, 126, 234, 0.35);
    transform: translateY(-1px);
  }
  &:active { transform: translateY(0); }
`;

// ─── Conversation List ────────────────────────────────────────────────────────
const ConvList = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 6px 8px;
`;

const ConvItem = styled.div<{ $active?: boolean }>`
  padding: 10px 12px;
  border-radius: 14px;
  cursor: pointer;
  background: ${p => p.$active ? 'rgba(102, 126, 234, 0.15)' : 'transparent'};
  border: 1px solid ${p => p.$active ? 'rgba(102, 126, 234, 0.3)' : 'transparent'};
  transition: all 0.18s;
  margin-bottom: 2px;

  &:hover {
    background: ${p => p.$active ? 'rgba(102, 126, 234, 0.2)' : 'rgba(255,255,255,0.04)'};
  }
`;

const ConvTitle = styled.div`
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ConvPreview = styled.div`
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

// ─── Sidebar Footer ───────────────────────────────────────────────────────────
const SidebarFooter = styled.div`
  padding: 12px 8px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  gap: 2px;
`;

const FooterBtn = styled.button`
  width: 100%;
  padding: 10px 14px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.18s;
  text-align: left;

  &:hover {
    background: rgba(255,255,255,0.06);
    color: var(--text-primary);
  }
  svg { flex-shrink: 0; opacity: 0.7; }
`;

const FooterBtnIcon = styled.span`
  font-size: 18px;
  width: 20px;
  text-align: center;
`;

// ─── Chat Area ────────────────────────────────────────────────────────────────
const ChatArea = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
`;

// ─── Chat Header ─────────────────────────────────────────────────────────────
const ChatHeader = styled.div`
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-primary);
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-secondary);
  backdrop-filter: blur(8px);
  min-height: 60px;
`;

const MenuToggle = styled.button`
  display: none;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: rgba(255,255,255,0.06);
  color: var(--text-primary);
  cursor: pointer;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.18s;

  &:hover { background: rgba(255,255,255,0.1); }

  @media (max-width: 768px) { display: flex; }
`;

const ChatTitle = styled.div`
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ModelSelector = styled.select`
  padding: 6px 10px;
  border: 1px solid var(--border-primary);
  border-radius: 20px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  max-width: 180px;
  transition: border-color 0.18s;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
  option { background: var(--bg-primary); }
`;

// ─── Messages ─────────────────────────────────────────────────────────────────
const Messages = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 20px 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  scroll-behavior: smooth;
`;

const MessageRow = styled.div<{ $user?: boolean }>`
  display: flex;
  align-items: flex-end;
  gap: 8px;
  justify-content: ${p => p.$user ? 'flex-end' : 'flex-start'};
  animation: ${fadeIn} 0.22s ease-out both;
`;

const AIAvatar = styled.div`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: white;
  flex-shrink: 0;
  margin-bottom: 2px;
`;

const BubbleWrap = styled.div<{ $user?: boolean }>`
  max-width: min(68%, 520px);
  display: flex;
  flex-direction: column;
  align-items: ${p => p.$user ? 'flex-end' : 'flex-start'};
`;

// Telegram-style bubble: user has rounded corners except bottom-right; AI except bottom-left
const Bubble = styled.div<{ $user?: boolean }>`
  padding: 10px 14px;
  border-radius: ${p => p.$user
    ? '18px 18px 4px 18px'
    : '18px 18px 18px 4px'};
  background: ${p => p.$user
    ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    : 'var(--bg-tertiary)'};
  color: ${p => p.$user ? '#fff' : 'var(--text-primary)'};
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
  white-space: pre-wrap;
  box-shadow: ${p => p.$user
    ? '0 2px 12px rgba(102, 126, 234, 0.25)'
    : '0 1px 4px rgba(0,0,0,0.15)'};
  border: ${p => p.$user ? 'none' : '1px solid var(--border-primary)'};
`;

const BubbleTime = styled.div<{ $user?: boolean }>`
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
  padding: 0 2px;
`;

// Typing indicator
const TypingBubble = styled(Bubble)`
  display: flex;
  gap: 5px;
  align-items: center;
  padding: 12px 16px;
`;

const TypingDot = styled.span<{ $delay: number }>`
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
  animation: ${pulse} 1.2s infinite ease-in-out;
  animation-delay: ${p => p.$delay}s;
`;

// ─── Input Area ───────────────────────────────────────────────────────────────
const InputArea = styled.div`
  padding: 12px 16px;
  border-top: 1px solid var(--border-primary);
  background: var(--bg-secondary);
  display: flex;
  gap: 10px;
  align-items: flex-end;
`;

const TextInput = styled.textarea`
  flex: 1;
  padding: 11px 16px;
  background: var(--bg-primary);
  border: 1.5px solid var(--border-primary);
  border-radius: 22px;
  color: var(--text-primary);
  font-size: 14px;
  font-family: inherit;
  resize: none;
  max-height: 130px;
  line-height: 1.5;
  transition: border-color 0.2s;
  scrollbar-width: none;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
  &::-webkit-scrollbar { display: none; }
  &::placeholder { color: var(--text-muted); }
`;

const SendBtn = styled.button<{ $active?: boolean }>`
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: ${p => p.$active
    ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    : 'var(--bg-elevated, var(--bg-tertiary))'};
  border: none;
  color: ${p => p.$active ? '#fff' : 'var(--text-muted)'};
  cursor: ${p => p.$active ? 'pointer' : 'not-allowed'};
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
  box-shadow: ${p => p.$active ? '0 2px 10px rgba(102,126,234,0.35)' : 'none'};

  &:hover:not(:disabled) {
    transform: ${p => p.$active ? 'scale(1.08)' : 'none'};
  }
`;

// ─── Welcome Screen ───────────────────────────────────────────────────────────
const Welcome = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  gap: 12px;
  padding: 40px;
  text-align: center;
`;

const WelcomeTitle = styled.h1`
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
`;

const WelcomeSub = styled.p`
  font-size: 14px;
  color: var(--text-muted);
  margin: 0;
`;

// ─── Types ────────────────────────────────────────────────────────────────────
interface Conversation {
  id: string;
  title: string;
  last_message?: string;
  updated_at: string;
}

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  created_at: string;
}

interface AIModel {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  is_default: boolean;
}

const ensureArray = <T,>(v: unknown): T[] => (Array.isArray(v) ? (v as T[]) : []);

// ─── Component ────────────────────────────────────────────────────────────────
const Chat: React.FC = () => {
  const navigate = useNavigate();
  const { conversationId: routeConvId } = useParams<{ conversationId?: string }>();
  const { user, logout } = useAuth();
  const { settings, updateSettings } = useSettings();

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConv, setActiveConv] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [models, setModels] = useState<AIModel[]>([]);
  const [selectedModel, setSelectedModel] = useState<string>('');

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Load conversations & models on mount
  useEffect(() => {
    loadConversations();
    loadModels();
  }, []);

  // Activate conversation from URL param on mount
  useEffect(() => {
    if (routeConvId) setActiveConv(routeConvId);
  }, [routeConvId]);

  useEffect(() => {
    if (activeConv) loadMessages(activeConv);
  }, [activeConv]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Auto-grow textarea
  useEffect(() => {
    const ta = textareaRef.current;
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = `${Math.min(ta.scrollHeight, 130)}px`;
  }, [input]);

  const loadConversations = async () => {
    try {
      const r = await apiClient.get('/conversations');
      setConversations(ensureArray<Conversation>(r.data?.conversations));
    } catch {
      setConversations([]);
    }
  };

  const loadMessages = async (id: string) => {
    try {
      const r = await apiClient.get(`/conversations/${id}/messages`);
      setMessages(ensureArray<Message>(r.data?.messages));
    } catch {
      setMessages([]);
    }
  };

  const loadModels = async () => {
    try {
      const r = await apiClient.models.list();
      const list = ensureArray<AIModel>(r?.models);
      setModels(list);
      // Pre-select: user preference > default model > first available
      const preferred = settings.chatPreferences.defaultModelId;
      const def = list.find(m => m.id === preferred) || list.find(m => m.is_default) || list[0];
      if (def) setSelectedModel(def.id);
    } catch {
      // Models endpoint may return empty when none configured — that's fine
    }
  };

  const handleNewChat = async () => {
    try {
      const r = await apiClient.post('/conversations', {
        title: '新对话',
        model_id: selectedModel || undefined,
      });
      await loadConversations();
      setActiveConv(r.data.id);
      setSidebarOpen(false);
    } catch {
      // ignore
    }
  };

  const handleSend = async () => {
    if (!input.trim() || !activeConv || loading) return;

    const text = input.trim();
    setInput('');
    setLoading(true);

    const tempMsg: Message = {
      id: `tmp-${Date.now()}`,
      role: 'user',
      content: text,
      created_at: new Date().toISOString(),
    };
    setMessages(prev => [...prev, tempMsg]);

    try {
      await apiClient.post(`/conversations/${activeConv}/messages`, {
        content: text,
        model_id: selectedModel || undefined,
      });
      await loadMessages(activeConv);
      await loadConversations();
    } catch {
      setMessages(prev => prev.filter(m => m.id !== tempMsg.id));
    } finally {
      setLoading(false);
    }
  };

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleModelChange = (modelId: string) => {
    setSelectedModel(modelId);
    // Persist as user default
    updateSettings({ chatPreferences: { ...settings.chatPreferences, defaultModelId: modelId } });
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const fmt = (d: string) =>
    new Date(d).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

  const activeTitle = conversations.find(c => c.id === activeConv)?.title ?? '对话';
  const canAdmin = user?.role === 'admin' || user?.role === 'super_admin';

  return (
    <Container>
      {/* Mobile sidebar overlay */}
      <SidebarOverlay $open={sidebarOpen} onClick={() => setSidebarOpen(false)} />

      {/* ── Sidebar ── */}
      <Sidebar $open={sidebarOpen}>
        <SidebarTop>
          <AppBrand>
            <BrandAvatar>A</BrandAvatar>
            <BrandName>AI Chat</BrandName>
          </AppBrand>
          <NewChatBtn onClick={handleNewChat}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            新对话
          </NewChatBtn>
        </SidebarTop>

        <ConvList>
          {conversations.map(c => (
            <ConvItem
              key={c.id}
              $active={activeConv === c.id}
              onClick={() => { setActiveConv(c.id); setSidebarOpen(false); }}
            >
              <ConvTitle>{c.title}</ConvTitle>
              {c.last_message && <ConvPreview>{c.last_message}</ConvPreview>}
            </ConvItem>
          ))}
        </ConvList>

        <SidebarFooter>
          <FooterBtn onClick={() => navigate('/memory')}>
            <FooterBtnIcon>🧠</FooterBtnIcon> 记忆管理
          </FooterBtn>
          <FooterBtn onClick={() => navigate('/leaderboard')}>
            <FooterBtnIcon>🏆</FooterBtnIcon> 排行榜
          </FooterBtn>
          <FooterBtn onClick={() => navigate('/settings')}>
            <FooterBtnIcon>⚙️</FooterBtnIcon> 设置
          </FooterBtn>
          {canAdmin && (
            <FooterBtn onClick={() => navigate('/admin')}>
              <FooterBtnIcon>🛡️</FooterBtnIcon> 管理后台
            </FooterBtn>
          )}
          <FooterBtn onClick={handleLogout} style={{ color: 'var(--text-muted)' }}>
            <FooterBtnIcon>🚪</FooterBtnIcon> 退出登录
          </FooterBtn>
        </SidebarFooter>
      </Sidebar>

      {/* ── Chat Area ── */}
      <ChatArea>
        <ChatHeader>
          <MenuToggle onClick={() => setSidebarOpen(o => !o)}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="3" y1="12" x2="21" y2="12"/>
              <line x1="3" y1="6" x2="21" y2="6"/>
              <line x1="3" y1="18" x2="21" y2="18"/>
            </svg>
          </MenuToggle>

          <ChatTitle>{activeConv ? activeTitle : 'AI Chat'}</ChatTitle>

          {models.length > 0 && (
            <ModelSelector
              value={selectedModel}
              onChange={e => handleModelChange(e.target.value)}
              title="选择模型"
            >
              {models.map(m => (
                <option key={m.id} value={m.id}>
                  {m.display_name || m.name}
                </option>
              ))}
            </ModelSelector>
          )}
        </ChatHeader>

        {activeConv ? (
          <>
            <Messages>
              {messages.map(msg => (
                <MessageRow key={msg.id} $user={msg.role === 'user'}>
                  {msg.role !== 'user' && <AIAvatar>✦</AIAvatar>}
                  <BubbleWrap $user={msg.role === 'user'}>
                    <Bubble $user={msg.role === 'user'}>{msg.content}</Bubble>
                    <BubbleTime $user={msg.role === 'user'}>{fmt(msg.created_at)}</BubbleTime>
                  </BubbleWrap>
                </MessageRow>
              ))}

              {loading && (
                <MessageRow>
                  <AIAvatar>✦</AIAvatar>
                  <BubbleWrap>
                    <TypingBubble>
                      <TypingDot $delay={0} />
                      <TypingDot $delay={0.2} />
                      <TypingDot $delay={0.4} />
                    </TypingBubble>
                  </BubbleWrap>
                </MessageRow>
              )}

              <div ref={messagesEndRef} />
            </Messages>

            <InputArea>
              <TextInput
                ref={textareaRef}
                rows={1}
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={handleKey}
                placeholder="输入消息… (Shift+Enter 换行)"
                disabled={loading}
              />
              <SendBtn $active={!!input.trim() && !loading} onClick={handleSend} disabled={!input.trim() || loading}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="22" y1="2" x2="11" y2="13"/>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"/>
                </svg>
              </SendBtn>
            </InputArea>
          </>
        ) : (
          <Welcome>
            <WelcomeTitle>欢迎使用 AI Chat</WelcomeTitle>
            <WelcomeSub>从左侧选择一个对话，或点击「新对话」开始聊天</WelcomeSub>
          </Welcome>
        )}
      </ChatArea>
    </Container>
  );
};

export default Chat;
