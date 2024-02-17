// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useCallback, useEffect, useState } from 'react';
import './App.css';
import * as api from './api';
import ComponentsBar from './ComponentsBar';
import { ApiCall, Message, Status } from './api';
import {
    Alert,
    AlertTitle,
    Badge,
    createTheme,
    IconButton,
    Snackbar,
    ThemeProvider,
    Tooltip
} from '@mui/material';
import MainPanel from './MainPanel';
import { AxiosError, CanceledError } from 'axios';
import { BrowserRouter } from 'react-router-dom';
import OfflinePanel from './OfflinePanel';
import ApiLogDialog, {
    ApiLogEntry,
    formatApiFailure,
    formatApiSuccess
} from './ApiLogDialog';
import FilterListIcon from '@mui/icons-material/FilterList';
import FavoriteIcon from '@mui/icons-material/Favorite';
import ProductTour, { ProductTourProps } from './ProductTour';

const REFRESH_STATUS_INTERVAL = 5000;

const theme = createTheme({
    palette: {
        mode: 'dark',
        primary: {
            main: '#00ffbb'
        }
    },
    typography: {
        fontFamily: "'Montserrat', sans-serif"
    }
});

interface ApiMessage {
    type: 'error' | 'success';
    title: string;
    msg: string;
}

function App() {
    const [apiCall, setApiCall] = useState<ApiCall<any> | null>(null);
    const [enabledIds, setEnabledIds] = useState<Set<string> | null>(null);
    const [status, setStatus] = useState<Status | null>(null);
    const [output, setOutput] = useState({} as { [key: string]: Message[] });
    const [apiMessage, setApiMessage] = useState<ApiMessage | null>(null);
    const [apiNotification, setApiNotification] = useState(false);
    const [apiLog, setApiLog] = useState<ApiLogEntry[]>([]);
    const [logDialog, setLogDialog] = useState(false);
    const [productTourProps, setProductTourProps] =
        useState<ProductTourProps | null>(null);

    const handleCloseApiMessage = useCallback(
        (_?: React.SyntheticEvent | Event, reason?: string) => {
            if (reason === 'clickaway') {
                return;
            }
            setApiMessage(null);
        },
        []
    );

    const reportApiSuccess = useCallback((call: ApiCall<any>) => {
        setApiCall(null);
        setApiLog((prevState) => [...prevState, { call, date: new Date() }]);
        setApiMessage({
            type: 'success',
            title: 'Success',
            msg: formatApiSuccess(call)
        });
    }, []);

    const reportApiError = useCallback(
        (call: ApiCall<any>, error: any) => {
            setApiCall(null);
            const msg = formatHttpError(error);
            setApiLog((prevState) => [
                ...prevState,
                { call: call, err: msg, date: new Date() }
            ]);
            if (status && !(error instanceof CanceledError)) {
                setApiNotification(true);
                setApiMessage({
                    type: 'error',
                    title: `${call.title} failed`,
                    msg: formatApiFailure(call, msg)
                });
            }
        },
        [status]
    );

    const fetchStatus = useCallback(async () => {
        const call = api.getStatus('Fetching Status');
        let newStatus;
        try {
            newStatus = await call.promise;
        } catch (e) {
            reportApiError(call, e);
            setStatus(null);
            return;
        }
        if (!status || newStatus.id !== status.id) {
            document.title = `ENVITE - ${newStatus.id}`;
        }
        if (!status || isDifferentStatus(newStatus, status)) {
            setStatus(newStatus);
        }
        if (!enabledIds) {
            const enabled = new Set(
                newStatus.components
                    ?.flat()
                    .filter((c) => c.status !== 'stopped')
                    .map((s) => s.id)
            );
            setEnabledIds(enabled);
        }
    }, [reportApiError, enabledIds, status]);

    const showProductTour = useCallback(() => {
        if (completedTour()) {
            return Promise.resolve();
        }
        return new Promise((resolve) => {
            setProductTourProps({
                open: true,
                close: () => {
                    tourCompleted();
                    setProductTourProps(null);
                    resolve(null);
                }
            });
        });
    }, [setProductTourProps]);

    const startComponent = useCallback(
        async (id: string) => {
            await showProductTour();
            const call = api.startComponent(`Starting ${id}`, id);
            setApiCall(call);
            try {
                await call.promise;
            } catch (e) {
                reportApiError(call, e);
                return;
            }
            reportApiSuccess(call);
            await fetchStatus();
        },
        [fetchStatus, reportApiError, reportApiSuccess, showProductTour]
    );

    const stopComponent = useCallback(
        async (id: string) => {
            await showProductTour();
            const call = api.stopComponent(`Stopping ${id}`, id);
            setApiCall(call);
            try {
                await call.promise;
            } catch (e) {
                reportApiError(call, e);
                return;
            }
            reportApiSuccess(call);
            await fetchStatus();
        },
        [fetchStatus, reportApiError, reportApiSuccess, showProductTour]
    );

    const restartComponent = useCallback(
        async (id: string) => {
            await showProductTour();
            const call = api.restartComponent(`Restarting ${id}`, id);
            setApiCall(call);
            try {
                await call.promise;
            } catch (e) {
                reportApiError(call, e);
                return;
            }
            reportApiSuccess(call);
            await fetchStatus();
        },
        [fetchStatus, reportApiError, reportApiSuccess, showProductTour]
    );

    const apply = useCallback(async () => {
        const call = api.apply(`Applying state`, Array.from(enabledIds || []));
        setApiCall(call);
        try {
            await call.promise;
        } catch (e) {
            reportApiError(call, e);
            return;
        }
        reportApiSuccess(call);
        await fetchStatus();
    }, [enabledIds, fetchStatus, reportApiError, reportApiSuccess]);

    const stopAll = useCallback(
        async (cleanup: boolean) => {
            const call = api.stopAll(`Stopping all`, cleanup);
            setApiCall(call);
            try {
                await call.promise;
            } catch (e) {
                reportApiError(call, e);
                return;
            }
            reportApiSuccess(call);
            await fetchStatus();
        },
        [fetchStatus, reportApiError, reportApiSuccess]
    );

    const getOutputContinuously = useCallback(async () => {
        try {
            await api.getOutput((component, messages) => {
                setOutput((prevState) => {
                    const newOutput = { ...prevState };
                    newOutput[component] = [...(newOutput[component] || [])];
                    newOutput[component].push(...messages);
                    return newOutput;
                });
            });
        } catch (e) {
            setTimeout(getOutputContinuously, REFRESH_STATUS_INTERVAL);
        }
    }, []);

    const getStatusContinuously = useCallback(() => {
        const interval = setInterval(() => {
            fetchStatus().then();
        }, REFRESH_STATUS_INTERVAL);
        return () => {
            clearInterval(interval);
        };
    }, [fetchStatus]);

    useEffect(() => {
        fetchStatus().then();
        return getStatusContinuously();
    }, [fetchStatus, getStatusContinuously]);

    useEffect(() => {
        getOutputContinuously().then();
    }, [getOutputContinuously]);

    const runningComponents = countRunningComponents(status);
    return (
        <div className="App">
            <ThemeProvider theme={theme}>
                <header>
                    <div>
                        <img
                            className="logo"
                            src="/logo-small.svg"
                            alt="envite logo"
                        />
                        <span className="sub-title">
                            {status
                                ? status.id
                                : 'dev environments for testing and continuous integration'}
                        </span>
                    </div>
                    <div className="action-bar">
                        <Tooltip title="View log">
                            <IconButton
                                className="action-button"
                                onClick={() => {
                                    setApiNotification(false);
                                    setLogDialog(true);
                                }}
                            >
                                <Badge
                                    color="error"
                                    invisible={!apiNotification}
                                    variant="dot"
                                >
                                    <FilterListIcon
                                        fontSize="inherit"
                                        color="inherit"
                                    />
                                </Badge>
                            </IconButton>
                        </Tooltip>
                        <Tooltip
                            title={`${
                                runningComponents || 'No'
                            } components currently running`}
                            className={
                                runningComponents ? 'active' : 'inactive'
                            }
                        >
                            <span className="component-counter">
                                <span className="component-counter-number">
                                    {runningComponents}
                                </span>
                                <FavoriteIcon
                                    fontSize="inherit"
                                    color="inherit"
                                />
                            </span>
                        </Tooltip>
                    </div>
                </header>
                <main>
                    {status ? (
                        <BrowserRouter>
                            <ComponentsBar
                                apiCall={apiCall}
                                status={status}
                                enabledIds={enabledIds || new Set()}
                                setEnabledIds={setEnabledIds}
                                apply={apply}
                                stopAll={() => stopAll(false)}
                                stopAllAndClear={() => stopAll(true)}
                            />
                            <MainPanel
                                loading={apiCall !== null}
                                status={status}
                                output={output}
                                stopComponent={stopComponent}
                                startComponent={startComponent}
                                restartComponent={restartComponent}
                                clearOutput={(componentId) => {
                                    setOutput((prevState) => {
                                        const newOutput = { ...prevState };
                                        newOutput[componentId] = [];
                                        return newOutput;
                                    });
                                }}
                                clearAllOutput={() => {
                                    setOutput({});
                                }}
                            />
                        </BrowserRouter>
                    ) : (
                        <OfflinePanel />
                    )}
                </main>
                <Snackbar
                    open={apiMessage !== null}
                    autoHideDuration={6000}
                    onClose={handleCloseApiMessage}
                >
                    <Alert
                        onClose={handleCloseApiMessage}
                        severity={apiMessage?.type}
                    >
                        <AlertTitle>{apiMessage?.title}</AlertTitle>
                        {apiMessage?.msg}
                    </Alert>
                </Snackbar>
                <ApiLogDialog
                    apiLog={apiLog}
                    open={logDialog}
                    onClose={() => setLogDialog(false)}
                />
                {productTourProps && <ProductTour {...productTourProps} />}
                {/* we cache logo-large.svg, so we're able to use it in OfflinePanel.tsx */}
                {/* when the backend is down and can't service it */}
                <img
                    id="logo-large-cache"
                    src="/logo-large.svg"
                    alt="ENVITE Icon - large"
                />
            </ThemeProvider>
        </div>
    );
}

