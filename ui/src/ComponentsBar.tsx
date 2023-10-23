import React, { useEffect, useRef } from 'react';
import './ComponentsBar.css';
import { ApiCall, Status } from './api';
import ComponentBox from './ComponentBox';
import { Badge, Button, IconButton, Tooltip } from '@mui/material';
import PowerSettingsNewIcon from '@mui/icons-material/PowerSettingsNew';
import SendIcon from '@mui/icons-material/Send';
import CancelScheduleSendIcon from '@mui/icons-material/CancelScheduleSend';
import DeleteIcon from '@mui/icons-material/Delete';
import ComponentBoxHeader from './ComponentBoxHeader';
import { useNavigate } from 'react-router-dom';
import SplitButton from './SplitButton';
import Loader from './Loader';
import EditNoteIcon from '@mui/icons-material/EditNote';

interface ComponentsBarProps {
    apiCall: ApiCall<any> | null;
    enabledIds: Set<string>;
    setEnabledIds: (v: Set<string>) => void;
    status: Status;
    apply: () => void;
    stopAll: () => void;
    stopAllAndClear: () => void;
    clearAllOutput: () => void;
    apiNotification: boolean;
    openApiLog: () => void;
}

function ComponentsBar(props: ComponentsBarProps) {
    const navigate = useNavigate();
    const elapsedRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!elapsedRef.current) {
            return;
        }

        if (!props.apiCall) {
            elapsedRef.current.innerText = '';
            return;
        }

        const interval = setInterval(() => {
            if (!elapsedRef.current) {
                return;
            }

            elapsedRef.current.innerText = formatElapsed(
                props.apiCall?.durationMs
            );
        }, 100);

        return () => clearInterval(interval);
    }, [elapsedRef, props.apiCall]);

    const components = props.status.components.flat() || [];
    return (
        <div className="ComponentsBar">
            <div className="components-bar-components">
                <ComponentBoxHeader
                    enabled={props.enabledIds.size > 0}
                    setEnabled={() => {
                        if (props.enabledIds.size === 0) {
                            const set = new Set(
                                components.map((component) => component.id)
                            );
                            props.setEnabledIds(set);
                        } else {
                            props.setEnabledIds(new Set());
                        }
                    }}
                />
                {components.map((component, i) => (
                    <div key={component.id}>
                        <ComponentBox
                            component={component}
                            selected={false}
                            select={() => {}}
                            enabled={props.enabledIds.has(component.id)}
                            setEnabled={(v) => {
                                const set = new Set(props.enabledIds.values());
                                if (v) {
                                    set.add(component.id);
                                } else {
                                    set.delete(component.id);
                                }
                                props.setEnabledIds(set);
                            }}
                        />
                        {i < components.length - 1 && <div className="trail" />}
                    </div>
                ))}
            </div>
            <div className="components-bar-footer">
                {props.apiCall ? (
                    <>
                        <div>
                            <div className="main-loader">
                                <div className="elapsed">
                                    <div>{props.apiCall?.title}</div>
                                    <div ref={elapsedRef} />
                                </div>
                                <Loader type="main" />
                            </div>
                        </div>
                        <div>
                            <Button
                                onClick={props.apiCall.cancel}
                                size="small"
                                variant="outlined"
                                color="error"
                                startIcon={<CancelScheduleSendIcon />}
                            >
                                Cancel
                            </Button>
                        </div>
                    </>
                ) : (
                    <>
                        <div className="view-log-holder">
                            <Tooltip title="View log">
                                <IconButton
                                    size="small"
                                    onClick={props.openApiLog}
                                >
                                    <Badge
                                        color="error"
                                        invisible={!props.apiNotification}
                                        variant="dot"
                                    >
                                        <EditNoteIcon
                                            fontSize="inherit"
                                            color="inherit"
                                        />
                                    </Badge>
                                </IconButton>
                            </Tooltip>
                        </div>
                        <div>
                            <SplitButton
                                variant="outlined"
                                options={[
                                    {
                                        title: 'Stop all',
                                        button: (
                                            <Button
                                                onClick={props.stopAll}
                                                size="small"
                                                startIcon={
                                                    <PowerSettingsNewIcon />
                                                }
                                            >
                                                Stop All
                                            </Button>
                                        )
                                    },
                                    {
                                        title: 'Cleanup',
                                        button: (
                                            <Button
                                                onClick={() => {
                                                    if (
                                                        window.confirm(
                                                            'This will clear all component data including removing resources' +
                                                                ' like docker images. Are you sure you want to proceed?'
                                                        )
                                                    ) {
                                                        props.stopAllAndClear();
                                                    }
                                                }}
                                                size="small"
                                                startIcon={<DeleteIcon />}
                                            >
                                                Cleanup
                                            </Button>
                                        )
                                    }
                                ]}
                            />
                            <Button
                                style={{ marginLeft: 15, marginTop: -10 }}
                                onClick={() => {
                                    navigate('/');
                                    props.apply();
                                }}
                                size="small"
                                variant="contained"
                                startIcon={<SendIcon />}
                            >
                                Apply
                            </Button>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}

export function formatElapsed(elapsedMs?: number) {
    if (!elapsedMs) {
        return '0';
    }

    const secondsTotal = Math.floor(elapsedMs / 1000);

    const minutes = Math.floor(secondsTotal / 60);
    const seconds = secondsTotal % 60;
    const ms = Math.floor((elapsedMs % 1000) / 100);

    if (!minutes) {
        return `${seconds}.${ms}s`;
    }

    const strMinutes = String(minutes).padStart(2, '0');
    const strSeconds = String(seconds).padStart(2, '0');
    const strMs = String(ms).padStart(1, '0');
    return `${strMinutes}:${strSeconds}.${strMs}`;
}

export default ComponentsBar;
