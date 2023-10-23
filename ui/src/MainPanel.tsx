import React from 'react';
import './MainPanel.css';
import { Message, Status } from './api';
import SplitPanel from './SplitPanel';
import ComponentPanel from './ComponentPanel';
import { Route, Routes } from 'react-router-dom';

interface MainPanelProps {
    loading: boolean;
    status: Status;
    output: { [key: string]: Message[] };
    clearOutput: (componentId: string) => void;
    clearAllOutput: () => void;
    startComponent: (id: string) => void;
    stopComponent: (id: string) => void;
    restartComponent: (id: string) => void;
}

function MainPanel(props: MainPanelProps) {
    const components = props.status.components.flat() || [];
    return (
        <div className="MainPanel">
            <Routes>
                {components.map((s) => (
                    <Route
                        key={s.id}
                        path={`/${s.id}/*`}
                        element={
                            <ComponentPanel
                                loading={props.loading}
                                component={s}
                                data={props.output[s.id]}
                                clear={() => props.clearOutput(s.id)}
                                startComponent={() =>
                                    props.startComponent(s.id)
                                }
                                stopComponent={() => props.stopComponent(s.id)}
                                restartComponent={() =>
                                    props.restartComponent(s.id)
                                }
                            />
                        }
                    />
                ))}
                <Route
                    path="/*"
                    element={
                        <SplitPanel
                            status={props.status}
                            output={props.output}
                            clearOutput={props.clearOutput}
                            clearAllOutput={props.clearAllOutput}
                        />
                    }
                />
            </Routes>
        </div>
    );
}

export default MainPanel;
