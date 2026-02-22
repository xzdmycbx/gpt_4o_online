import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import { ensureArray } from '../../utils/safe';
import { GlassCard, GlassModal } from '../../styles/glass';

// ─── Styled components ────────────────────────────────────────────────────────

const ProviderCard = styled(GlassCard)`
  margin-bottom: 16px;
  overflow: hidden;
`;

const ProviderHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 20px;
  cursor: pointer;
  user-select: none;

  &:hover {
    background: rgba(102, 126, 234, 0.04);
  }
`;

const ProviderName = styled.span`
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
`;

const ProviderEndpoint = styled.span`
  font-size: 12px;
  color: var(--text-muted);
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`;

const Badge = styled.span<{ color?: string }>`
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  background: ${p => p.color ? `rgba(${p.color}, 0.15)` : 'rgba(102,126,234,0.15)'};
  color: ${p => p.color ? `rgb(${p.color})` : '#667eea'};
  flex-shrink: 0;
`;

const ChevronIcon = styled.span<{ open: boolean }>`
  font-size: 12px;
  color: var(--text-muted);
  transform: rotate(${p => p.open ? '180deg' : '0deg'});
  transition: transform 0.2s;
  flex-shrink: 0;
`;

const ProviderActions = styled.div`
  display: flex;
  gap: 6px;
  flex-shrink: 0;
`;

const IconBtn = styled.button<{ danger?: boolean }>`
  padding: 5px 12px;
  border-radius: 8px;
  border: 1px solid ${p => p.danger ? 'rgba(255,69,58,0.3)' : 'var(--border-primary)'};
  background: transparent;
  color: ${p => p.danger ? '#ff453a' : 'var(--text-secondary)'};
  font-size: 12px;
  cursor: pointer;
  transition: all 0.18s;

  &:hover {
    background: ${p => p.danger ? 'rgba(255,69,58,0.1)' : 'rgba(102,126,234,0.08)'};
    color: ${p => p.danger ? '#ff453a' : '#667eea'};
    border-color: ${p => p.danger ? '#ff453a' : '#667eea'};
  }
`;

const ProviderBody = styled.div`
  border-top: 1px solid var(--glass-border);
  padding: 12px 20px 16px;
`;

const ModelRow = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--border-primary);
  margin-bottom: 8px;
  background: rgba(255,255,255,0.03);
  transition: border-color 0.18s;

  &:hover { border-color: rgba(102,126,234,0.3); }
`;

const ModelName = styled.span`
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  flex: 1;
`;

const ModelMeta = styled.span`
  font-size: 12px;
  color: var(--text-muted);
`;

const AddModelBtn = styled.button`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: 10px;
  border: 1.5px dashed var(--border-primary);
  background: transparent;
  color: var(--text-muted);
  font-size: 13px;
  cursor: pointer;
  margin-top: 6px;
  transition: all 0.18s;

  &:hover {
    border-color: #667eea;
    color: #667eea;
    background: rgba(102,126,234,0.06);
  }
`;

const TopBar = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const AddProviderBtn = styled.button`
  padding: 9px 20px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;

  &:hover { opacity: 0.88; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(102,126,234,0.3); }
`;

const Backdrop = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
`;

const ModalContent = styled(GlassModal)`
  padding: 28px;
  width: 100%;
  max-width: 520px;
  max-height: 85vh;
  overflow-y: auto;
`;

const ModalTitle = styled.h3`
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 20px;
  color: var(--text-primary);
`;

const FormGroup = styled.div`
  margin-bottom: 16px;
