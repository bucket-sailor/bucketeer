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

import { createPromiseClient, type PromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import PQueue from 'p-queue'
import pRetry, { AbortError } from 'p-retry'
import isNetworkError from 'is-network-error'
import { Upload } from '../gen/upload/v1alpha1/upload_connect'
import { CompletionStatus } from '../gen/upload/v1alpha1/upload_pb'
import { createXXHash64 } from 'hash-wasm'

export interface ClientOptions {
  numConnections?: number
  chunkSizeBytes?: number
  maxRetryAttempts?: number
  requestTimeout?: number
  onProgress?: (uploadedBytes: number, totalBytes: number) => void
}

interface FetchWithRetryOptions extends RequestInit {
  maxRetryAttempts?: number
  requestTimeout?: number
}

class Client {
  private readonly baseURL: string
  private readonly opts: ClientOptions
  private readonly apiClient: PromiseClient<typeof Upload>

  constructor (baseURL: string, opts: ClientOptions = {}) {
    const transport = createConnectTransport({
      baseUrl: baseURL + '/api',
      defaultTimeoutMs: opts.requestTimeout
    })

    this.apiClient = createPromiseClient(Upload, transport)
    this.baseURL = baseURL
    this.opts = opts
  }

  // Upload a file to the server.
  async upload (path: string, file: File): Promise<void> {
    const checksum = await this.checksum(file)

    const uploadIDResp = await this.apiClient.new({
      path,
      size: BigInt(file.size),
      checksum
    })

    const uploadID = uploadIDResp.value

    await this.uploadChunks(uploadID, file)

    await this.apiClient.complete({ value: uploadID })

    await this.pollForCompletion(uploadID)
  }

  private async uploadChunks (uploadID: string, file: File): Promise<void> {
    const fileSize = file.size
    const chunkSizeBytes = this.opts.chunkSizeBytes ?? 16000000 // 16 MB
    const numConnections = this.opts.numConnections ?? 4 // Number of concurrent uploads.

    const queue = new PQueue({ concurrency: numConnections })

    let uploadedBytes: number = 0
    const errors: Error[] = []

    const tasks = []
    for (let start = 0; start < fileSize; start += chunkSizeBytes) {
      const end = Math.min(start + chunkSizeBytes, fileSize)
      const chunk = file.slice(start, end)

      tasks.push(async () => {
        try {
          await this.uploadChunk(uploadID, chunk, start, end, fileSize)

          uploadedBytes += chunk.size

          if (this.opts.onProgress !== undefined) {
            this.opts.onProgress(uploadedBytes, fileSize)
          }
        } catch (error) {
          errors.push(error as Error)
        }
      })
    }

    await queue.addAll(tasks)

    if (errors.length > 0) {
      throw new Error(`Failed to upload ${errors.length} chunks`)
    }
  }

  private async uploadChunk (uploadID: string, chunk: Blob, start: number, end: number, size: number): Promise<void> {
    const formData = new FormData()
    formData.append('file', chunk, uploadID)

    await this.fetchWithRetry(`${this.baseURL}/files/upload`, {
      method: 'PATCH',
      headers: {
        'Content-Range': `bytes ${start}-${end - 1}/${size}`
      },
      body: formData,
      maxRetryAttempts: this.opts.maxRetryAttempts
    })
  }

  private async pollForCompletion (uploadID: string): Promise<void> {
    const response = await this.apiClient.pollForCompletion({ value: uploadID })

    switch (response.status) {
      case CompletionStatus.COMPLETED:
        return
      case CompletionStatus.FAILED:
        throw new Error('Upload failed: ' + response.error)
      case CompletionStatus.PENDING:
      { await new Promise((resolve) => setTimeout(resolve, 1000)).then(async () => { await this.pollForCompletion(uploadID) }) }
    }
  }

  private async checksum (file: File): Promise<string> {
    const chunkSize = 1024 * 1024 // 1 MB chunks

    const readChunk = async (chunk: Blob): Promise<ArrayBuffer> => {
      return await new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = () => { resolve(reader.result as ArrayBuffer) }
        reader.onerror = () => { reject(new Error('Error reading file chunk.')) }
        reader.readAsArrayBuffer(chunk)
      })
    }

    const h = await createXXHash64()
    h.init()

    let currentOffset = 0
    while (currentOffset < file.size) {
      const nextChunkEnd = Math.min(currentOffset + chunkSize, file.size)
      const chunk = file.slice(currentOffset, nextChunkEnd)
      const arrayBuffer = await readChunk(chunk)
      h.update(new Uint8Array(arrayBuffer))
      currentOffset += chunk.size
    }

    return 'xxh64:' + h.digest('hex')
  }

  private async fetchWithRetry (url: string, opts?: FetchWithRetryOptions): Promise<Response> {
    const { maxRetryAttempts = this.opts.maxRetryAttempts, requestTimeout = this.opts.requestTimeout, ...fetchOptions } = opts ?? {}

    const doRequest = async (): Promise<Response> => {
      const controller = new AbortController()

      let id: ReturnType<typeof setTimeout> | undefined
      if (requestTimeout !== undefined) {
        id = setTimeout(() => { controller.abort() }, requestTimeout)
      }

      try {
        const response = await fetch(url, { ...fetchOptions, signal: controller.signal })
        return await this.handleResponse(response)
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

  private async handleResponse (response: Response): Promise<Response> {
    if (!response.ok) {
      const message = await response.text()
      if (response.status >= 400 && response.status < 500) {
        throw new AbortError(message)
      }
      throw new Error(message)
    }
    return response
  }
}

export default Client
