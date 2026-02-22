import styled from 'styled-components';

export const GlassPanel = styled.div`
  background: var(--glass-bg);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
`;

export const GlassCard = styled.div`
  background: var(--glass-bg);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  box-shadow: var(--glass-shadow);
`;

export const GlassModal = styled.div`
  background: var(--glass-bg);
  backdrop-filter: blur(calc(var(--glass-blur) * 1.5));
  -webkit-backdrop-filter: blur(calc(var(--glass-blur) * 1.5));
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  box-shadow: var(--glass-shadow), 0 20px 60px rgba(0, 0, 0, 0.4);
`;
