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

import React, { type ForwardedRef, createContext, forwardRef, useCallback, useContext, useEffect, useMemo } from 'react'
import { FixedSizeGrid as Grid, type GridChildComponentProps } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'
import InfiniteLoader from 'react-window-infinite-loader'
import { Box } from '@mui/material'
import FileItem from './FileItem'

import styles from './FileGrid.module.css'
import { type FileInfo } from '../api/client'

interface FileGridProps extends React.HTMLAttributes<HTMLElement> {
  sideBarCollapsed: boolean
  currentDirectory?: string
  directoryContents: Map<number, FileInfo>
  selectedFile?: string
  refreshFiles: () => void
  onFileClick: (fileName: string) => void
  onFileDoubleClick: (fileName: string) => void
  onFileMenuOpen: (anchorEl: HTMLElement, fileName: string) => void
  loadFiles: (path: string, startIndex: number, stopIndex: number) => Promise<boolean>
}

const FileGrid = forwardRef(({ sideBarCollapsed, currentDirectory, directoryContents, selectedFile, refreshFiles, onFileClick, onFileDoubleClick, onFileMenuOpen, loadFiles, ...props }: FileGridProps, ref: ForwardedRef<InfiniteLoader>) => {
  // We don't know ahead of time how many files there are.
  const itemCountInfinity = 1000000

  // Because its an associative array, we can't just use the length.
  const directoryContentsSize = useMemo((): number => {
    let highestIndex = -Infinity
    directoryContents.forEach((_, index) => {
      if (index > highestIndex) {
        highestIndex = index
      }
    })

    return highestIndex === -Infinity ? 0 : highestIndex + 1
  }, [directoryContents])

  const isItemLoaded = useCallback((index: number): boolean => {
    return directoryContents.get(index) !== undefined
  }, [directoryContents])

  const loadMoreItems = useCallback(async (startIndex: number, stopIndex: number): Promise<void> => {
    if (currentDirectory !== undefined) {
      await loadFiles(currentDirectory, startIndex, stopIndex)
    }
  }, [currentDirectory])

  // Trigger a refresh of the files when the current directory changes.
  useEffect(() => {
    if (currentDirectory !== undefined) {
      refreshFiles()
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

          if (!sideBarCollapsed && gutterSize < 12) {
            columnCount--

            totalGuttersWidth = width - (columnCount * desiredColumnWidth)
            gutterSize = columnCount > 1 ? totalGuttersWidth / (columnCount - 1) : 0
          }

          return (
            <InfiniteLoader
              ref={ref}
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
                    rowCount={Math.ceil(directoryContentsSize / columnCount) + 1}
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
})
FileGrid.displayName = 'FileGrid'

const FileGridCell: React.FC<GridChildComponentProps> = ({
  columnIndex,
  rowIndex,
  style
}: GridChildComponentProps) => {
  const { columnCount, gutterSize, directoryContents, selectedFile, onFileClick, onFileDoubleClick, onFileMenuOpen } = useContext(FileGridContext)

  const index = rowIndex * columnCount + columnIndex
  const fileInfo = directoryContents.get(index)
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
  directoryContents: Map<number, FileInfo>
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
  directoryContents: new Map(),
  onFileClick: () => { },
  onFileDoubleClick: () => { },
  onFileMenuOpen: () => { }
})

const FileGridProvider: React.FC<FileGridProviderProps> = ({ children, value }) => {
  return <FileGridContext.Provider value={value}>{children}</FileGridContext.Provider>
}

export default FileGrid
