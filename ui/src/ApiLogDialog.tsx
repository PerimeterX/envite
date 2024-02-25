// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React from 'react';
import './ApiLogDialog.css';
import { ApiCall } from './api';
import {
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableRow,
    Typography
} from '@mui/material';
import Button from '@mui/material/Button';
import { formatElapsed } from './ComponentsBar';

interface ApiLogDialogProps {
    open: boolean;
    onClose: () => void;
    apiLog: ApiLogEntry[];
}

export interface ApiLogEntry {
    date: Date;
    call: ApiCall<any>;
    err?: string;
}

function ApiLogDialog(props: ApiLogDialogProps) {
    return (
        <Dialog
            fullWidth={true}
            maxWidth={'xl'}
            open={props.open}
            onClose={props.onClose}
            scroll="paper"
            className="dialog"
        >
            <DialogTitle>Operation Log</DialogTitle>
            <DialogContent
                ref={(el) => {
                    if (el) {
                        const element = el as any;
                        element.scrollTop = element.scrollHeight;
                    }
                }}
            >
                <TableContainer>
                    <Table>
                        <TableBody>
                            {props.apiLog.map((entry, i) => (
                                <TableRow
                                    key={i}
                                    sx={{
                                        '&:last-child td, &:last-child th': {
                                            border: 0
                                        }
                                    }}
                                >
                                    <TableCell>
                                        {formatDate(entry.date)}
                                    </TableCell>
                                    <TableCell>
                                        {entry.err ? (
                                            <Typography
                                                fontSize="inherit"
                                                color="error"
                                            >
                                                {formatApiFailure(
                                                    entry.call,
                                                    entry.err
                                                )}
                                            </Typography>
                                        ) : (
                                            <Typography
                                                fontSize="inherit"
                                                color="primary"
                                            >
                                                {formatApiSuccess(entry.call)}
                                            </Typography>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.onClose}>Close</Button>
            </DialogActions>
        </Dialog>
    );
}

export function formatApiSuccess(call: ApiCall<any>) {
    const duration = formatElapsed(call);
    return `${call.title} finished after ${duration}`;
}

export function formatApiFailure(call: ApiCall<any>, err: string) {
    const duration = formatElapsed(call);
    return `${call.title} failed after ${duration} due to: ${err}`;
}

function formatDate(date: Date) {
    const hoursValue = date.getHours();
    const minutesValue = date.getMinutes();
    const secondsValue = date.getSeconds();

    const hours = String(hoursValue).padStart(2, '0');
    const minutes = String(minutesValue).padStart(2, '0');
    const seconds = String(secondsValue).padStart(2, '0');

    return `${hours}:${minutes}:${seconds}`;
}

export default ApiLogDialog;
