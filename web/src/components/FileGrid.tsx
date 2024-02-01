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

import React, { createContext, useCallback, useContext, useEffect } from 'react'
import { FixedSizeGrid as Grid, type GridChildComponentProps } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'
import InfiniteLoader from 'react-window-infinite-loader'
import { Box } from '@mui/material'
import { type FileInfo } from '../api/client'
import FileItem from './FileItem'

import styles from './FileGrid.module.css'

interface FileGridProps extends React.HTMLAttributes<HTMLElement> {
  smallScreen: boolean
  currentDirectory?: string
  directoryContents: FileInfo[]
  selectedFile?: string
  onFileClick: (fileName: string) => void
  onFileDoubleClick: (fileName: string) => void
  onFileMenuOpen: (anchorEl: HTMLElement, fileName: string) => void
  loadFiles: (path: string, startIndex: number, stopIndex: number) => Promise<boolean>
}

const FileGrid: React.FC<FileGridProps> = ({ smallScreen, currentDirectory, directoryContents, selectedFile, onFileClick, onFileDoubleClick, onFileMenuOpen, loadFiles, ...props }: FileGridProps) => {
  // We don't know ahead of time how many files there are.
  const itemCountInfinity = 1000000

  useEffect(() => {
    // Kick off the initial load.
    if (currentDirectory !== undefined) {
      loadMoreItems(0, 20).then(() => { }).catch((error) => {
        console.error(error)
      })
    }
  }, [currentDirectory])

  const isItemLoaded = useCallback((index: number): boolean => {
    return directoryContents.length !== 0 && index < directoryContents.length
  }, [directoryContents])

  const loadMoreItems = useCallback(async (startIndex: number, stopIndex: number): Promise<void> => {
    if (currentDirectory !== undefined) {
      await loadFiles(currentDirectory, startIndex, stopIndex)
    }
  }, [currentDirectory])

  return (
    <Box className={styles.container} {...props}>
      <AutoSizer>
        {({ height, width }) => {
          const desiredColumnWidth = 200

          let columnCount = Math.max(Math.floor(width / desiredColumnWidth), 1)
          let totalGuttersWidth = width - (columnCount * desiredColumnWidth)
          let gutterSize = columnCount > 1 ? totalGuttersWidth / (columnCount - 1) : 0

          if (!smallScreen && gutterSize < 12) {
            columnCount--

            totalGuttersWidth = width - (columnCount * desiredColumnWidth)
            gutterSize = columnCount > 1 ? totalGuttersWidth / (columnCount - 1) : 0
          }

          return (
            <InfiniteLoader
              isItemLoaded={isItemLoaded}
              itemCount={itemCountInfinity}
              loadMoreItems={loadMoreItems}
            >
              {({ onItemsRendered, ref }) => (
                <FileGridProvider value={{ columnCount, gutterSize, directoryContents, selectedFile, onFileClick, onFileDoubleClick, onFileMenuOpen }}>
                  <Grid
                    ref={ref}
                    width={width}
                    height={height}
                    columnCount={columnCount}
                    columnWidth={(width - gutterSize) / columnCount}
                    rowCount={Math.ceil(directoryContents.length / columnCount) + 1}
                    rowHeight={195 + gutterSize}
                    onItemsRendered={({ visibleRowStartIndex, visibleRowStopIndex, visibleColumnStartIndex, visibleColumnStopIndex, overscanRowStartIndex, overscanRowStopIndex }) => {
                      const startIdx = visibleRowStartIndex * columnCount + visibleColumnStartIndex
                      const stopIdx = visibleRowStopIndex * columnCount + visibleColumnStopIndex
                      onItemsRendered({
                        overscanStartIndex: overscanRowStartIndex * columnCount,
                        overscanStopIndex: (overscanRowStopIndex + 1) * columnCount - 1,
                        visibleStartIndex: startIdx,
                        visibleStopIndex: stopIdx
                      })
                    }}>
                    {FileGridCell}
                  </Grid>
                </FileGridProvider>
              )}
            </InfiniteLoader>
          )
        }}
      </AutoSizer>
    </Box >
  )
}

const FileGridCell: React.FC<GridChildComponentProps> = ({
  columnIndex,
  rowIndex,
  style
}: GridChildComponentProps) => {
  const { columnCount, gutterSize, directoryContents, selectedFile, onFileClick, onFileDoubleClick, onFileMenuOpen } = useContext(FileGridContext)

  const fileInfo = directoryContents[rowIndex * columnCount + columnIndex]
  if (fileInfo === undefined) {
    return (<div style={style} />)
  }

  return (
    <FileItem
      key={fileInfo.name}
      isDir={fileInfo.isDir}
      fileName={fileInfo.name}
      selected={fileInfo.name === selectedFile}
      onFileClick={onFileClick}
      onFileDoubleClick={onFileDoubleClick}
      onFileMenuOpen={onFileMenuOpen}
      style={{
        ...style,
        left: style.left as number + gutterSize,
        top: style.top as number + gutterSize,
        width: style.width as number - gutterSize,
        height: style.height as number - gutterSize
      }}
    />
  )
}

interface FileGridContextValue {
  columnCount: number
  gutterSize: number
  directoryContents: FileInfo[]
  selectedFile?: string
  onFileClick: (fileName: string) => void
  onFileDoubleClick: (fileName: string) => void
  onFileMenuOpen: (anchorEl: HTMLElement, fileName: string) => void
}

interface FileGridProviderProps {
  children: React.ReactNode
  value: FileGridContextValue
}

// Workaround for some questionable design decisions in react-window.
const FileGridContext = createContext<FileGridContextValue>({
  columnCount: 1,
  gutterSize: 0,
  directoryContents: [],
  onFileClick: () => { },
  onFileDoubleClick: () => { },
  onFileMenuOpen: () => { }
})

const FileGridProvider: React.FC<FileGridProviderProps> = ({ children, value }) => {
  return <FileGridContext.Provider value={value}>{children}</FileGridContext.Provider>
}

export default FileGrid
