import React, { ReactNode } from 'react';
import styled from 'styled-components';
import { media } from '../../styles/responsive';

interface ResponsiveLayoutProps {
  children: ReactNode;
  sidebar?: ReactNode;
  showSidebar?: boolean;
}

const LayoutContainer = styled.div`
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: var(--bg-primary);
`;

const Sidebar = styled.aside<{ $show: boolean }>`
  width: 280px;
  background-color: var(--bg-tertiary);
  border-right: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  transition: transform 250ms ease-in-out;

  ${media.mobile} {
    position: fixed;
    top: 0;
    left: 0;
    height: 100%;
    z-index: 1030;
    transform: translateX(${props => props.$show ? '0' : '-100%'});
  }

  ${media.tablet} {
    width: 240px;
  }
`;

const MainContent = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
`;

const Overlay = styled.div<{ $show: boolean }>`
  display: none;

  ${media.mobile} {
    display: ${props => props.$show ? 'block' : 'none'};
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 1020;
  }
`;

export const ResponsiveLayout: React.FC<ResponsiveLayoutProps> = ({
  children,
  sidebar,
  showSidebar = false,
}) => {
  return (
    <LayoutContainer>
      {sidebar && (
        <>
          <Sidebar $show={showSidebar}>
            {sidebar}
          </Sidebar>
          <Overlay $show={showSidebar} />
        </>
      )}
      <MainContent>
        {children}
      </MainContent>
    </LayoutContainer>
  );
};

export default ResponsiveLayout;

