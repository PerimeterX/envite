// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useEffect, useState } from 'react';
import './SplitPanel.css';
import { Message, Status } from './api';
import ComponentOutput from './ComponentOutput';
import { Box, Grid, IconButton, Tooltip } from '@mui/material';
import ComponentTitle from './ComponentTitle';
import BlockIcon from '@mui/icons-material/Block';
import ZoomInIcon from '@mui/icons-material/ZoomIn';
import ZoomOutIcon from '@mui/icons-material/ZoomOut';

interface SplitPanelProps {
    status: Status;
    output: { [key: string]: Message[] };
    clearOutput: (componentId: string) => void;
    clearAllOutput: () => void;
}

type Zoom = 1 | 2 | 3 | 4 | 5;

function SplitPanel(props: SplitPanelProps) {
    const [zoom, setZoom] = useState(getLatestZoom(2));
    const [scrolling, setScrolling] = useState(false);

    useEffect(() => {
        const onKeyDown = (e: any) => {
            if (e.keyCode === 93 || e.keyCode === 17) {
                setScrolling(true);
            }
        };
        const onKeyUp = (e: any) => {
            if (e.keyCode === 93 || e.keyCode === 17) {
                setScrolling(false);
            }
        };
        document.addEventListener('keydown', onKeyDown);
        document.addEventListener('keyup', onKeyUp);
        return () => {
            document.removeEventListener('keydown', onKeyDown);
            document.removeEventListener('keyup', onKeyUp);
        };
    }, []);

    const components = props.status.components.flat() || [];

    return (
        <div className="SplitPanel">
            <div className="title">
                {scrolling ? (
                    <div className="info scrolling-mode">Scrolling mode...</div>
                ) : (
                    <div className="info">{`Hold ${scrollButton()} for scrolling mode`}</div>
                )}
                <div>
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
                                <ZoomOutIcon
                                    fontSize="inherit"
                                    color="inherit"
                                />
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
                                <ZoomInIcon
                                    fontSize="inherit"
                                    color="inherit"
                                />
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
                </div>
            </div>
            <Box className="split">
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
                            <div
                                className={`panel ${
                                    scrolling ? 'scrolling' : ''
                                }`}
                            >
                                <ComponentOutput
                                    title={<ComponentTitle component={c} />}
                                    component={c}
                                    key={c.id}
                                    data={props.output[c.id]}
                                    clear={() => props.clearOutput(c.id)}
                                    scrolling={scrolling}
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

function scrollButton() {
    const isMac = /(Mac|iPhone|iPod|iPad)/i.test(navigator.platform);
    return isMac ? '⌘' : 'ctrl';
}

export default SplitPanel;
