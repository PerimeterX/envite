// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import React, { useRef, useState } from 'react';
import Button from '@mui/material/Button';
import ButtonGroup from '@mui/material/ButtonGroup';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Grow from '@mui/material/Grow';
import Paper from '@mui/material/Paper';
import Popper from '@mui/material/Popper';
import MenuItem from '@mui/material/MenuItem';
import MenuList from '@mui/material/MenuList';
import { OverridableStringUnion } from '@mui/types';
import { ButtonGroupPropsVariantOverrides } from '@mui/material/ButtonGroup/ButtonGroup';

interface SplitButtonProps {
    variant: OverridableStringUnion<
        'text' | 'outlined' | 'contained',
        ButtonGroupPropsVariantOverrides
    >;

    options: ButtonOption[];
}

interface ButtonOption {
    button: React.JSX.Element;
    title: string;
}

export default function SplitButton(props: SplitButtonProps) {
    const [open, setOpen] = useState(false);
    const [selectedIndex, setSelectedIndex] = useState(0);
    const anchorRef = useRef<HTMLDivElement>(null);

    const handleClose = (event: Event) => {
        if (
            anchorRef.current &&
            anchorRef.current.contains(event.target as HTMLElement)
        ) {
            return;
        }

        setOpen(false);
    };

    return (
        <>
            <ButtonGroup variant={props.variant} ref={anchorRef}>
                {props.options[selectedIndex].button}
                <Button
                    size="small"
                    onClick={() => setOpen((prevOpen) => !prevOpen)}
                >
                    <ArrowDropDownIcon />
                </Button>
            </ButtonGroup>
            <Popper
                sx={{ zIndex: 1 }}
                open={open}
                anchorEl={anchorRef.current}
                transition
                disablePortal
            >
                {({ TransitionProps, placement }) => (
                    <Grow
                        {...TransitionProps}
                        style={{
                            transformOrigin:
                                placement === 'bottom'
                                    ? 'center top'
                                    : 'center bottom'
                        }}
                    >
                        <Paper>
                            <ClickAwayListener onClickAway={handleClose}>
                                <MenuList id="split-button-menu" autoFocusItem>
                                    {props.options.map((option, index) => (
                                        <MenuItem
                                            key={option.title}
                                            // disabled={index === 2}
                                            selected={index === selectedIndex}
                                            onClick={() => {
                                                setSelectedIndex(index);
                                                setOpen(false);
                                            }}
                                        >
                                            {option.title}
                                        </MenuItem>
                                    ))}
                                </MenuList>
                            </ClickAwayListener>
                        </Paper>
                    </Grow>
                )}
            </Popper>
        </>
    );
}
