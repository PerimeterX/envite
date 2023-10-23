import React, { useCallback, useEffect, useRef, useState } from 'react';
import './SplitPanel.css';
import { Message, Status } from './api';
import ComponentOutput from './ComponentOutput';
import { Box, Grid, IconButton, Tooltip } from '@mui/material';
import ComponentTitle from './ComponentTitle';
import BlockIcon from '@mui/icons-material/Block';
import ZoomInIcon from '@mui/icons-material/ZoomIn';
import ZoomOutIcon from '@mui/icons-material/ZoomOut';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

interface SplitPanelProps {
    status: Status;
    output: { [key: string]: Message[] };
    clearOutput: (componentId: string) => void;
    clearAllOutput: () => void;
}

type Zoom = 1 | 2 | 3 | 4 | 5;

function SplitPanel(props: SplitPanelProps) {
    const [zoom, setZoom] = useState(getLatestZoom(2));
    const [canJumpUp, setCanJumpUp] = useState(false);
    const [canJumpDown, setCanJumpDown] = useState(false);
    const boxRef = useRef(null as any);

    const updateJumpOptions = useCallback((target: any) => {
        const { scrollTop, scrollHeight, offsetHeight } = target;
        if (scrollTop === 0) {
            setCanJumpUp(false);
        } else {
            setCanJumpUp(true);
        }

        if (Math.ceil(scrollTop) === scrollHeight - offsetHeight) {
            setCanJumpDown(false);
        } else {
            setCanJumpDown(true);
        }
    }, []);

    useEffect(() => {
        if (boxRef.current) {
            updateJumpOptions(boxRef.current);
        }
    }, [boxRef, updateJumpOptions, props.status]);

    const components = props.status.components.flat() || [];

    return (
        <div className="SplitPanel">
            <div className="title">
                <Tooltip title="Zoom out">
                    <span>
                        <IconButton
                            className="action"
                            disabled={zoom === 5}
                            onClick={() => {
                                const newZoom = (zoom + 1) as Zoom;
                                setZoom(newZoom);
                                setLatestZoom(newZoom);
                            }}
                        >
                            <ZoomOutIcon fontSize="inherit" color="inherit" />
                        </IconButton>
                    </span>
                </Tooltip>
                <Tooltip title="Zoom in">
                    <span>
                        <IconButton
                            className="action"
                            disabled={zoom === 1}
                            onClick={() => {
                                const newZoom = (zoom - 1) as Zoom;
                                setZoom(newZoom);
                                setLatestZoom(newZoom);
                            }}
                        >
                            <ZoomInIcon fontSize="inherit" color="inherit" />
                        </IconButton>
                    </span>
                </Tooltip>
                <Tooltip title="Clear output">
                    <span>
                        <IconButton
                            className="action"
                            onClick={props.clearAllOutput}
                        >
                            <BlockIcon fontSize="inherit" color="inherit" />
                        </IconButton>
                    </span>
                </Tooltip>
                <Tooltip title="Jump to top">
                    <span>
                        <IconButton
                            className="action"
                            disabled={!boxRef.current || !canJumpUp}
                            onClick={() => {
                                boxRef.current.scrollTop = 0;
                            }}
                        >
                            <KeyboardArrowUpIcon
                                fontSize="inherit"
                                color="inherit"
                            />
                        </IconButton>
                    </span>
                </Tooltip>
                <Tooltip title="Jump to bottom">
                    <span>
                        <IconButton
                            className="action"
                            disabled={!boxRef.current || !canJumpDown}
                            onClick={() => {
                                boxRef.current.scrollTop =
                                    boxRef.current.scrollHeight;
                            }}
                        >
                            <KeyboardArrowDownIcon
                                fontSize="inherit"
                                color="inherit"
                            />
                        </IconButton>
                    </span>
                </Tooltip>
            </div>
            <Box
                className="split"
                ref={boxRef}
                onScroll={(e) => updateJumpOptions(e.currentTarget)}
            >
                <Grid container>
                    {components.map((c) => (
                        <Grid
                            key={c.id}
                            lg={
                                zoom === 1
                                    ? 12
                                    : zoom === 2
                                    ? 6
                                    : zoom === 3
                                    ? 4
                                    : zoom === 4
                                    ? 3
                                    : 2
                            }
                            item={true}
                        >
                            <div className="panel">
                                <ComponentOutput
                                    title={<ComponentTitle component={c} />}
                                    component={c}
                                    key={c.id}
                                    data={props.output[c.id]}
                                    clear={() => props.clearOutput(c.id)}
                                />
                            </div>
                        </Grid>
                    ))}
                </Grid>
            </Box>
        </div>
    );
}

const STORAGE_KEY = 'split-screen-zoom';

function getLatestZoom(defaultValue: Zoom): Zoom {
    const value = localStorage.getItem(STORAGE_KEY);
    try {
        if (value) {
            const parsed = parseInt(value);
            if (parsed >= 1 && parsed <= 5) {
                return parsed as Zoom;
            }
        }
    } catch (e) {}
    return defaultValue;
}

function setLatestZoom(zoom: Zoom) {
    localStorage.setItem(STORAGE_KEY, zoom + '');
}

export default SplitPanel;
