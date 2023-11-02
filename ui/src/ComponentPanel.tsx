import React from 'react';
import './ComponentPanel.css';
import { Message, Component } from './api';
import { Button, Tab, Tabs } from '@mui/material';
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
        `/${props.component.id}/output`,
        `/${props.component.id}/env`,
        `/${props.component.id}`
    ]);

    return (
        <div className="ComponentPanel">
            <div className="title">
                <div>
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
                        <Tab
                            className="tab"
                            iconPosition="start"
                            label="Env Vars"
                            value={`/${props.component.id}/env`}
                            to={`/${props.component.id}/env`}
                            component={NavLink}
                        />
                    </Tabs>
                </div>
                <div className="actions">
                    <Button
                        disabled={
                            props.loading ||
                            props.component.status !== 'running'
                        }
                        onClick={props.restartComponent}
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
                            props.component.status !== 'running'
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
                            props.component.status === 'running'
                        }
                        onClick={() => {
                            navigate(`/${props.component.id}/output`);
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
                                />
                            }
                        />
                        <Route
                            path={`/info`}
                            element={
                                <ComponentInfo info={props.component.info} />
                            }
                        />
                        <Route
                            path={`/env`}
                            element={
                                <ComponentEnvVars
                                    envVars={props.component.env_vars}
                                />
                            }
                        />
                    </Routes>
                </div>
            </div>
        </div>
    );
}

export default ComponentPanel;
