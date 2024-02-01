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
import { Paper, type PaperProps, useTheme } from '@mui/material'

interface HoverPaperProps extends PaperProps {
  selected: boolean
  children: React.ReactNode
}

const HoverPaper: React.FC<HoverPaperProps> = ({ selected, children, ...props }) => {
  const theme = useTheme()

  const [raised, setRaised] = useState(false)

  return (
    <Paper
      elevation={raised || selected ? 5 : 1}
      onMouseOver={() => { setRaised(true) }}
      onMouseOut={() => { setRaised(false) }}
      sx={{
        backgroundColor: selected ? theme.palette.primary.light : theme.palette.background.default,
        color: selected ? theme.palette.primary.contrastText : theme.palette.text.primary,
        transition: 'background-color 0.2s ease-in-out, color 0.2s ease-in-out'
      }}
      {...props}
    >
      {children}
    </Paper>
  )
}

export default HoverPaper
