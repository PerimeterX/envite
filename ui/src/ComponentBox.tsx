import React from 'react';
import './ComponentBox.css';
import { Component } from './api';
import CloseIcon from '@mui/icons-material/Close';
import CheckIcon from '@mui/icons-material/Check';
import FavoriteIcon from '@mui/icons-material/Favorite';
import ArrowForwardIosIcon from '@mui/icons-material/ArrowForwardIos';
import { Tooltip } from '@mui/material';
import { NavLink } from 'react-router-dom';

interface ComponentBoxProps {
  component: Component;
  selected: boolean;
  select: () => void;
  enabled: boolean;
  setEnabled: (v: boolean) => void;
}

interface UiParams {
  success: boolean;
  icon: any;
  iconClass: string;
  tooltip: (name: string) => string;
}

export const statusToUiParams: { [key: string]: UiParams } = {
  stopped: {
    success: false,
    icon: CloseIcon,
    iconClass: '',
    tooltip: (name) => `${name} is not running`
  },
  running: {
    success: true,
    icon: FavoriteIcon,
    iconClass: 'heartbeat',
    tooltip: (name) => `${name} is running`
  },
  finished: {
    success: true,
    icon: CheckIcon,
    iconClass: '',
    tooltip: (name) => `${name} finished`
  }
};

function ComponentBox(props: ComponentBoxProps) {
  const uiParams =
    statusToUiParams[props.component.status] || statusToUiParams.stopped;
  const Icon = uiParams.icon;
  return (
    <div className="ComponentBox">
      <Tooltip
        title={
          props.enabled
            ? `Disable ${props.component.id}`
            : `Enable ${props.component.id}`
        }
      >
        <div
          className={`component-enabled ${props.enabled ? 'highlighted' : ''}`}
          onClick={() => props.setEnabled(!props.enabled)}
        >
          <div className="component-enabled-icon">
            <CheckIcon fontSize="inherit" />
          </div>
        </div>
      </Tooltip>
      <Tooltip title={uiParams.tooltip(props.component.id)}>
        <div
          className={`component-icon ${uiParams.success ? 'highlighted' : ''}`}
        >
          <Icon className={uiParams.iconClass} />
        </div>
      </Tooltip>
      <NavLink
        to={`/${props.component.id}`}
        className={({ isActive }) => 'nav ' + (isActive ? 'highlighted' : '')}
      >
        <div className="component-text">
          <div className="component-description">
            <div className="component-id">{props.component.id}</div>
            <div className="component-type">{props.component.type}</div>
          </div>
          <div className="component-arrow">
            <ArrowForwardIosIcon fontSize="inherit" />
          </div>
        </div>
      </NavLink>
    </div>
  );
}

export default ComponentBox;
