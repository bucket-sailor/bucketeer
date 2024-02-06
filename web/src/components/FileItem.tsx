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

import React, { useRef } from 'react'
import { Grid, IconButton, Typography, useTheme } from '@mui/material'
import AutoSizer from 'react-virtualized-auto-sizer'
import MoreVertIcon from '@mui/icons-material/MoreVert'
import FileIcon from './FileIcon'
import HoverPaper from './HoverPaper'
import TruncateText from './TruncateText'

import styles from './FileItem.module.css'

interface FileItemProps extends React.HTMLAttributes<HTMLElement> {
  isDir: boolean
  fileName: string
  selected: boolean
  onFileClick: (fileName: string) => void
  onFileDoubleClick: (fileName: string) => void
  onFileMenuOpen: (anchorEl: HTMLElement, fileName: string) => void
}

const FileItem: React.FC<FileItemProps> = ({ isDir, fileName, selected, onFileClick, onFileDoubleClick, onFileMenuOpen, ...props }: FileItemProps) => {
  const theme = useTheme()
  const containerRef = useRef<HTMLElement | null>(null)

  return (
    <HoverPaper
        selected={selected}
        onClick={() => { onFileClick(fileName) } }
        onDoubleClick={() => { onFileDoubleClick(fileName) } }
        {...props}>
      <AutoSizer>
        {({ width }) =>
          <Grid container direction="column" alignItems="center" sx={{ width }}>
            <Grid item xs={12} style={{ width: '100%', display: 'flex', justifyContent: 'end', userSelect: 'none' }}>
              <IconButton
                onClick={(event) => {
                  onFileMenuOpen(event.currentTarget, fileName)
                  event.stopPropagation()
                }}
              >
                <MoreVertIcon />
              </IconButton>
            </Grid>
            <Grid item xs={12}>
              <FileIcon className={styles.icon} isDir={isDir} fileExtension={fileName.split('.').pop()} />
            </Grid>
            <Grid item xs={12}>
              <Typography ref={containerRef} variant="body2" align="center" className={styles.fileNameText} sx={{ width: `${width - dimToNumber(theme.spacing(10))}px` }}>
                <TruncateText text={fileName} containerEl={containerRef.current} containerHeight={42 /* enough for two lines */} containerWidth={width - dimToNumber(theme.spacing(10))} />
              </Typography>
            </Grid>
          </Grid>
          }
        </AutoSizer>
      </HoverPaper>
  )
}

const dimToNumber = (dim: string): number => {
  return parseInt(dim.replace('px', ''))
}

export default FileItem
