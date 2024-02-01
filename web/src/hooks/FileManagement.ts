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

import { useState, useRef, useCallback, useEffect } from 'react'
import PQueue from 'p-queue'
import ApiClient, { type FileInfo } from '../api/client'
import FilesClient from '../files/client'

export interface UseFileManagement {
  directoryContents: FileInfo[]
  error: Error | undefined
  resetFilesState: () => void
  loadFiles: (path: string, startIndex: number, stopIndex: number) => Promise<boolean>
  uploadFile: (path: string) => Promise<void>
  downloadFile: (path: string) => Promise<void>
  getFileInfo: (path: string) => Promise<FileInfo | undefined>
  deleteFile: (path: string) => Promise<void>
  makeDirectory: (path: string) => Promise<void>
}

export const useFileManagement = (baseURL: string): UseFileManagement => {
  const [directoryContents, setDirectoryContents] = useState<FileInfo[]>([])
  const [error, setError] = useState<Error | undefined>(undefined)

  // TODO: we shouldn't need to limit concurrency anymore.
  const loadFilesQueueRef = useRef<PQueue>(new PQueue({ concurrency: 1 }))
  const loadFilesListIDRef = useRef<string | undefined>(undefined)
  const abortControllerRef = useRef<AbortController | undefined>(undefined)

  const apiClient = new ApiClient(baseURL, { maxRetryAttempts: 3 })
  const filesClient = new FilesClient(`${baseURL}/files`)

  const resetFilesState = useCallback(() => {
    // Remove any queued up requests, and create a new queue.
    loadFilesQueueRef.current.clear()
    loadFilesQueueRef.current = new PQueue({ concurrency: 1 })

    // Abort any inflight requests.
    if (abortControllerRef.current !== undefined) {
      abortControllerRef.current.abort()
    }
    abortControllerRef.current = undefined

    // Reset the directory contents.
    loadFilesListIDRef.current = undefined
    setDirectoryContents([])
  }, [])

  const loadFiles = useCallback(async (path: string, startIndex: number, stopIndex: number): Promise<boolean> => {
    const loadFilesPage = async (): Promise<boolean> => {
      setError(undefined)

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
          // TODO: shouldn't need to do this anymore (we can use the index).
          setDirectoryContents(prev => {
            const newFiles = response?.files?.filter(newItem =>
              !prev.some(existingItem => existingItem.name === newItem.name)
            )
            if (newFiles !== undefined) {
              return [...prev, ...newFiles]
            }
            return prev
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
    setError(undefined)

    try {
      const file = await filesClient.upload(path)
      setDirectoryContents(prev => [...prev, { name: file.name, isDir: false }])
    } catch (e) {
      setError(new Error(`Failed to upload file.: ${String(e)}`))
    }
  }, [filesClient])

  const downloadFile = useCallback(async (path: string) => {
    setError(undefined)

    try {
      filesClient.download(path)
    } catch (e) {
      setError(new Error(`Failed to download file.: ${String(e)}`))
    }
  }, [filesClient])

  const getFileInfo = useCallback(async (path: string): Promise<FileInfo | undefined> => {
    setError(undefined)

    try {
      return await apiClient.info(path)
    } catch (e) {
      setError(new Error(`Failed to get file info.: ${String(e)}`))
    }
  }, [apiClient])

  const deleteFile = useCallback(async (path: string) => {
    setError(undefined)

    try {
      await apiClient.remove(path)

      const fileName = path.split('/').pop() ?? ''
      setDirectoryContents(prev => prev.filter(file => file.name !== fileName))
    } catch (e) {
      setError(new Error(`Failed to delete file.: ${String(e)}`))
    }
  }, [apiClient])

  const makeDirectory = useCallback(async (path: string) => {
    setError(undefined)

    try {
      await apiClient.mkdir(path)
      setDirectoryContents(prev => [...prev, { name: path.split('/').pop() ?? '', isDir: true }])
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

  return { directoryContents, error, resetFilesState, loadFiles, uploadFile, downloadFile, getFileInfo, deleteFile, makeDirectory }
}
