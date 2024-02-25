// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useEffect, useRef, useState } from 'react';
import './ComponentOutput.css';
import { Message, Component } from './api';
import { ansicolor } from 'ansicolor';
import AlarmIcon from '@mui/icons-material/Alarm';
import AlarmOffIcon from '@mui/icons-material/AlarmOff';
import OpenInFullIcon from '@mui/icons-material/OpenInFull';
import BlockIcon from '@mui/icons-material/Block';
import KeyboardDoubleArrowDownIcon from '@mui/icons-material/KeyboardDoubleArrowDown';
import { Badge, IconButton, Tooltip } from '@mui/material';
import { ReactJSXElement } from '@emotion/react/types/jsx-namespace';

const MAX_ENTRIES = 100;

interface ComponentOutputProps {
    title: ReactJSXElement;
    component: Component;
    data: Message[] | undefined;
    clear: () => void;
    scrolling: boolean;
}

function ComponentOutput(props: ComponentOutputProps) {
    const [showTime, setShowTime] = useState(false);
    const [stickToBottom, setStickToBottom] = useState(true);
    const [missedEntries, setMissedEntries] = useState(0);
    const codeRef = useRef(null as any);

    useEffect(() => {
        if (!codeRef.current) {
            return;
        }
        const update = () => {
            const entries: Message[] | undefined = stickToBottom
                ? props.data?.slice(-MAX_ENTRIES)
                : codeRef.current.lastEntries;
            codeRef.current.innerHTML = formatOutput(entries, showTime);
            codeRef.current.size = props.data?.length || 0;
            codeRef.current.showingTime = showTime;
            codeRef.current.lastEntries = entries;
            if (stickToBottom) {
                codeRef.current.scrollTop = codeRef.current.scrollHeight;
            }
        };
        if ((!props.data || props.data.length === 0) && codeRef.current.size) {
            codeRef.current.size = 0;
            setStickToBottom(true);
            update();
            return;
        }
        if (stickToBottom || codeRef.current.showingTime !== showTime) {
            update();
            return;
        }
        if (props.data) {
            setMissedEntries(props.data.length - codeRef.current.size);
        }
    }, [props.data, stickToBottom, showTime]);

    const TimeIcon = showTime ? AlarmOffIcon : AlarmIcon;

    return (
        <div className="ComponentOutput">
            <div className="component-output-title">
                {props.title}
                <div>
                    {!stickToBottom && (
                        <Tooltip title="Scroll down">
                            <IconButton
                                className="action"
                                onClick={() => setStickToBottom(true)}
                                size="small"
                            >
                                <Badge
                                    color="primary"
                                    badgeContent={missedEntries}
                                >
                                    <KeyboardDoubleArrowDownIcon
                                        fontSize="inherit"
                                        color="inherit"
                                    />
                                </Badge>
                            </IconButton>
                        </Tooltip>
                    )}
                    <Tooltip title="Clear output">
                        <IconButton
                            className="action"
                            onClick={props.clear}
                            size="small"
                        >
                            <BlockIcon fontSize="inherit" color="inherit" />
                        </IconButton>
                    </Tooltip>
                    <Tooltip title="View full history">
                        <IconButton
                            className="action"
                            onClick={() =>
                                showFullHistory(
                                    props.component.id,
                                    formatOutput(props.data, true)
                                )
                            }
                            size="small"
                        >
                            <OpenInFullIcon
                                fontSize="inherit"
                                color="inherit"
                            />
                        </IconButton>
                    </Tooltip>
                    <Tooltip
                        title={showTime ? 'Hide timestamps' : 'Show timestamps'}
                    >
                        <IconButton
                            className="action"
                            onClick={() => setShowTime((preState) => !preState)}
                            size="small"
                        >
                            <TimeIcon fontSize="inherit" color="inherit" />
                        </IconButton>
                    </Tooltip>
                </div>
            </div>
            <pre
                className={`component-output-content ${
                    props.scrolling ? 'scrolling' : ''
                }`}
                ref={codeRef}
                onScroll={(e) => {
                    if (
                        Math.ceil(e.currentTarget.scrollTop) ===
                        e.currentTarget.scrollHeight -
                            e.currentTarget.offsetHeight
                    ) {
                        setStickToBottom(true);
                    } else {
                        setStickToBottom(false);
                    }
                }}
            ></pre>
        </div>
    );
}

const formattingStyles: { [key: string]: string } = {
    italic: 'font-style: italic',
    bold: 'font-weight: bold',
    red: 'color: rgb(190 86 86)',
    green: 'color: #00ffbb',
    yellow: 'color: rgb(223 138 88)',
    blue: 'color: rgb(124 124 221)',
    magenta: 'color: rgb(205 108 205)',
    cyan: 'color: rgb(82 177 255)',
    dim: 'color: rgb(90 90 90)'
};

function formatOutput(data: Message[] | undefined, showTime: boolean) {
    if (!data) {
        return '';
    }
    return data
        .map((msg) => {
            let time = '';
            if (showTime) {
                time = `<span style="${formattingStyles.dim}">${msg.time}</span> `;
            }
            return (
                time +
                ansicolor
                    .parse(msg.data)
                    .spans.map((span) => {
                        let styles = [];
                        if (
                            span.color?.name &&
                            formattingStyles[span.color?.name]
                        ) {
                            styles.push(formattingStyles[span.color?.name]);
                        }
                        if (span.italic) {
                            styles.push(formattingStyles.italic);
                        }
                        if (span.bold) {
                            styles.push(formattingStyles.bold);
                        }
                        return `<span style="${styles.join(';')}">${
                            span.text
                        }</span>`;
                    })
                    .join('')
            );
        })
        .join('');
}

function showFullHistory(componentId: string, output?: string) {
    const child = window.open('about:blank')!;
    child.document.write(`
        <head>
            <title>Full history of ${componentId}</title>
        </head>
        <body style="background:#101010;color:#9b9b9b;font-size:14px;">
            <pre><code>${output}</code></pre>
        </body>
    `);
    child.document.close();
}

export default ComponentOutput;
