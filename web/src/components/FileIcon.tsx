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

import React, { type ReactNode } from 'react'
import { ListItemIcon } from '@mui/material'
import FolderIcon from '@mui/icons-material/Folder'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import ImageIcon from '@mui/icons-material/Image'
import DescriptionIcon from '@mui/icons-material/Description'
import FolderZipIcon from '@mui/icons-material/FolderZip'
import MovieIcon from '@mui/icons-material/Movie'
import AudiotrackIcon from '@mui/icons-material/Audiotrack'
import CodeIcon from '@mui/icons-material/Code'
import TableChartIcon from '@mui/icons-material/TableChart'
import NoteIcon from '@mui/icons-material/Note'
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf'

interface FileIconProps extends React.HTMLAttributes<HTMLElement> {
  isDir: boolean
  fileExtension?: string
}

const FileIcon: React.FC<FileIconProps> = ({ isDir, fileExtension, ...props }) => {
  const iconMapping: Record<string, ReactNode> = {
    jpg: <ImageIcon fontSize='inherit' color='primary' />,
    jpeg: <ImageIcon fontSize='inherit' color='primary' />,
    gif: <ImageIcon fontSize='inherit' color='primary' />,
    png: <ImageIcon fontSize='inherit' color='primary' />,
    pdf: <PictureAsPdfIcon fontSize='inherit' color='primary' />,
    doc: <DescriptionIcon fontSize='inherit' color='primary' />,
    docx: <DescriptionIcon fontSize='inherit' color='primary' />,
    xls: <TableChartIcon fontSize='inherit' color='primary' />,
    xlsx: <TableChartIcon fontSize='inherit' color='primary' />,
    csv: <TableChartIcon fontSize='inherit' color='primary' />,
    mp3: <AudiotrackIcon fontSize='inherit' color='primary' />,
    wav: <AudiotrackIcon fontSize='inherit' color='primary' />,
    mp4: <MovieIcon fontSize='inherit' color='primary' />,
    avi: <MovieIcon fontSize='inherit' color='primary' />,
    mov: <MovieIcon fontSize='inherit' color='primary' />,
    mkv: <MovieIcon fontSize='inherit' color='primary' />,
    txt: <NoteIcon fontSize='inherit' color='primary' />,
    md: <NoteIcon fontSize='inherit' color='primary' />,
    json: <CodeIcon fontSize='inherit' color='primary' />,
    js: <CodeIcon fontSize='inherit' color='primary' />,
    jsx: <CodeIcon fontSize='inherit' color='primary' />,
    ts: <CodeIcon fontSize='inherit' color='primary' />,
    tsx: <CodeIcon fontSize='inherit' color='primary' />,
    html: <CodeIcon fontSize='inherit' color='primary' />,
    css: <CodeIcon fontSize='inherit' color='primary' />,
    zip: <FolderZipIcon fontSize='inherit' color='primary' />,
    rar: <FolderZipIcon fontSize='inherit' color='primary' />,
    '7z': <FolderZipIcon fontSize='inherit' color='primary' />
  }

  return (
    <ListItemIcon {...props}>
      {isDir
        ? (
          <FolderIcon fontSize='inherit' color='primary' />
          )
        : (
            (iconMapping[fileExtension ?? '']) ?? <InsertDriveFileIcon fontSize='inherit' color='primary' />
          )}
    </ListItemIcon>
  )
}

export default FileIcon
