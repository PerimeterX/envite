import React, { useEffect, useState } from 'react';
import './OfflinePanel.css';
import WifiOffIcon from '@mui/icons-material/WifiOff';
import Loader from './Loader';

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
                        <WifiOffIcon fontSize="inherit" />
                    </div>
                    <div>You seem offline</div>
                    <div>Check if Feng Shui server is up and running...</div>
                </>
            )}
        </div>
    );
}

export default OfflinePanel;
