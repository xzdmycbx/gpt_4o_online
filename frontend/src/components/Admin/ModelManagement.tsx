import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import apiClient from '../../api/client';
import { ensureArray } from '../../utils/safe';

const Card = styled.div`
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid var(--border-primary);
`;

const CardHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`;

const ModelName = styled.h3`
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
`;

const Badge = styled.span<{ isDefault?: boolean }>`
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  background: ${props => props.isDefault ? 'rgba(72, 187, 120, 0.2)' : 'rgba(102, 126, 234, 0.2)'};
  color: ${props => props.isDefault ? '#48bb78' : '#667eea'};
`;

const Info = styled.div`
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
`;

const ButtonGroup = styled.div`
  display: flex;
  gap: 12px;
  margin-top: 16px;
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
  transition: all 0.2s;

  &:hover {
    transform: translateY(-1px);
    opacity: 0.9;
  }
`;

const AddButton = styled(Button)`
  margin-bottom: 24px;
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
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 32px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  border: 1px solid var(--border-primary);
`;

const ModalTitle = styled.h3`
  margin: 0 0 24px 0;
  font-size: 20px;
  color: var(--text-primary);
`;

const FormGroup = styled.div`
  margin-bottom: 20px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  color: var(--text-secondary);
  font-size: 14px;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
`;

const Select = styled.select`
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;

  &:focus {
    outline: none;
    border-color: #667eea;
  }
`;

interface Model {
  id: string;
  name: string;
  display_name: string;
  provider: string;
  api_endpoint: string;
  model_identifier: string;
  max_tokens: number;
  is_default: boolean;
}

const ModelManagement: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);
  const [saveError, setSaveError] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    provider: 'openai',
    api_endpoint: '',
    api_key: '',
    model_identifier: '',
    max_tokens: 4096,
  });

  useEffect(() => {
    loadModels();
  }, []);

  const loadModels = async () => {
    try {
      const response = await apiClient.get('/admin/models');
      setModels(ensureArray<Model>(response.data?.models));
    } catch (error) {
      console.error('Failed to load models:', error);
      setModels([]);
    }
  };

  const handleAdd = () => {
    setEditingModel(null);
    setSaveError('');
    setFormData({
      name: '',
      display_name: '',
      provider: 'openai',
      api_endpoint: 'https://api.openai.com/v1/chat/completions',
      api_key: '',
      model_identifier: 'gpt-4',
      max_tokens: 4096,
    });
    setShowModal(true);
  };

  const handleEdit = (model: Model) => {
    setEditingModel(model);
    setSaveError('');
    setFormData({
      name: model.name,
      display_name: model.display_name || '',
      provider: model.provider,
      api_endpoint: model.api_endpoint,
      api_key: '', // API key never returned from server; user must re-enter to change
      model_identifier: model.model_identifier,
      max_tokens: model.max_tokens,
    });
    setShowModal(true);
  };

  const handleSave = async () => {
    setSaveError('');
    try {
      if (editingModel) {
        await apiClient.put(`/admin/models/${editingModel.id}`, formData);
      } else {
        await apiClient.post('/admin/models', formData);
      }
      await loadModels();
      setShowModal(false);
    } catch (error: any) {
      setSaveError(error?.response?.data?.error || '保存失败，请重试');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除这个模型吗？')) return;
    try {
      await apiClient.delete(`/admin/models/${id}`);
      await loadModels();
    } catch (error) {
      console.error('Failed to delete model:', error);
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      await apiClient.put(`/admin/models/${id}/default`);
      await loadModels();
    } catch (error) {
      console.error('Failed to set default:', error);
    }
  };

  return (
    <div>
      <AddButton onClick={handleAdd}>添加新模型</AddButton>

      {models.map(model => (
        <Card key={model.id}>
          <CardHeader>
            <ModelName>{model.name}</ModelName>
            {model.is_default && <Badge isDefault>默认</Badge>}
          </CardHeader>
          <Info><strong>提供商:</strong> {model.provider}</Info>
          <Info><strong>模型标识:</strong> {model.model_identifier}</Info>
          <Info><strong>API端点:</strong> {model.api_endpoint}</Info>
          <Info><strong>最大Tokens:</strong> {model.max_tokens}</Info>
          <ButtonGroup>
            <Button onClick={() => handleEdit(model)}>编辑</Button>
            {!model.is_default && (
              <Button onClick={() => handleSetDefault(model.id)}>设为默认</Button>
            )}
            <Button variant="danger" onClick={() => handleDelete(model.id)}>删除</Button>
          </ButtonGroup>
        </Card>
      ))}

      {showModal && (
        <Modal onMouseDown={(e) => { if (e.target === e.currentTarget) setShowModal(false); }}>
          <ModalContent>
            <ModalTitle>{editingModel ? '编辑模型' : '添加新模型'}</ModalTitle>
            {saveError && (
              <div style={{ marginBottom: 16, padding: '10px 14px', borderRadius: 8, background: 'rgba(252,129,129,0.1)', border: '1px solid rgba(252,129,129,0.3)', color: '#fc8181', fontSize: 13 }}>
                {saveError}
              </div>
            )}
            <FormGroup>
              <Label>模型名称</Label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="例如: gpt-4"
              />
            </FormGroup>
            <FormGroup>
              <Label>显示名称</Label>
              <Input
                value={formData.display_name}
                onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                placeholder="例如: GPT-4"
              />
            </FormGroup>
            <FormGroup>
              <Label>提供商</Label>
              <Select
                value={formData.provider}
                onChange={(e) => setFormData({ ...formData, provider: e.target.value })}
              >
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="custom">自定义</option>
              </Select>
            </FormGroup>
            <FormGroup>
              <Label>API 端点</Label>
              <Input
                value={formData.api_endpoint}
                onChange={(e) => setFormData({ ...formData, api_endpoint: e.target.value })}
                placeholder="https://api.openai.com/v1/chat/completions"
              />
            </FormGroup>
            <FormGroup>
              <Label>API 密钥</Label>
              <Input
                type="password"
                value={formData.api_key}
                onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                placeholder="sk-..."
              />
            </FormGroup>
            <FormGroup>
              <Label>模型标识符</Label>
              <Input
                value={formData.model_identifier}
                onChange={(e) => setFormData({ ...formData, model_identifier: e.target.value })}
                placeholder="gpt-4"
              />
            </FormGroup>
            <FormGroup>
              <Label>最大 Tokens</Label>
              <Input
                type="number"
                value={formData.max_tokens}
                onChange={(e) => setFormData({ ...formData, max_tokens: parseInt(e.target.value) || 4096 })}
              />
            </FormGroup>
            <ButtonGroup>
              <Button onClick={handleSave}>保存</Button>
              <Button variant="secondary" onClick={() => setShowModal(false)}>取消</Button>
            </ButtonGroup>
          </ModalContent>
        </Modal>
      )}
    </div>
  );
};

export default ModelManagement;

