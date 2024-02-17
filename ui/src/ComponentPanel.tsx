// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React from 'react';
import './ComponentPanel.css';
import { Message, Component } from './api';
import { Button, IconButton, Tab, Tabs, Tooltip } from '@mui/material';
import PowerSettingsNewIcon from '@mui/icons-material/PowerSettingsNew';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import ComponentOutput from './ComponentOutput';
import ComponentInfo from './ComponentInfo';
import ComponentEnvVars from './ComponentEnvVars';
import {
    matchPath,
    useLocation,
    NavLink,
    Route,
    Routes,
    useNavigate
} from 'react-router-dom';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';

interface ComponentPanelProps {
    loading: boolean;
    component: Component;
    data: Message[] | undefined;
    clear: () => void;
    startComponent: () => void;
    stopComponent: () => void;
    restartComponent: () => void;
}

function useRouteMatch(patterns: readonly string[]) {
    const { pathname } = useLocation();

    for (let i = 0; i < patterns.length; i += 1) {
        const pattern = patterns[i];
        const possibleMatch = matchPath(pattern, pathname);
        if (possibleMatch !== null) {
            return possibleMatch;
        }
    }

    return null;
}

function ComponentPanel(props: ComponentPanelProps) {
    const navigate = useNavigate();
    const routeMatch = useRouteMatch([
        `/${props.component.id}/info`,
        `/${props.component.id}/env`,
        `/${props.component.id}`
    ]);

    return (
        <div className="ComponentPanel">
            <div className="header">
                <div>
                    <div className="title-control">
                        <div className="back-button">
                            <Tooltip title="Back to all logs">
                                <NavLink to={`/`}>
                                    <IconButton className="back-icon">
                                        <ArrowBackIcon color="inherit" />
                                    </IconButton>
                                </NavLink>
                            </Tooltip>
                        </div>
                        <div className="component-details">
                            <span className="component-name">
                                {props.component.id}
                            </span>
                            <span className="component-status">
                                {props.component.status}
                            </span>
                        </div>
                    </div>
                    <div className="tabs">
                        <Tabs value={routeMatch?.pattern?.path}>
                            <Tab
                                className="tab"
                                iconPosition="start"
                                label="Output"
                                value={`/${props.component.id}`}
                                to={`/${props.component.id}`}
                                component={NavLink}
                            />
                            <Tab
                                className="tab"
                                iconPosition="start"
                                label="Info"
                                value={`/${props.component.id}/info`}
                                to={`/${props.component.id}/info`}
                                component={NavLink}
                            />
                            {props.component.config.env && (
                                <Tab
                                    className="tab"
                                    iconPosition="start"
                                    label="Env Vars"
                                    value={`/${props.component.id}/env`}
                                    to={`/${props.component.id}/env`}
                                    component={NavLink}
                                />
                            )}
                        </Tabs>
                    </div>
                </div>
                <div className="actions">
                    <Button
                        disabled={
                            props.loading ||
                            (props.component.status !== 'running' &&
                                props.component.status !== 'starting')
                        }
                        onClick={() => {
                            navigate(`/${props.component.id}`);
                            props.restartComponent();
                        }}
                        size="small"
                        variant="outlined"
                        style={{ marginRight: 15 }}
                        startIcon={<RestartAltIcon />}
                    >
                        Restart
                    </Button>
                    <Button
                        disabled={
                            props.loading ||
                            (props.component.status !== 'running' &&
                                props.component.status !== 'starting')
                        }
                        onClick={props.stopComponent}
                        size="small"
                        variant="outlined"
                        style={{ marginRight: 15 }}
                        startIcon={<PowerSettingsNewIcon />}
                    >
                        Stop
                    </Button>
                    <Button
                        disabled={
                            props.loading ||
                            props.component.status === 'running' ||
                            props.component.status === 'starting'
                        }
                        onClick={() => {
                            navigate(`/${props.component.id}`);
                            props.startComponent();
                        }}
                        size="small"
                        variant="outlined"
                        startIcon={<PlayArrowIcon />}
                    >
                        Start
                    </Button>
                </div>
            </div>
            <div className="content">
                <div className="tabs-content">
                    <Routes>
                        <Route
                            path={`/`}
                            element={
                                <ComponentOutput
                                    component={props.component}
                                    data={props.data}
                                    title={<div />}
                                    clear={props.clear}
                                    scrolling={false}
                                />
                            }
                        />
                        <Route
                            path={`/info`}
                            element={
                                <ComponentInfo info={props.component.config} />
                            }
                        />
                        {props.component.config.env && (
                            <Route
                                path={`/env`}
                                element={
                                    <ComponentEnvVars
                                        envVars={props.component.config.env}
                                    />
                                }
                            />
                        )}
                    </Routes>
                </div>
            </div>
        </div>
    );
}

export default ComponentPanel;