function formatHttpError(e: any): string {
    if (e instanceof AxiosError) {
        const data = e.response?.data;
        if (data) {
            const message = data.error;
            if (message) {
                return message;
            }
            if (typeof data === 'string') {
                return data;
            }
            return JSON.stringify(data);
        }
    }
    return e.toString();
}

function isDifferentStatus(newStatus: Status, oldStatus: Status): boolean {
    const newStatuses = newStatus.components.flat().map((c) => c.status);
    const oldStatuses = oldStatus.components.flat().map((c) => c.status);
    if (
        (newStatuses && !oldStatuses) ||
        (oldStatuses && !newStatuses) ||
        newStatuses?.length !== oldStatuses?.length
    ) {
        return true;
    }

    if (!newStatuses || !oldStatuses) {
        return false;
    }

    for (let i = 0; i < newStatuses.length; i++) {
        if (newStatuses[i] !== oldStatuses[i]) {
            return true;
        }
    }

    return false;
}

function countRunningComponents(status: Status | null) {
    if (!status || !status.components) {
        return 0;
    }
    let count = 0;
    status.components.flat().forEach((c) => {
        if (c.status === 'running') {
            count++;
        }
    });
    return count;
}

const PRODUCT_TOUR_STORAGE_KEY = 'product-tour';

function completedTour() {
    return Boolean(localStorage.getItem(PRODUCT_TOUR_STORAGE_KEY));
}

function tourCompleted() {
    localStorage.setItem(PRODUCT_TOUR_STORAGE_KEY, '1');
}

export default App;
