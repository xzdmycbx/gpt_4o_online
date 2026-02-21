import React, { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const Container = styled.div`
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  position: relative;
`;

const Sidebar = styled.div<{ $isOpen: boolean }>`
  width: 320px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  transition: transform 0.3s ease-in-out;

  @media (max-width: 768px) {
    width: 100%;
    position: absolute;
    height: 100%;
    z-index: 10;
    transform: ${props => props.$isOpen ? 'translateX(0)' : 'translateX(-100%)'};
  }
`;

const MobileMenuButton = styled.button`
  display: none;
  position: fixed;
  top: 16px;
  left: 16px;
  z-index: 20;
  width: 44px;
  height: 44px;
  background: #667eea;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);

  @media (max-width: 768px) {
    display: flex;
  }

  &:active {
    transform: scale(0.95);
  }
`;

const Overlay = styled.div<{ $isOpen: boolean }>`
  display: none;

  @media (max-width: 768px) {
    display: ${props => props.$isOpen ? 'block' : 'none'};
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 9;
  }
`;

const SidebarHeader = styled.div`
  padding: 20px;
  border-bottom: 1px solid var(--border-primary);
`;

const Title = styled.h2`
  margin: 0 0 16px 0;
  font-size: 20px;
  font-weight: 600;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const NewChatButton = styled.button`
  width: 100%;
  padding: 12px 20px;
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
`;

const ConversationList = styled.div`
  flex: 1;
  overflow-y: auto;
`;

const ConversationItem = styled.div<{ active?: boolean }>`
  padding: 16px 20px;
  cursor: pointer;
  border-left: 3px solid ${props => props.active ? '#667eea' : 'transparent'};
  background: ${props => props.active ? 'rgba(102, 126, 234, 0.1)' : 'transparent'};
  border-bottom: 1px solid var(--border-primary);
  transition: all 0.2s;

  &:hover {
    background: rgba(102, 126, 234, 0.05);
  }
`;

const ConversationTitle = styled.div`
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ConversationPreview = styled.div`
  font-size: 13px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ChatArea = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
`;

const ChatHeader = styled.div`
  padding: 20px;
  border-bottom: 1px solid var(--border-primary);
`;

const Messages = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
`;

const MessageBubble = styled.div<{ isUser?: boolean }>`
  max-width: 70%;
  align-self: ${props => props.isUser ? 'flex-end' : 'flex-start'};
  background: ${props => props.isUser ? '#2b5278' : 'var(--bg-tertiary)'};
  padding: 12px 16px;
  border-radius: 12px;
  color: var(--text-primary);
  word-wrap: break-word;
  line-height: 1.5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
`;

const MessageTime = styled.div<{ isUser?: boolean }>`
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
  text-align: ${props => props.isUser ? 'right' : 'left'};
`;

const InputArea = styled.div`
  padding: 20px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  gap: 12px;
  background: var(--bg-secondary);
`;

const Input = styled.textarea`
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 14px;
  resize: none;
  font-family: inherit;
  max-height: 120px;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
`;

const SendButton = styled.button`
  padding: 12px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;

  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const WelcomeMessage = styled.div`
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
  margin: auto;
`;

const WelcomeTitle = styled.h1`
  font-size: 32px;
  margin-bottom: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const TopActions = styled.div`
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 21;
  display: flex;
  gap: 8px;

  @media (max-width: 768px) {
    top: 16px;
    right: 16px;
    flex-wrap: wrap;
    justify-content: flex-end;
    max-width: calc(100vw - 84px);
  }
`;

const ActionButton = styled.button`
  padding: 8px 12px;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: #667eea;
    color: #667eea;
  }
