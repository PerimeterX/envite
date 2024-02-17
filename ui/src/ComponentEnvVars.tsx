// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useEffect, useState } from 'react';
import './ComponentOutput.css';
import { IconButton, MenuItem, Select, Tooltip } from '@mui/material';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LibraryAddCheckIcon from '@mui/icons-material/LibraryAddCheck';
import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-bash';
import './prism.css';

interface ComponentEnvVarsProps {
    envVars: { [key: string]: string };
}

const modeToLanguage = {
    json: 'json',
    terminal: 'bash',
    jetbrains: 'bash'
};

type Mode = 'json' | 'terminal' | 'jetbrains';

function ComponentEnvVars(props: ComponentEnvVarsProps) {
    const [mode, setMode] = useState(getLatestMode('json'));
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        Prism.highlightAll();
    }, [props.envVars, mode]);

    const CopyIcon = copied ? LibraryAddCheckIcon : ContentCopyIcon;
    const data = !props.envVars
        ? ''
        : mode === 'json'
        ? JSON.stringify(props.envVars, null, '    ')
        : mode === 'terminal'
        ? formatEnv(props.envVars, ' ')
        : formatEnv(props.envVars, ';');

    return (
        <div className="ComponentOutput">
            <div className="component-output-title">
                <div className="title-secondary">
                    <Select
                        className="select"
                        variant="standard"
                        value={mode}
                        classes={{
                            icon: 'dropdown'
                        }}
                        sx={{
                            ':before': { borderBottom: 'none' },
                            ':after': { borderBottom: 'none' }
                        }}
                        onChange={(e) => {
                            const mode = e.target.value as Mode;
                            setMode(mode);
                            setLatestMode(mode);
                        }}
                    >
                        <MenuItem value="json">JSON format</MenuItem>
                        <MenuItem value="terminal">Terminal format</MenuItem>
                        <MenuItem value="jetbrains">JetBrains format</MenuItem>
                    </Select>
                </div>
                <div>
                    <Tooltip title="Copy to clipboard">
                        <IconButton
                            style={{ marginLeft: 20 }}
                            className="action"
                            onClick={() => {
                                window.navigator.clipboard.writeText(data);
                                setCopied(true);
                                setTimeout(() => setCopied(false), 3000);
                            }}
                            size="small"
                        >
                            <CopyIcon
                                fontSize="inherit"
                                color={copied ? 'primary' : 'inherit'}
                            />
                        </IconButton>
                    </Tooltip>
                </div>
            </div>
            <pre className="component-output-content" style={{ fontSize: 16 }}>
                <code className={`language-${modeToLanguage[mode]}`}>
                    {data}
                </code>
            </pre>
        </div>
    );
}

const STORAGE_KEY = 'component-env-mode';

function getLatestMode(defaultValue: Mode): Mode {
    return (localStorage.getItem(STORAGE_KEY) as Mode) || defaultValue;
}

function setLatestMode(mode: Mode) {
    localStorage.setItem(STORAGE_KEY, mode);
}

function formatEnv(
    envVars: { [key: string]: string },
    separator: string
): string {
    const entries = [];
    for (const key of Object.keys(envVars)) {
        entries.push(`${key}=${envVars[key]}`);
    }
    return entries.join(separator);
}

export default ComponentEnvVars;
