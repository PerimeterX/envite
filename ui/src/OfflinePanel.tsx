// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useEffect, useState } from 'react';
import './OfflinePanel.css';
import Loader from './Loader';
import { getUrl } from './api';

function OfflinePanel() {
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        setTimeout(() => setLoading(false), 3000);
    }, []);

    return (
        <div className="OfflinePanel">
            {loading ? (
                <Loader type="central"></Loader>
            ) : (
                <>
                    <div className="icon">
                        <img src="/logo-large.svg" alt="ENVITE Icon - large" />
                    </div>
                    <div>ENVITE is not reachable at {getUrl()}</div>
                    <div>Make sure ENVITE server is running</div>
                </>
            )}
        </div>
    );
}

export default OfflinePanel;
