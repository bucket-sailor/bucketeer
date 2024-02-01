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

import React, { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ListItemIcon, useTheme, useMediaQuery, Menu, MenuItem } from '@mui/material'
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline'
import DownloadOutlinedIcon from '@mui/icons-material/DownloadOutlined'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { useFileManagement } from './hooks/FileManagement'
import type { FileInfo } from './api/client'
import NavBar from './components/NavBar'
import SideBar from './components/SideBar'
import ConfirmDeleteModal from './components/ConfirmDeleteModal'
import FilePropertiesModal from './components/FilePropertiesModal'
import FileGrid from './components/FileGrid'

import styles from './App.module.css'

// defined in vite.config.ts
const basePath = '/browse/'

const App = (): React.ReactElement => {
  const params = useParams()
  const navigate = useNavigate()

  const theme = useTheme()
  const smallScreen = useMediaQuery(theme.breakpoints.down('md'))

  const backendBaseURL = 'http://localhost:8082'
  const { directoryContents, resetFilesState, loadFiles, uploadFile, downloadFile, getFileInfo, deleteFile, makeDirectory } = useFileManagement(backendBaseURL)

  const [currentDirectory, setCurrentDirectory] = useState<string | undefined>(undefined)
  const [selectedFile, setSelectedFile] = useState<string | undefined>(undefined)
  const [selectedFileInfo, setSelectedFileInfo] = useState<FileInfo | undefined>(undefined)

  const [fileMenuOpen, setFileMenuOpen] = useState(false)
  const [fileMenuAnchorEl, setFileMenuAnchorEl] = useState<HTMLElement | null>(null)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [isPropertiesModalOpen, setIsPropertiesModalOpen] = useState(false)

  const handleFileClick = (fileName: string): void => {
    setSelectedFile((prevSelectedFile) => {
      if (prevSelectedFile === fileName) {
        return undefined
      }
      return fileName
    })
  }

  const handleFileDoubleClick = useCallback((fileName: string): void => {
    const file = directoryContents.find((item) => item.name === fileName)
    if (file !== undefined && file.isDir) {
      let currentPath = (params['*'] ?? '')
      if (!currentPath.endsWith('/')) {
        currentPath += '/'
      }

      navigate(`${currentPath}${fileName}`)
    }
  }, [directoryContents, params, navigate])

  const handleFileMenuOpen = (anchorEl: HTMLElement, fileName: string): void => {
    setFileMenuAnchorEl(anchorEl)
    setFileMenuOpen(true)
    setSelectedFile(fileName)
  }

  const handleFileMenuClose = (): void => {
    setFileMenuAnchorEl(null)
    setFileMenuOpen(false)
  }

  const handleFileDownload = useCallback((): void => {
    handleFileMenuClose()

    if (currentDirectory !== undefined && selectedFile !== undefined) {
      downloadFile(currentDirectory + '/' + selectedFile).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory, selectedFile])

  const handleOpenDeleteModal = (): void => {
    setFileMenuOpen(false)
    setIsDeleteModalOpen(true)
  }

  const handleCloseDeleteModal = (): void => {
    setIsDeleteModalOpen(false)
  }

  const handleConfirmDelete = useCallback((): void => {
    handleCloseDeleteModal()

    if (currentDirectory !== undefined && selectedFile !== undefined) {
      deleteFile(currentDirectory + '/' + selectedFile).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory, selectedFile])

  const handleOpenPropertiesModal = useCallback((): void => {
    setFileMenuOpen(false)
    setIsPropertiesModalOpen(true)

    if (currentDirectory !== undefined && selectedFile !== undefined) {
      getFileInfo(currentDirectory + '/' + selectedFile).then((response) => {
        if (response !== undefined) {
          setSelectedFileInfo(response)
        }
      }).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory, selectedFile])

  const handleClosePropertiesModal = (): void => {
    setIsPropertiesModalOpen(false)
    setSelectedFileInfo(undefined)
  }

  const handleCreateDirectory = useCallback((directoryName: string): void => {
    if (currentDirectory !== undefined) {
      makeDirectory(currentDirectory + '/' + directoryName).then(() => { }).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory])

  const handleFileUpload = useCallback((): void => {
    if (currentDirectory !== undefined) {
      uploadFile(currentDirectory).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory])

  useEffect(() => {
    // Reset the pagination state (and abort any pending requests).
    resetFilesState()

    // clear the selected file.
    setSelectedFile(undefined)

    // Update the current directory.
    setCurrentDirectory(removePrefix(requirePrefix(params['*'] ?? '', '/'), basePath))
  }, [params])

  return (
    <>
      <NavBar smallScreen={smallScreen} basePath={basePath} currentDirectory={currentDirectory} />
      <div className={styles.content}>
        <SideBar
          smallScreen={smallScreen}
          basePath={basePath}
          onCreateDirectory={handleCreateDirectory}
          onUploadFile={handleFileUpload}
        />

        <Menu
          id="file-menu"
          anchorEl={fileMenuAnchorEl}
          open={fileMenuOpen}
          onClose={handleFileMenuClose}
          MenuListProps={{
            'aria-labelledby': 'file-menu-button'
          }}>
          <MenuItem onClick={handleFileDownload}>
            <ListItemIcon>
              <DownloadOutlinedIcon />
            </ListItemIcon>
            Download
          </MenuItem>

          <MenuItem onClick={handleOpenDeleteModal}>
            <ListItemIcon>
              <DeleteOutlineIcon />
            </ListItemIcon>
            Delete
          </MenuItem>

          <MenuItem onClick={handleOpenPropertiesModal}>
            <ListItemIcon>
              <InfoOutlinedIcon />
            </ListItemIcon>
            Properties
          </MenuItem>
        </Menu>

        <FileGrid
          smallScreen={smallScreen}
          currentDirectory={currentDirectory}
          directoryContents={directoryContents}
          selectedFile={selectedFile}
          onFileClick={handleFileClick}
          onFileDoubleClick={handleFileDoubleClick}
          onFileMenuOpen={handleFileMenuOpen}
          loadFiles={loadFiles}
        />
      </div >

      <ConfirmDeleteModal
        open={isDeleteModalOpen}
        onClose={handleCloseDeleteModal}
        onConfirm={handleConfirmDelete}
      />

      <FilePropertiesModal
        open={isPropertiesModalOpen}
        onClose={handleClosePropertiesModal}
        fileInfo={selectedFileInfo}
      />
    </>
  )
}

const removePrefix = (value: string, prefix: string): string =>
  value.startsWith(prefix) ? value.slice(prefix.length) : value

const requirePrefix = (value: string, prefix: string): string => {
  if (!value.startsWith(prefix)) {
    return prefix + value
  }
  return value
}

export default App
