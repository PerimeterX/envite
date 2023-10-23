import React from 'react';
import './Loader.css';

interface LoaderProps {
    type: 'main' | 'secondary' | 'central';
}

function Loader(props: LoaderProps) {
    return <span className={`Loader ${props.type}`} />;
}

export default Loader;
