// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React from 'react';
import './Loader.css';

interface LoaderProps {
    type: 'main' | 'secondary' | 'central';
}

function Loader(props: LoaderProps) {
    return <span className={`Loader ${props.type}`} />;
}

export default Loader;
