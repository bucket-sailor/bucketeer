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
import { Dialog, DialogContent, DialogTitle, IconButton, List, ListItem, ListItemText, Typography, useTheme } from '@mui/material'
import { type FileInfo } from '../gen/filesystem/v1alpha1/filesystem_pb'
import CloseIcon from '@mui/icons-material/Close'

interface FilePropertiesModalProps extends React.HTMLAttributes<HTMLElement> {
  open: boolean
  onClose: () => void
  fileInfo?: FileInfo
}

const FilePropertiesModal: React.FC<FilePropertiesModalProps> = ({ open, onClose, fileInfo, ...props }) => {
  const theme = useTheme()

  return (
    <Dialog
      open={open}
      onClose={onClose}
      {...props}
      PaperProps={{
        style: { minWidth: theme.spacing(50), userSelect: 'none' }
      }}
    >
      <DialogTitle>
        File Properties
        <IconButton
          aria-label="close"
          onClick={onClose}
          style={{ position: 'absolute', right: 8, top: 8 }}>
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <List>
          {fileInfo !== undefined && <>
            <ListItem>
              <ListItemText
                primary={<Typography variant="subtitle2">Name</Typography>}
                secondary={fileInfo.name}
              />
            </ListItem>
            <ListItem>
              <ListItemText
                primary={<Typography variant="subtitle2">Last Modified</Typography>}
                secondary={fileInfo?.modTime?.toDate()?.toLocaleString()}
              />
            </ListItem>
            <ListItem>
              <ListItemText
                primary={<Typography variant="subtitle2">Size</Typography>}
                secondary={Number(fileInfo.size)}
              />
            </ListItem>
          </>}
        </List>
      </DialogContent>
    </Dialog>
  )
}

export default FilePropertiesModal
