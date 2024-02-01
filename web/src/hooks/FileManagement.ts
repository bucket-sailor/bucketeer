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

import type React from 'react'
import { useState, useRef, useCallback, useEffect } from 'react'
import PQueue from 'p-queue'
import ApiClient, { type FileInfo } from '../api/client'
import FilesClient from '../files/client'
import type InfiniteLoader from 'react-window-infinite-loader'

export interface UseFileManagementProps {
  baseURL: string
  fileGridLoaderRef: React.MutableRefObject<InfiniteLoader | null>
}

export interface FileInfoWithIndex extends FileInfo {
  index: number
}

export interface UseFileManagement {
  directoryContents: FileInfoWithIndex[]
  error: Error | undefined
  refreshFiles: () => void
  loadFiles: (path: string, startIndex: number, stopIndex: number) => Promise<boolean>
  uploadFile: (path: string) => Promise<void>
  downloadFile: (path: string) => Promise<void>
  getFileInfo: (path: string) => Promise<FileInfo | undefined>
  deleteFile: (path: string) => Promise<void>
  makeDirectory: (path: string) => Promise<void>
}

export const useFileManagement = ({ baseURL, fileGridLoaderRef }: UseFileManagementProps): UseFileManagement => {
  const [directoryContents, setDirectoryContents] = useState<FileInfoWithIndex[]>([])
  const [error, setError] = useState<Error | undefined>(undefined)

  // TODO: we shouldn't need to limit concurrency anymore (we'll keep it in the short term as it simplifies abort handling).
  const loadFilesQueueRef = useRef<PQueue>(new PQueue({ concurrency: 1 }))
  const loadFilesListIDRef = useRef<string | undefined>(undefined)
  const abortControllerRef = useRef<AbortController | undefined>(undefined)

  const apiClient = new ApiClient(baseURL)
  const filesClient = new FilesClient(`${baseURL}/files`)

  const refreshFiles = useCallback(() => {
    // Remove any queued up requests, and create a new queue.
    loadFilesQueueRef.current.clear()
    loadFilesQueueRef.current = new PQueue({ concurrency: 1 })

    // Abort any inflight requests.
    if (abortControllerRef.current !== undefined) {
      abortControllerRef.current.abort()
    }
    abortControllerRef.current = undefined

    // Reset the list ID.
    loadFilesListIDRef.current = undefined

    // Reset the error.
    setError(undefined)

    // Reset the directory contents.
    setDirectoryContents([])

    if (fileGridLoaderRef.current !== null) {
      fileGridLoaderRef.current.resetloadMoreItemsCache(true)
    }
  }, [fileGridLoaderRef])

  const loadFiles = useCallback(async (path: string, startIndex: number, stopIndex: number): Promise<boolean> => {
    const loadFilesPage = async (): Promise<boolean> => {
      try {
        abortControllerRef.current = new AbortController()

        const response = await apiClient.list(path, startIndex, stopIndex, loadFilesListIDRef.current, abortControllerRef.current)
        if (abortControllerRef.current !== undefined && abortControllerRef.current.signal.aborted) {
          // Don't attempt to load any more files if the request was aborted.
          return true
        }

        loadFilesListIDRef.current = response.id

        const filesCount = response.files?.length ?? 0
        if (filesCount !== 0) {
          setDirectoryContents((prev) => {
            return [...prev, ...response.files?.map((file, index) => {
              return { ...file, index: startIndex + index }
            }) ?? []]
          })
        }

        // Are we done loading files?
        return filesCount < (stopIndex - startIndex)
      } catch (e) {
        if (e instanceof Error && e.name !== 'AbortError') {
          setError(new Error(`Failed to load files: ${e.message}`))
        }

        return true
      }
    }

    return await loadFilesQueueRef.current.add(loadFilesPage) as boolean
  }, [apiClient])

  const uploadFile = useCallback(async (path: string) => {
    try {
      await filesClient.upload(path)

      refreshFiles()
    } catch (e) {
      setError(new Error(`Failed to upload file.: ${String(e)}`))
    }
  }, [filesClient])

  const downloadFile = useCallback(async (path: string) => {
    try {
      filesClient.download(path)
    } catch (e) {
      setError(new Error(`Failed to download file.: ${String(e)}`))
    }
  }, [filesClient])

  const getFileInfo = useCallback(async (path: string): Promise<FileInfo | undefined> => {
    try {
      return await apiClient.info(path)
    } catch (e) {
      setError(new Error(`Failed to get file info.: ${String(e)}`))
    }
  }, [apiClient])

  const deleteFile = useCallback(async (path: string) => {
    try {
      await apiClient.remove(path)

      refreshFiles()
    } catch (e) {
      setError(new Error(`Failed to delete file.: ${String(e)}`))
    }
  }, [apiClient])

  const makeDirectory = useCallback(async (path: string) => {
    try {
      await apiClient.mkdir(path)

      refreshFiles()
    } catch (e) {
      setError(new Error(`Failed to make directory.: ${String(e)}`))
    }
  }, [apiClient])

  // Abort any inflight requests when the component unmounts.
  useEffect(() => {
    return () => {
      if (abortControllerRef.current !== undefined) {
        abortControllerRef.current.abort()
      }
    }
  }, [])

  return { directoryContents, error, refreshFiles, loadFiles, uploadFile, downloadFile, getFileInfo, deleteFile, makeDirectory }
}
