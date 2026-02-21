import React from 'react';
import styled from 'styled-components';
import { media } from '../../styles/responsive';

interface MobileNavBarProps {
  onMenuClick?: () => void;
  title?: string;
}

const NavContainer = styled.nav`
  display: none;

  ${media.mobile} {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    padding: 0 16px;
    background-color: var(--bg-tertiary);
    border-bottom: 1px solid var(--border-primary);
  }
`;

const MenuButton = styled.button`
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 24px;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 150ms ease-in-out;

  &:hover {
    background-color: rgba(232, 234, 237, 0.08);
  }

  &:active {
    background-color: rgba(232, 234, 237, 0.12);
  }
`;

const Title = styled.h1`
  font-size: 18px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
`;

export const MobileNavBar: React.FC<MobileNavBarProps> = ({ onMenuClick, title = 'AI Chat' }) => {
  return (
    <NavContainer>
      <MenuButton onClick={onMenuClick} aria-label="Menu">
        ☰
      </MenuButton>
      <Title>{title}</Title>
      <div style={{ width: '40px' }} /> {/* Spacer for centering */}
    </NavContainer>
  );
};

export default MobileNavBar;

