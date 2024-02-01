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

import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Divider, Drawer, IconButton, List, ListItemButton, ListItemIcon, ListItemText, Typography, useTheme } from '@mui/material'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import CreateNewFolderOutlinedIcon from '@mui/icons-material/CreateNewFolderOutlined'
import UploadFileOutlinedIcon from '@mui/icons-material/UploadFileOutlined'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import CreateDirectoryModal from './CreateDirectoryModal'
import AboutModal from './AboutModal'
import BucketeerLogo from '../assets/bucketeer.svg'

import styles from './SideBar.module.css'

interface SideBarProps extends React.HTMLAttributes<HTMLElement> {
  smallScreen: boolean
  basePath: string
  onCreateDirectory: (directoryName: string) => void
  onUploadFile: () => void
}

const SideBar: React.FC<SideBarProps> = ({ smallScreen, basePath, onCreateDirectory, onUploadFile, ...props }) => {
  const theme = useTheme()
  const navigate = useNavigate()

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [isCreateDirectoryModalOpen, setIsCreateDirectoryModalOpen] = useState(false)
  const [isAboutModalOpen, setIsAboutModalOpen] = useState(false)

  return (
        <>
            {smallScreen && !drawerOpen && (
                <IconButton className={styles.openButton} onClick={() => { setDrawerOpen(true) }}>
                    <ChevronRightIcon />
                </IconButton>
            )}

            <Drawer
                variant={smallScreen ? 'temporary' : 'persistent'}
                anchor="left"
                open={smallScreen ? drawerOpen : true}
                onClose={() => { setDrawerOpen(false) }}
                className={styles.sidebar}

                PaperProps={{
                  className: styles.sidebarPaper,
                  sx: {
                    marginTop: smallScreen ? '0' : '46px'
                  }
                }}
                {...props} >
                {smallScreen && (
                    <IconButton className={styles.closeButton} onClick={() => { setDrawerOpen(false) }}>
                        <ChevronLeftIcon />
                    </IconButton>
                )}

                <List>
                    <ListItemButton onClick={() => { navigate(basePath) }}>
                        <img src={BucketeerLogo} alt="Bucketeer Logo" className={styles.logo} />
                        <Typography variant="h6" style={{ marginLeft: 10 }}>
                            Bucketeer
                        </Typography>
                    </ListItemButton>

                    <Divider sx={{ marginTop: theme.spacing(1), marginBottom: theme.spacing(1) }} />

                    <nav aria-label="file and directory operations">
                        <ListItemButton onClick={() => { setIsCreateDirectoryModalOpen(true) }}>
                            <ListItemIcon>
                                <CreateNewFolderOutlinedIcon />
                            </ListItemIcon>
                            <ListItemText primary="New Folder" />
                        </ListItemButton>

                        <ListItemButton onClick={onUploadFile}>
                            <ListItemIcon>
                                <UploadFileOutlinedIcon />
                            </ListItemIcon>
                            <ListItemText primary="Upload file" />
                        </ListItemButton>
                    </nav>

                    <Divider sx={{ marginTop: theme.spacing(1), marginBottom: theme.spacing(1) }} />

                    <nav aria-label="about">
                        <ListItemButton onClick={() => { setIsAboutModalOpen(true) }}>
                            <ListItemIcon>
                                <InfoOutlinedIcon />
                            </ListItemIcon>
                            <ListItemText primary="About" />
                        </ListItemButton>
                    </nav>
                </List>
            </Drawer>

            <CreateDirectoryModal
                open={isCreateDirectoryModalOpen}
                onClose={() => { setIsCreateDirectoryModalOpen(false) }}
                onCreate={(directoryName): void => {
                  setIsCreateDirectoryModalOpen(false)
                  onCreateDirectory(directoryName)
                }}
            />

            <AboutModal
                open={isAboutModalOpen}
                onClose={() => { setIsAboutModalOpen(false) }}
            />
        </>
  )
}

export default SideBar
