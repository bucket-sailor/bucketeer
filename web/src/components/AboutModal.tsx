/* SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * Copyright 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import React from 'react'
import { Box, Dialog, DialogContent, DialogTitle, Typography, Link, IconButton, useTheme } from '@mui/material'
import CloseIcon from '@mui/icons-material/Close'
import BucketeerLogo from '../assets/bucketeer.svg'

interface AboutModalProps {
  open: boolean
  onClose: () => void
}

const AboutModal: React.FC<AboutModalProps> = ({ open, onClose }) => {
  const theme = useTheme()

  return (
        <Dialog
            open={open}
            onClose={onClose}
            PaperProps={{
              style: { minWidth: theme.spacing(50), userSelect: 'none' }
            }}
        >
            <DialogTitle>
                About Bucketeer
                <IconButton
                    aria-label="close"
                    onClick={onClose}
                    style={{ position: 'absolute', right: theme.spacing(1), top: theme.spacing(1) }}>
                    <CloseIcon />
                </IconButton>
            </DialogTitle>
            <DialogContent>
                <Box textAlign="center" marginBottom={theme.spacing(2)}>
                    <img src={BucketeerLogo} alt="Bucketeer Logo" style={{ width: theme.spacing(12.5), height: theme.spacing(12.5) }} />
                </Box>
                <Typography variant="body1" gutterBottom>
                    <b>Author:</b> <Link href="mailto:damian@pecke.tt">Damian Peckett</Link>
                </Typography>
                <Typography variant="body1" gutterBottom>
                    <b>Version:</b> v0.1.0
                </Typography>
                <Typography variant="body1" gutterBottom>
                    <b>GitHub:</b> <Link href="https://github.com/bucket-sailor/bucketeer" target="_blank">bucket-sailor/bucketeer</Link>
                </Typography>
                <Typography variant="body1">
                    <b>License:</b> <Link href="http://github.com/bucket-sailor/bucketeer/blob/main/LICENSE" target="_blank">AGPL-3.0-or-later</Link>
                </Typography>
            </DialogContent>
        </Dialog>
  )
}

export default AboutModal
