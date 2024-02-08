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

import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import type React from 'react'
import { useCallback, useRef, useState } from 'react'
import { type FileInfo } from '../gen/filesystem/v1alpha1/filesystem_pb'
import { Filesystem } from '../gen/filesystem/v1alpha1/filesystem_connect'
import UploadClient from '../upload/Client'
import type InfiniteLoader from 'react-window-infinite-loader'
import { generateID } from '../util/GenerateID'

export interface UseFileManagementProps {
  baseURL: string
  fileGridLoaderRef: React.MutableRefObject<InfiniteLoader | null>
}

export interface UseFileManagement {
  directoryContents: Map<number, FileInfo>
  error: Error | undefined
  refreshFiles: () => void
  loadFiles: (path: string, startIndex: number, stopIndex: number) => Promise<boolean>
  getFileInfo: (path: string) => Promise<FileInfo | undefined>
  deleteFile: (path: string) => Promise<void>
  makeDirectory: (path: string) => Promise<void>
  uploadFile: (path: string) => Promise<void>
  downloadFile: (path: string) => void
}

export const useFileManagement = ({ baseURL, fileGridLoaderRef }: UseFileManagementProps): UseFileManagement => {
  const [directoryContents, setDirectoryContents] = useState<Map<number, FileInfo>>(new Map())
  const [error, setError] = useState<Error | undefined>(undefined)

  const loadFilesListIDRef = useRef<string | undefined>(undefined)

  const filesystemClient = createPromiseClient(Filesystem, createConnectTransport({ baseUrl: baseURL + '/api' }))
  const uploadClient = new UploadClient(baseURL)

  const refreshFiles = useCallback(() => {
    // Reset the list ID.
    loadFilesListIDRef.current = undefined

    // Reset the error.
    setError(undefined)

    // Reset the directory contents.
    setDirectoryContents(new Map())

    if (fileGridLoaderRef.current !== null) {
      fileGridLoaderRef.current.resetloadMoreItemsCache(true)
    }
  }, [fileGridLoaderRef])

  const loadFiles = useCallback(async (path: string, startIndex: number, stopIndex: number): Promise<boolean> => {
    try {
      // Generate a new list ID if we don't have one.
      if (loadFilesListIDRef.current === undefined) {
        loadFilesListIDRef.current = generateID(16)
      }

      const req = {
        id: loadFilesListIDRef.current ?? '',
        path,
        startIndex: BigInt(startIndex),
        stopIndex: BigInt(stopIndex)
      }

      const response = await filesystemClient.readDir(req)

      // Is the request stale?
      if (response.id !== loadFilesListIDRef.current) {
        return true
      }

      const filesCount = response.files?.length ?? 0
      if (filesCount !== 0) {
        setDirectoryContents((prev) => {
          const next = new Map(prev)

          response.files?.forEach((file, _) => {
            if (!next.has(Number(file.index)) && file.fileInfo !== undefined) {
              next.set(Number(file.index), file.fileInfo)
            }
          })

          return next
        })
      }

      // Are we done loading files?
      return filesCount < ((stopIndex + 1) - startIndex)
    } catch (e) {
      if (e instanceof Error && e.name !== 'AbortError') {
        setError(new Error(`Failed to load files: ${e.message}`))
      }

      return true
    }
  }, [filesystemClient])

  const getFileInfo = useCallback(async (path: string): Promise<FileInfo | undefined> => {
    try {
      return await filesystemClient.stat({ value: path })
    } catch (e) {
      setError(new Error(`Failed to get file info.: ${String(e)}`))
    }
  }, [filesystemClient])

  const deleteFile = useCallback(async (path: string) => {
    try {
      await filesystemClient.removeAll({ value: path })

      refreshFiles()
    } catch (e) {
      setError(new Error(`Failed to delete file.: ${String(e)}`))
    }
  }, [filesystemClient])

  const makeDirectory = useCallback(async (path: string) => {
    try {
      await filesystemClient.mkdirAll({ value: path })

      refreshFiles()
    } catch (e) {
      setError(new Error(`Failed to make directory.: ${String(e)}`))
    }
  }, [filesystemClient])

  const downloadFile = useCallback((path: string) => {
    const a = document.createElement('a')
    a.href = encodeURI(`${baseURL}/files/download/${path}`)
    a.setAttribute('download', '')
    a.style.display = 'none'

    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }, [baseURL])

  const uploadFile = useCallback(async (directory: string) => {
    await new Promise<void>((resolve, reject) => {
      const input = document.createElement('input')
      input.type = 'file'
      input.style.display = 'none'
      input.onchange = async (event) => {
        if (event.target === null) {
          document.body.removeChild(input)
          reject(new Error('No input target found'))
          return
        }

        const file = (event.target as HTMLInputElement).files?.[0]
        if (file === undefined) {
          document.body.removeChild(input)
          reject(new Error('No file selected'))
          return
        }

        const path = (directory !== '' ? directory + '/' : '') + file.name
        uploadClient.upload(path, file).then(() => {
          refreshFiles()
          resolve()
        }).catch((e) => {
          reject(e)
        }).finally(() => {
          document.body.removeChild(input)
        })
      }

      document.body.appendChild(input)
      input.click()
    })
  }, [uploadClient])

  return { directoryContents, error, refreshFiles, loadFiles, getFileInfo, deleteFile, makeDirectory, uploadFile, downloadFile }
}
