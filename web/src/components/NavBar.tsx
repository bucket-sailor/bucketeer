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

import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Typography } from '@mui/material'
import BucketeerLogo from '../assets/bucketeer.svg'

import styles from './NavBar.module.css'

interface NavBarProps {
  sideBarCollapsed: boolean
  basePath: string
  currentDirectory?: string
}

const NavBar: React.FC<NavBarProps> = ({ sideBarCollapsed, basePath, currentDirectory }) => {
  const navigate = useNavigate()
  const [pathParts, setPathParts] = useState<string[]>([])

  useEffect(() => {
    if (currentDirectory !== undefined) {
      setPathParts(currentDirectory.split('/').filter((part) => part !== ''))
    }
  }, [currentDirectory])

  const handleNavigateToPathPart = (index: number): void => {
    const path = basePath + pathParts.slice(0, index + 1).join('/')
    navigate(path + '/')
  }

  return (
        <Box className={styles.navbar}>
            <Box className={styles.navbox} sx={{ marginLeft: !sideBarCollapsed ? '260px' : 'auto' }}>
                <Box className={styles.logoContainer} onClick={() => { navigate(basePath) }}>
                    <img src={BucketeerLogo} alt="Bucketeer Logo" style={{ paddingRight: '4px', height: '0.8em' }} />
                    <Typography variant="body1">Bucketeer</Typography>
                </Box>
                {pathParts.length > 0 && <Typography variant="body1" className={styles.separator}>&gt;</Typography>}

                {pathParts.map((part, index) => (
                    <Box key={index} sx={{ display: 'flex', justifyContent: 'start', alignItems: 'center' }}>
                      <Typography
                          variant="body1"
                          className={styles.entry}
                          onClick={() => { handleNavigateToPathPart(index) }}>
                          {part}
                      </Typography>
                      {index < pathParts.length - 1 && <Typography variant="body1" className={styles.separator}>&gt;</Typography>}
                    </Box>
                ))}
            </Box>
        </Box>
  )
}

export default NavBar
