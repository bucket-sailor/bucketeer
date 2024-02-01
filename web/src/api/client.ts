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

import pRetry, { AbortError } from 'p-retry'
import isNetworkError from 'is-network-error'

export interface FileInfo {
  name: string
  isDir: boolean
  size?: number
  lastModified?: string
}

export interface ErrorResponse {
  message: string
}

export interface ListResponse {
  id: string
  files?: FileInfo[]
}

export interface ClientOptions {
  maxRetryAttempts?: number
  requestTimeout?: number
}

interface FetchWithRetryOptions extends RequestInit {
  abortController?: AbortController
  maxRetryAttempts?: number
  requestTimeout?: number
}

class Client {
  private readonly baseUrl: string
  private readonly opts: ClientOptions

  constructor (baseUrl: string, opts: ClientOptions = {}) {
    this.baseUrl = baseUrl
    this.opts = opts
  }

  async info (path: string): Promise<FileInfo> {
    const response = await this.fetchWithRetry(`${this.baseUrl}/api/v1alpha1/fs/info?path=${encodeURIComponent(path)}`)
    return await this.handleResponse<FileInfo>(response)
  }

  async list (path: string, startIndex: number, stopIndex: number, id?: string, abortController?: AbortController): Promise<ListResponse> {
    let url = `${this.baseUrl}/api/v1alpha1/fs/list?path=${encodeURIComponent(path)}&startIndex=${startIndex}&stopIndex=${stopIndex}`
    if (id !== undefined) {
      url = url.concat(`&id=${id}`)
    }

    const response = await this.fetchWithRetry(url, { abortController })
    return await this.handleResponse<ListResponse>(response)
  }

  async mkdir (path: string): Promise<void> {
    const body = new URLSearchParams({ path })
    const response = await this.fetchWithRetry(`${this.baseUrl}/api/v1alpha1/fs/mkdir`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: body.toString()
    })
    await this.handleResponse(response)
  }

  async remove (path: string): Promise<void> {
    const body = new URLSearchParams({ path })
    const response = await this.fetchWithRetry(`${this.baseUrl}/api/v1alpha1/fs/remove`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: body.toString()
    })
    await this.handleResponse(response)
  }

  async rename (oldPath: string, newPath: string): Promise<void> {
    const body = new URLSearchParams({ oldPath, newPath })
    const response = await this.fetchWithRetry(`${this.baseUrl}/api/v1alpha1/fs/rename`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: body.toString()
    })
    await this.handleResponse(response)
  }

  private async fetchWithRetry (url: string, opts?: FetchWithRetryOptions): Promise<Response> {
    const { maxRetryAttempts = this.opts.maxRetryAttempts, requestTimeout = this.opts.requestTimeout, ...fetchOptions } = opts ?? {}

    const doRequest = async (): Promise<Response> => {
      const abortController = opts?.abortController ?? new AbortController()

      let id: ReturnType<typeof setTimeout> | undefined
      if (requestTimeout !== undefined) {
        id = setTimeout(() => { abortController.abort() }, requestTimeout)
      }

      try {
        const response = await fetch(url, { ...fetchOptions, signal: abortController.signal })
        if (!response.ok) {
          await this.handleResponse(response)
        }
        return response
      } catch (error) {
        if (isNetworkError(error)) {
          throw new Error('Network error: Unable to connect to the server')
        } else {
          throw error
        }
      } finally {
        if (id !== undefined) {
          clearTimeout(id)
        }
      }
    }

    if (maxRetryAttempts !== undefined) {
      return await pRetry(doRequest, { retries: maxRetryAttempts })
    }

    return await doRequest()
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const errorBody: ErrorResponse = await response.json().catch(() => ({ message: 'Failed to parse error response' }))
      const error = new Error(errorBody.message)
      if (response.status >= 400 && response.status < 500) {
        throw new AbortError(error.message)
      }
      throw error
    }
    return await response.json().catch(() => {})
  }
}

export default Client
