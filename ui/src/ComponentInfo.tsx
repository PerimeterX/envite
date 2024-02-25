// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useEffect, useState } from 'react';
import './ComponentOutput.css';
import { IconButton, MenuItem, Select, Tooltip } from '@mui/material';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LibraryAddCheckIcon from '@mui/icons-material/LibraryAddCheck';
import { dump } from 'js-yaml';
import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-yaml';
import './prism.css';

interface ComponentInfoProps {
    info: any;
}

type Mode = 'yaml' | 'json';

function ComponentInfo(props: ComponentInfoProps) {
    const [mode, setMode] = useState(getLatestMode('yaml'));
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        Prism.highlightAll();
    }, [props.info, mode]);

    const CopyIcon = copied ? LibraryAddCheckIcon : ContentCopyIcon;
    const data =
        mode === 'json'
            ? JSON.stringify(props.info, null, '    ')
            : dump(props.info, { indent: 4, noArrayIndent: false });

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
                        <MenuItem value="yaml">YAML format</MenuItem>
                        <MenuItem value="json">JSON format</MenuItem>
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
                <code className={`language-${mode}`}>{data}</code>
            </pre>
        </div>
    );
}

const STORAGE_KEY = 'component-info-mode';

function getLatestMode(defaultValue: Mode): Mode {
    return (localStorage.getItem(STORAGE_KEY) as Mode) || defaultValue;
}

function setLatestMode(mode: Mode) {
    localStorage.setItem(STORAGE_KEY, mode);
}

export default ComponentInfo;