`;

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

const ensureArray = <T,>(value: unknown): T[] => {
  return Array.isArray(value) ? (value as T[]) : [];
};

const Chat: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConversation, setActiveConversation] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadConversations();
  }, []);

  useEffect(() => {
    if (activeConversation) {
      loadMessages(activeConversation);
    }
  }, [activeConversation]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadConversations = async () => {
    try {
      const response = await apiClient.get('/conversations');
      // Backend returns {conversations: [...]}; handle null/invalid payloads safely.
      setConversations(ensureArray<Conversation>(response.data?.conversations));
    } catch (error) {
      console.error('Failed to load conversations:', error);
      setConversations([]);
    }
  };

  const loadMessages = async (conversationId: string) => {
    try {
      const response = await apiClient.get(`/conversations/${conversationId}/messages`);
      // Backend returns {messages: [...]}; handle null/invalid payloads safely.
      setMessages(ensureArray<Message>(response.data?.messages));
    } catch (error) {
      console.error('Failed to load messages:', error);
      setMessages([]);
    }
  };

  const handleNewChat = async () => {
    try {
      const response = await apiClient.post('/conversations', {
        title: '新对话',
      });
      await loadConversations();
      setActiveConversation(response.data.id);
    } catch (error) {
      console.error('Failed to create conversation:', error);
    }
  };

  const handleSendMessage = async () => {
    if (!input.trim() || !activeConversation || loading) return;

    const userMessage = input.trim();
    setInput('');
    setLoading(true);

    // Add user message optimistically
    const tempUserMsg: Message = {
      id: 'temp-' + Date.now(),
      role: 'user',
      content: userMessage,
      created_at: new Date().toISOString(),
    };
    setMessages(prev => [...prev, tempUserMsg]);

    try {
      await apiClient.post(`/conversations/${activeConversation}/messages`, {
        content: userMessage,
      });

      // Reload messages to get the AI response
      await loadMessages(activeConversation);
      await loadConversations(); // Update conversation list
    } catch (error) {
      console.error('Failed to send message:', error);
      // Remove optimistic message on error
      setMessages(prev => prev.filter(m => m.id !== tempUserMsg.id));
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const canAccessAdmin = user?.role === 'admin' || user?.role === 'super_admin';

  return (
    <Container>
      <TopActions>
        <ActionButton onClick={() => navigate('/chat')}>对话</ActionButton>
        <ActionButton onClick={() => navigate('/settings')}>设置</ActionButton>
        {canAccessAdmin && (
          <ActionButton onClick={() => navigate('/admin')}>管理后台</ActionButton>
        )}
        <ActionButton onClick={handleLogout}>退出登录</ActionButton>
      </TopActions>

      <MobileMenuButton onClick={() => setSidebarOpen(!sidebarOpen)}>
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="3" y1="12" x2="21" y2="12"/>
          <line x1="3" y1="6" x2="21" y2="6"/>
          <line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </MobileMenuButton>

      <Overlay $isOpen={sidebarOpen} onClick={() => setSidebarOpen(false)} />

      <Sidebar $isOpen={sidebarOpen}>
        <SidebarHeader>
          <Title>对话列表</Title>
          <NewChatButton onClick={handleNewChat}>+ 新建对话</NewChatButton>
        </SidebarHeader>
        <ConversationList>
          {conversations.map(conv => (
            <ConversationItem
              key={conv.id}
              active={activeConversation === conv.id}
              onClick={() => {
                setActiveConversation(conv.id);
                setSidebarOpen(false); // Close sidebar on mobile after selecting
              }}
            >
              <ConversationTitle>{conv.title}</ConversationTitle>
              {conv.last_message && (
                <ConversationPreview>{conv.last_message}</ConversationPreview>
              )}
            </ConversationItem>
          ))}
        </ConversationList>
      </Sidebar>

      <ChatArea>
        {activeConversation ? (
          <>
            <ChatHeader>
              <Title>
                {conversations.find(c => c.id === activeConversation)?.title || '对话'}
              </Title>
            </ChatHeader>
            <Messages>
              {messages.map(msg => (
                <div key={msg.id}>
                  <MessageBubble isUser={msg.role === 'user'}>
                    {msg.content}
                  </MessageBubble>
                  <MessageTime isUser={msg.role === 'user'}>
                    {formatTime(msg.created_at)}
                  </MessageTime>
                </div>
              ))}
              {loading && (
                <MessageBubble>
                  正在思考...
                </MessageBubble>
              )}
              <div ref={messagesEndRef} />
            </Messages>
            <InputArea>
              <Input
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="输入消息... (Shift+Enter 换行)"
                rows={1}
              />
              <SendButton onClick={handleSendMessage} disabled={loading || !input.trim()}>
                发送
              </SendButton>
            </InputArea>
          </>
        ) : (
          <WelcomeMessage>
            <WelcomeTitle>欢迎使用 AI Chat</WelcomeTitle>
            <p>选择一个对话或创建新对话开始聊天</p>
          </WelcomeMessage>
        )}
      </ChatArea>
    </Container>
  );
};

export default Chat;

