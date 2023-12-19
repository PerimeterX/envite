import React from 'react';
import './ComponentBox.css';
import CheckIcon from '@mui/icons-material/Check';
import ArrowForwardIosIcon from '@mui/icons-material/ArrowForwardIos';
import { Tooltip } from '@mui/material';
import { NavLink } from 'react-router-dom';

interface ComponentBoxHeaderProps {
    enabled: boolean;
    setEnabled: () => void;
}

function ComponentBoxHeader(props: ComponentBoxHeaderProps) {
    return (
        <div className="ComponentBox">
            <Tooltip title={props.enabled ? 'Disable all' : 'Enable all'}>
                <div
                    className={`component-enabled ${
                        props.enabled ? 'highlighted' : ''
                    }`}
                    onClick={() => props.setEnabled()}
                >
                    <div className="component-enabled-icon">
                        <CheckIcon fontSize="inherit" />
                    </div>
                </div>
            </Tooltip>
            <NavLink
                to="/"
                className={({ isActive }) =>
                    'nav ' + (isActive ? 'highlighted' : '')
                }
            >
                <div className="component-text">
                    <div className="component-description">
                        <div className="component-id">logs</div>
                        <div className="component-type">
                            Show Components Output
                        </div>
                    </div>
                    <div className="component-arrow">
                        <ArrowForwardIosIcon fontSize="inherit" />
                    </div>
                </div>
            </NavLink>
        </div>
    );
}

export default ComponentBoxHeader;
