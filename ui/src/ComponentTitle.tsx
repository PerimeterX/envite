import React from 'react';
import './ComponentTitle.css';
import { Component } from './api';
import { statusToUiParams } from './ComponentBox';
import { NavLink } from 'react-router-dom';

interface ComponentTitleProps {
    component: Component;
}

function ComponentTitle(props: ComponentTitleProps) {
    const Icon = statusToUiParams[props.component.status].icon;
    return (
        <div className="ComponentTitle">
            <NavLink
                to={`${props.component.id}`}
                className="component-title-link"
            >
                <div className="component-title">{props.component.id}</div>
            </NavLink>
            <div>
                <span
                    className={`status ${
                        props.component.status === 'stopped' ||
                        props.component.status === 'starting'
                            ? 'idle'
                            : props.component.status === 'failed'
                            ? 'failed'
                            : ''
                    }`}
                >
                    {props.component.status}&nbsp;
                    <Icon className="icon" fontSize="inherit" />
                </span>
            </div>
        </div>
    );
}

export default ComponentTitle;