`;

const Label = styled.label`
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
`;

const Input = styled.input`
  width: 100%;
  padding: 10px 14px;
  background: var(--bg-primary);
  border: 1.5px solid var(--border-primary);
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;
  transition: border-color 0.18s;
  box-sizing: border-box;

  &:focus { outline: none; border-color: #667eea; }
  &::placeholder { color: var(--text-muted); }
`;

const Select = styled.select`
  width: 100%;
  padding: 10px 14px;
  background: var(--bg-primary);
  border: 1.5px solid var(--border-primary);
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;

  &:focus { outline: none; border-color: #667eea; }
  option { background: var(--bg-primary); }
`;

const ModalActions = styled.div`
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
`;

const CancelBtn = styled.button`
  padding: 10px 20px;
  border-radius: 10px;
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
  border-radius: 10px;
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

const ErrorMsg = styled.div`
  margin-bottom: 14px;
  padding: 10px 14px;
  border-radius: 8px;
  background: rgba(252,129,129,0.1);
  border: 1px solid rgba(252,129,129,0.3);
  color: #fc8181;
  font-size: 13px;
`;

const EmptyModels = styled.div`
  font-size: 13px;
  color: var(--text-muted);
  padding: 8px 0;
  text-align: center;
`;

// ─── Types ────────────────────────────────────────────────────────────────────

interface Provider {
  id: string;
  name: string;
  display_name: string;
  provider_type: string;
  api_endpoint: string;
  is_active: boolean;
  description: string;
  model_count: number;
}

interface Model {
  id: string;
  name: string;
  display_name: string;
  provider: string;
  api_endpoint: string;
  model_identifier: string;
  max_tokens: number;
  is_default: boolean;
  provider_id?: string;
}

type ModalKind =
  | { type: 'addProvider' }
  | { type: 'editProvider'; provider: Provider }
  | { type: 'addModel'; provider: Provider }
  | { type: 'editModel'; model: Model };

// ─── Component ────────────────────────────────────────────────────────────────

const ModelManagement: React.FC = () => {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [modal, setModal] = useState<ModalKind | null>(null);
  const [saveError, setSaveError] = useState('');
  const [saving, setSaving] = useState(false);

  // Provider form
  const [providerForm, setProviderForm] = useState({
    name: '', display_name: '', provider_type: 'openai', api_endpoint: '', api_key: '', description: '',
  });

  // Model form
  const [modelForm, setModelForm] = useState({
    name: '', display_name: '', model_identifier: '', provider: 'openai',
    api_endpoint: '', api_key: '', max_tokens: 4096,
    supports_streaming: true, provider_id: '',
  });

  useEffect(() => {
    loadAll();
  }, []);

  const loadAll = async () => {
    try {
      const [pRes, mRes] = await Promise.all([
        apiClient.get('/admin/providers'),
        apiClient.get('/admin/models'),
      ]);
      setProviders(ensureArray<Provider>(pRes.data?.providers));
      setModels(ensureArray<Model>(mRes.data?.models));
    } catch (e) {
      console.error('Failed to load data:', e);
    }
  };

  const modelsForProvider = (providerID: string) =>
    models.filter(m => m.provider_id === providerID);

  const standaloneModels = models.filter(m => !m.provider_id);

  const toggleExpand = (id: string) =>
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }));

  // ── Provider actions ──

  const openAddProvider = () => {
    setProviderForm({ name: '', display_name: '', provider_type: 'openai', api_endpoint: '', api_key: '', description: '' });
    setSaveError('');
    setModal({ type: 'addProvider' });
  };

  const openEditProvider = (p: Provider) => {
    setProviderForm({ name: p.name, display_name: p.display_name, provider_type: p.provider_type, api_endpoint: p.api_endpoint, api_key: '', description: p.description });
    setSaveError('');
    setModal({ type: 'editProvider', provider: p });
  };

  const saveProvider = async () => {
    setSaving(true);
    setSaveError('');
    try {
      if (modal?.type === 'editProvider') {
        await apiClient.put(`/admin/providers/${modal.provider.id}`, {
          display_name: providerForm.display_name,
          provider_type: providerForm.provider_type,
          api_endpoint: providerForm.api_endpoint,
          api_key: providerForm.api_key || undefined,
          description: providerForm.description,
        });
      } else {
        await apiClient.post('/admin/providers', providerForm);
      }
      setModal(null);
      await loadAll();
    } catch (e: any) {
      setSaveError(e?.response?.data?.error || '保存失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  const deleteProvider = async (id: string) => {
    if (!confirm('确定删除此供应商？关联模型将解除绑定后才能删除。')) return;
    try {
      await apiClient.delete(`/admin/providers/${id}`);
      await loadAll();
    } catch (e: any) {
      alert(e?.response?.data?.error || '删除失败');
    }
  };

  // ── Model actions ──

  const openAddModel = (provider: Provider) => {
    setModelForm({
      name: '', display_name: '', model_identifier: '', provider: provider.provider_type,
      api_endpoint: '', api_key: '', max_tokens: 4096, supports_streaming: true,
      provider_id: provider.id,
    });
    setSaveError('');
    setModal({ type: 'addModel', provider });
  };

  const openEditModel = (m: Model) => {
    setModelForm({
      name: m.name, display_name: m.display_name || '', model_identifier: m.model_identifier,
      provider: m.provider, api_endpoint: m.api_endpoint || '', api_key: '',
      max_tokens: m.max_tokens, supports_streaming: true,
      provider_id: m.provider_id || '',
    });
    setSaveError('');
    setModal({ type: 'editModel', model: m });
  };

  const saveModel = async () => {
    setSaving(true);
    setSaveError('');
    try {
      if (modal?.type === 'editModel') {
        await apiClient.put(`/admin/models/${modal.model.id}`, {
          display_name: modelForm.display_name,
          max_tokens: modelForm.max_tokens,
          api_endpoint: modelForm.api_endpoint || undefined,
          api_key: modelForm.api_key || undefined,
          supports_streaming: modelForm.supports_streaming,
        });
      } else {
        await apiClient.post('/admin/models', {
          name: modelForm.name,
          display_name: modelForm.display_name,
          model_identifier: modelForm.model_identifier,
          provider: modelForm.provider,
          provider_id: modelForm.provider_id || undefined,
          api_endpoint: modelForm.api_endpoint || undefined,
          api_key: modelForm.api_key || undefined,
          max_tokens: modelForm.max_tokens,
          supports_streaming: modelForm.supports_streaming,
        });
      }
      setModal(null);
      await loadAll();
    } catch (e: any) {
      setSaveError(e?.response?.data?.error || '保存失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  const deleteModel = async (id: string) => {
    if (!confirm('确定删除这个模型吗？')) return;
    try {
      await apiClient.delete(`/admin/models/${id}`);
      await loadAll();
    } catch (e) {
      console.error('Failed to delete model:', e);
    }
  };

  const setDefaultModel = async (id: string) => {
    try {
      await apiClient.put(`/admin/models/${id}/default`);
      await loadAll();
    } catch (e) {
      console.error('Failed to set default:', e);
    }
  };

  // ── Modal form ──

  const renderModal = () => {
    if (!modal) return null;
    const isProviderModal = modal.type === 'addProvider' || modal.type === 'editProvider';
    const isModelModal = modal.type === 'addModel' || modal.type === 'editModel';

    return (
      <Backdrop onClick={e => { if (e.target === e.currentTarget) setModal(null); }}>
        <ModalContent>
          <ModalTitle>
            {modal.type === 'addProvider' && '添加供应商'}
            {modal.type === 'editProvider' && '编辑供应商'}
            {modal.type === 'addModel' && `在「${(modal as any).provider.display_name}」下添加模型`}
            {modal.type === 'editModel' && '编辑模型'}
          </ModalTitle>

          {saveError && <ErrorMsg>{saveError}</ErrorMsg>}

          {isProviderModal && (
            <>
              {modal.type === 'addProvider' && (
                <FormGroup>
                  <Label>供应商标识（唯一，英文）</Label>
                  <Input
                    value={providerForm.name}
                    onChange={e => setProviderForm({ ...providerForm, name: e.target.value })}
                    placeholder="如: my-openai"
                  />
                </FormGroup>
              )}
              <FormGroup>
                <Label>显示名称</Label>
                <Input
                  value={providerForm.display_name}
                  onChange={e => setProviderForm({ ...providerForm, display_name: e.target.value })}
                  placeholder="如: My OpenAI"
                />
              </FormGroup>
              <FormGroup>
                <Label>供应商类型</Label>
                <Select
                  value={providerForm.provider_type}
                  onChange={e => setProviderForm({ ...providerForm, provider_type: e.target.value })}
                >
                  <option value="openai">OpenAI 兼容</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="custom">自定义</option>
                </Select>
              </FormGroup>
              <FormGroup>
                <Label>API Endpoint</Label>
                <Input
                  value={providerForm.api_endpoint}
                  onChange={e => setProviderForm({ ...providerForm, api_endpoint: e.target.value })}
                  placeholder="https://api.openai.com/v1/chat/completions"
                />
              </FormGroup>
              <FormGroup>
                <Label>API Key{modal.type === 'editProvider' && ' (留空不修改)'}</Label>
                <Input
                  type="password"
                  value={providerForm.api_key}
                  onChange={e => setProviderForm({ ...providerForm, api_key: e.target.value })}
                  placeholder="sk-..."
                />
              </FormGroup>
              <FormGroup>
                <Label>描述（可选）</Label>
                <Input
                  value={providerForm.description}
                  onChange={e => setProviderForm({ ...providerForm, description: e.target.value })}
                  placeholder=""
                />
              </FormGroup>
            </>
          )}

          {isModelModal && (
            <>
              {modal.type === 'addModel' && (
                <>
                  <FormGroup>
                    <Label>模型名称（唯一）</Label>
                    <Input
                      value={modelForm.name}
                      onChange={e => setModelForm({ ...modelForm, name: e.target.value })}
                      placeholder="如: gpt-4o-mini"
                    />
                  </FormGroup>
                  <FormGroup>
                    <Label>模型标识符</Label>
                    <Input
                      value={modelForm.model_identifier}
                      onChange={e => setModelForm({ ...modelForm, model_identifier: e.target.value })}
                      placeholder="如: gpt-4o-mini"
                    />
                  </FormGroup>
                </>
              )}
              <FormGroup>
                <Label>显示名称</Label>
                <Input
                  value={modelForm.display_name}
                  onChange={e => setModelForm({ ...modelForm, display_name: e.target.value })}
                  placeholder="如: GPT-4o Mini"
                />
              </FormGroup>
              <FormGroup>
                <Label>最大 Tokens</Label>
                <Input
                  type="number"
                  value={modelForm.max_tokens}
                  onChange={e => setModelForm({ ...modelForm, max_tokens: parseInt(e.target.value) || 4096 })}
                />
              </FormGroup>
              <FormGroup>
                <Label>API 端点（覆盖供应商，可选）</Label>
                <Input
                  value={modelForm.api_endpoint}
                  onChange={e => setModelForm({ ...modelForm, api_endpoint: e.target.value })}
                  placeholder="留空使用供应商配置"
                />
              </FormGroup>
              <FormGroup>
                <Label>API Key（覆盖供应商，可选）</Label>
                <Input
                  type="password"
                  value={modelForm.api_key}
                  onChange={e => setModelForm({ ...modelForm, api_key: e.target.value })}
                  placeholder="留空使用供应商配置"
                />
              </FormGroup>
            </>
          )}

          <ModalActions>
            <CancelBtn onClick={() => setModal(null)}>取消</CancelBtn>
            <SaveBtn onClick={isProviderModal ? saveProvider : saveModel} disabled={saving}>
              {saving ? '保存中…' : '保存'}
            </SaveBtn>
          </ModalActions>
        </ModalContent>
      </Backdrop>
    );
  };

  return (
    <div>
      <TopBar>
        <span style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          {providers.length} 个供应商 · {models.length} 个模型
        </span>
        <AddProviderBtn onClick={openAddProvider}>
          + 添加供应商
        </AddProviderBtn>
      </TopBar>

      {/* Provider cards */}
      {providers.map(p => {
        const pModels = modelsForProvider(p.id);
        const isOpen = !!expanded[p.id];
        return (
          <ProviderCard key={p.id}>
            <ProviderHeader onClick={() => toggleExpand(p.id)}>
              <ProviderName>{p.display_name}</ProviderName>
              <ProviderEndpoint title={p.api_endpoint}>{p.api_endpoint}</ProviderEndpoint>
              <Badge>{pModels.length} 个模型</Badge>
              <ProviderActions onClick={e => e.stopPropagation()}>
                <IconBtn onClick={() => openEditProvider(p)}>编辑</IconBtn>
                <IconBtn danger onClick={() => deleteProvider(p.id)}>删除</IconBtn>
              </ProviderActions>
              <ChevronIcon open={isOpen}>▼</ChevronIcon>
            </ProviderHeader>

            {isOpen && (
              <ProviderBody>
                {pModels.length === 0 && (
                  <EmptyModels>此供应商下还没有模型</EmptyModels>
                )}
                {pModels.map(m => (
                  <ModelRow key={m.id}>
                    <ModelName>{m.display_name || m.name}</ModelName>
                    <ModelMeta>{m.model_identifier}</ModelMeta>
                    {m.is_default && <Badge color="52,199,89">默认</Badge>}
                    <IconBtn onClick={() => openEditModel(m)}>编辑</IconBtn>
                    {!m.is_default && <IconBtn onClick={() => setDefaultModel(m.id)}>设为默认</IconBtn>}
                    <IconBtn danger onClick={() => deleteModel(m.id)}>删除</IconBtn>
                  </ModelRow>
                ))}
                <AddModelBtn onClick={() => openAddModel(p)}>
                  + 添加模型
                </AddModelBtn>
              </ProviderBody>
            )}
          </ProviderCard>
        );
      })}

      {/* Standalone models (no provider) */}
      {standaloneModels.length > 0 && (
        <ProviderCard>
          <ProviderHeader onClick={() => toggleExpand('__standalone')}>
            <ProviderName>独立模型（未绑定供应商）</ProviderName>
            <Badge>{standaloneModels.length} 个模型</Badge>
            <ChevronIcon open={!!expanded['__standalone']}>▼</ChevronIcon>
          </ProviderHeader>
          {!!expanded['__standalone'] && (
            <ProviderBody>
              {standaloneModels.map(m => (
                <ModelRow key={m.id}>
                  <ModelName>{m.display_name || m.name}</ModelName>
                  <ModelMeta>{m.model_identifier}</ModelMeta>
                  {m.is_default && <Badge color="52,199,89">默认</Badge>}
                  <IconBtn onClick={() => openEditModel(m)}>编辑</IconBtn>
                  {!m.is_default && <IconBtn onClick={() => setDefaultModel(m.id)}>设为默认</IconBtn>}
                  <IconBtn danger onClick={() => deleteModel(m.id)}>删除</IconBtn>
                </ModelRow>
              ))}
            </ProviderBody>
          )}
        </ProviderCard>
      )}

      {renderModal()}
    </div>
  );
};

export default ModelManagement;
