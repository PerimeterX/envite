import axios, { CancelTokenSource } from 'axios';

const BASE_URL = process.env.REACT_APP_BASE_URL || ''; // = 'http://localhost:8080';

export interface Status {
    id: string;
    components: [[Component]];
}

export interface Component {
    id: string;
    type: string;
    status: 'stopped' | 'failed' | 'starting' | 'running' | 'finished';
    info: any;
    env_vars: { [key: string]: string };
}

export interface Message {
    time: string;
    data: string;
}

export interface ApiCall<T> {
    title: string;
    promise: Promise<T>;
    start: Date;
    end?: Date;
    cancel: () => void;
}

export function getStatus(title: string): ApiCall<Status> {
    const cancelSource = axios.CancelToken.source();
    const result: ApiCall<Status> = {
        title,
        start: new Date(),
        promise: axios
            .get(BASE_URL + '/status', { cancelToken: cancelSource.token })
            .then((res) => res.data)
            .finally(() => (result.end = new Date())),
        cancel: () => cancelSource.cancel('call canceled')
    };
    return result;
}

export function apply(
    title: string,
    enabledComponentIds: string[]
): ApiCall<void> {
    const cancelSource = axios.CancelToken.source();
    const result: ApiCall<any> = {
        title,
        start: new Date(),
        promise: axios
            .post(
                BASE_URL + '/apply',
                { enabled_component_ids: enabledComponentIds },
                { cancelToken: cancelSource.token }
            )
            .finally(() => (result.end = new Date())),
        cancel: () => cancelSource.cancel('call canceled')
    };
    return result;
}

export function startComponent(
    title: string,
    componentId: string
): ApiCall<void> {
    const cancelSource = axios.CancelToken.source();
    const result: ApiCall<any> = {
        title,
        start: new Date(),
        promise: axios
            .post(
                BASE_URL + '/start_component',
                { component_id: componentId },
                { cancelToken: cancelSource.token }
            )
            .finally(() => (result.end = new Date())),
        cancel: () => cancelSource.cancel('call canceled')
    };
    return result;
}

export function stopComponent(
    title: string,
    componentId: string
): ApiCall<void> {
    const cancelSource = axios.CancelToken.source();
    const result: ApiCall<any> = {
        title,
        start: new Date(),
        promise: axios
            .post(
                BASE_URL + '/stop_component',
                { component_id: componentId },
                { cancelToken: cancelSource.token }
            )
            .finally(() => (result.end = new Date())),
        cancel: () => cancelSource.cancel('call canceled')
    };
    return result;
}

export function restartComponent(
    title: string,
    componentId: string
): ApiCall<void> {
    const cancelSource = axios.CancelToken.source();
    const canceler = { canceled: false, cancelSource };
    const result: ApiCall<any> = {
        title,
        start: new Date(),
        promise: callRestart(componentId, canceler).finally(
            () => (result.end = new Date())
        ),
        cancel: () => {
            cancelSource.cancel('call canceled');
            canceler.canceled = true;
        }
    };
    return result;
}

export function stopAll(title: string, cleanup: boolean): ApiCall<void> {
    const cancelSource = axios.CancelToken.source();
    const result: ApiCall<any> = {
        title,
        start: new Date(),
        promise: axios
            .post(
                BASE_URL + '/stop_all',
                { cleanup },
                { cancelToken: cancelSource.token }
            )
            .finally(() => (result.end = new Date())),
        cancel: () => cancelSource.cancel('call canceled')
    };
    return result;
}

export async function getOutput(
    onData: (component: string, messages: Message[]) => void
) {
    const response = await fetch(BASE_URL + '/output');
    const reader = response.body!.getReader();
    let component = '';
    let time = '';
    while (true) {
        const { value, done } = await reader.read();
        if (done) {
            return;
        }
        const string = new TextDecoder('utf-8').decode(value);
        const lines = string.replace(/\r\n/g, '\n').split('\n');
        const byComponent: { [key: string]: Message[] } = {};
        for (const line of lines) {
            if (!line.trim()) {
                continue;
            }
            let data = line;
            if (line.startsWith('<component>')) {
                const timePos = line.indexOf('<time>');
                const msgPos = line.indexOf('<msg>');
                component = line.substring(11, timePos);
                time = line.substring(timePos + 6, msgPos).padEnd(28, '0');
                data = line.substring(msgPos + 5) + '\n';
            }
            if (!byComponent[component]) {
                byComponent[component] = [];
            }
            byComponent[component].push({ time, data });
        }
        for (const component of Object.keys(byComponent)) {
            onData(component, byComponent[component]);
        }
    }
}

interface Canceler {
    canceled: boolean;
    cancelSource: CancelTokenSource;
}

async function callRestart(componentId: string, canceler: Canceler) {
    if (canceler.canceled) {
        return;
    }
    await axios.post(
        BASE_URL + '/stop_component',
        { component_id: componentId },
        { cancelToken: canceler.cancelSource.token }
    );
    if (canceler.canceled) {
        return;
    }
    await axios.post(
        BASE_URL + '/start_component',
        { component_id: componentId },
        { cancelToken: canceler.cancelSource.token }
    );
}
