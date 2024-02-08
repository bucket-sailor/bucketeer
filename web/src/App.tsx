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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type InfiniteLoader from 'react-window-infinite-loader'
import { Alert, Box, ListItemIcon, Menu, MenuItem, useMediaQuery, useTheme } from '@mui/material'
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline'
import DownloadOutlinedIcon from '@mui/icons-material/DownloadOutlined'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { useFileManagement } from './hooks/FileManagement'
import { type FileInfo } from './gen/filesystem/v1alpha1/filesystem_pb'
import NavBar from './components/NavBar'
import SideBar from './components/SideBar'
import ConfirmDeleteModal from './components/ConfirmDeleteModal'
import FilePropertiesModal from './components/FilePropertiesModal'
import FileGrid from './components/FileGrid'
import { useTelemetry } from './hooks/Telemetry'
import { generateID } from './util/GenerateID'

import styles from './App.module.css'

interface AppProps {
  baseURL: string
  basePath: string
}

const App = ({ baseURL, basePath }: AppProps): React.ReactElement => {
  const params = useParams()
  const navigate = useNavigate()

  const theme = useTheme()
  const sideBarCollapsed = useMediaQuery(theme.breakpoints.down('md'))

  const sessionID = useRef<string>(generateID(16))

  const {
    registerErrorHandlers,
    reportEvent
  } = useTelemetry({ baseURL, sessionID: sessionID.current })

  const fileGridLoaderRef = useRef<InfiniteLoader | null>(null)

  const {
    directoryContents,
    error: fileManagementError,
    refreshFiles,
    loadFiles,
    getFileInfo,
    deleteFile,
    makeDirectory,
    uploadFile,
    downloadFile
  } = useFileManagement({ baseURL, fileGridLoaderRef })

  const [currentDirectory, setCurrentDirectory] = useState<string | undefined>(undefined)
  const [selectedFile, setSelectedFile] = useState<string | undefined>(undefined)
  const [selectedFileInfo, setSelectedFileInfo] = useState<FileInfo | undefined>(undefined)

  const [fileMenuOpen, setFileMenuOpen] = useState(false)
  const [fileMenuAnchorEl, setFileMenuAnchorEl] = useState<HTMLElement | null>(null)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [isPropertiesModalOpen, setIsPropertiesModalOpen] = useState(false)

  // For statistical purposes.
  const isMobile = useMediaQuery('(max-width:600px)')
  const isTablet = useMediaQuery('(min-width:601px) and (max-width:900px)')
  const isHighDPI = useMediaQuery('(-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi)')

  useEffect(() => {
    let deviceType = isMobile ? 'Mobile' : isTablet ? 'Tablet' : 'Desktop'
    if (isHighDPI) {
      deviceType += ' High DPI'
    }

    reportEvent({
      sessionId: sessionID.current,
      name: 'SessionStart',
      values: {
        userAgent: navigator.userAgent,
        deviceType
      }
    }).catch((e) => {
      console.error('Could not register session:', e)
    })

    const deregisterErrorHandlers = registerErrorHandlers()

    return () => {
      deregisterErrorHandlers()

      reportEvent({
        sessionId: sessionID.current,
        name: 'SessionStop'
      }).catch((e) => {
        console.error('Could not deregister session:', e)
      })
    }
  }, [])

  const handleFileClick = (fileName: string): void => {
    setSelectedFile((prevSelectedFile) => {
      if (prevSelectedFile === fileName) {
        return undefined
      }
      return fileName
    })
  }

  const handleFileDoubleClick = useCallback((fileName: string): void => {
    let file: FileInfo | undefined
    directoryContents.forEach((value, _) => {
      if (value.name === fileName) {
        file = value
      }
    })

    if (file !== undefined) {
      if (file.isDir) {
        let currentPath = (params['*'] ?? '')
        if (!currentPath.endsWith('/')) {
          currentPath += '/'
        }

        navigate(`${currentPath}${fileName}/`)
      } else {
        const path = (currentDirectory !== '' ? currentDirectory + '/' : '') + fileName
        downloadFile(path)
      }
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
      const path = (currentDirectory !== '' ? currentDirectory + '/' : '') + selectedFile
      downloadFile(path)
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
      const path = (currentDirectory !== '' ? currentDirectory + '/' : '') + selectedFile
      deleteFile(path).catch((e) => {
        console.error(e)
      })
    }
  }, [currentDirectory, selectedFile])

  const handleOpenPropertiesModal = useCallback((): void => {
    setFileMenuOpen(false)
    setIsPropertiesModalOpen(true)

    if (currentDirectory !== undefined && selectedFile !== undefined) {
      const path = (currentDirectory !== '' ? currentDirectory + '/' : '') + selectedFile
      getFileInfo(path).then((response) => {
        if (response !== undefined) {
          setSelectedFileInfo(response)
        }
      }).catch((e) => {
        console.error(e)
      })
    }
  }, [currentDirectory, selectedFile])

  const handleClosePropertiesModal = (): void => {
    setIsPropertiesModalOpen(false)
    setSelectedFileInfo(undefined)
  }

  const handleCreateDirectory = useCallback((directoryName: string): void => {
    if (currentDirectory !== undefined) {
      const path = (currentDirectory !== '' ? currentDirectory + '/' : '') + directoryName
      makeDirectory(path).then(() => {}).catch((e) => {
        console.error(e)
      })
    }
  }, [currentDirectory])

  const handleFileUpload = useCallback((): void => {
    if (currentDirectory !== undefined) {
      uploadFile(currentDirectory).catch((e) => {
        console.error(e)
      })
    }
  }, [currentDirectory])

  // Catch navigation events and update the current directory.
  useEffect(() => {
    // clear the selected file.
    setSelectedFile(undefined)

    // Update the current directory.
    setCurrentDirectory(removePrefix(requirePrefix(params['*'] ?? '', '/'), basePath))
  }, [params])

  return (
    <>
      <NavBar sideBarCollapsed={sideBarCollapsed} basePath={basePath} currentDirectory={currentDirectory} />
      <Box className={styles.content}>
        <SideBar
          sideBarCollapsed={sideBarCollapsed}
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

        <Box className={styles.fileGrid}>
          {fileManagementError !== undefined && (
            <Alert severity="error" className={styles.errorAlert}>{fileManagementError.message}</Alert>
          )}

          <FileGrid
            ref={fileGridLoaderRef}
            sideBarCollapsed={sideBarCollapsed}
            currentDirectory={currentDirectory}
            directoryContents={directoryContents}
            selectedFile={selectedFile}
            refreshFiles={refreshFiles}
            onFileClick={handleFileClick}
            onFileDoubleClick={handleFileDoubleClick}
            onFileMenuOpen={handleFileMenuOpen}
            loadFiles={loadFiles}
          />
        </Box>
      </Box >

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
