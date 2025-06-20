// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React from 'react';
import './ProductTour.css';
import Button from '@mui/material/Button';

export interface ProductTourProps {
    open: boolean;
    close: () => void;
}

function ProductTour(props: ProductTourProps) {
    if (!props.open) {
        return null;
    }
    return (
        <div className="ProductTour">
            <div className="content">
                Hi there 👋
                <br />
                <br />
                Just wanted to make sure you know you can use the Apply button
                on the bottom left.
                <br />
                <span className="marked">Applying state</span> is usually more
                friendly than manipulating components one by one.
                <br />
                <br />
                When you apply state, ENVITE takes care of the dependencies
                between components by loading them in the required order and
                monitoring their status. It will also stop and re-run components
                to your request.
                <br />
                <br />
                If you haven't already, give it a try. It's usually faster and
                easier.
                <div className="actions">
                    <Button onClick={props.close} variant="outlined">
                        Dismiss
                    </Button>
                </div>
            </div>
        </div>
    );
}

export default ProductTour;
