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

import { beforeEach, describe, expect, it, type jest } from '@jest/globals'
import fetchMock from 'jest-fetch-mock'
import Client from './client'
import { TextDecoder, TextEncoder } from 'util'
import { Empty, StringValue } from '@bufbuild/protobuf'
import { CompleteResponse, CompletionStatus } from '../gen/upload/v1alpha1/upload_pb'

global.TextEncoder = TextEncoder
global.TextDecoder = TextDecoder

fetchMock.enableMocks()

describe('Upload Client', () => {
  const baseURL = 'http://example.com'

  const client = new Client(baseURL, {
    chunkSizeBytes: 256
  })

  beforeEach(() => {
    fetchMock.resetMocks()

    let pollCounter = 0

    fetchMock.mockIf(/^http:\/\/example.com/, async (req) => {
      if (req.url === 'http://example.com/api/bucketeer.upload.v1alpha1.Upload/New') {
        const resp = new StringValue()
        resp.value = 'ba700fa9-0ea5-4071-9b3a-42f55597c12b'

        return {
          status: 200,
          body: resp.toJsonString()
        }
      }

      if (req.url === 'http://example.com/api/bucketeer.upload.v1alpha1.Upload/Complete') {
        return {
          status: 200,
          body: new Empty().toJsonString()
        }
      }

      if (req.url === 'http://example.com/api/bucketeer.upload.v1alpha1.Upload/PollForCompletion') {
        const resp = new CompleteResponse()
        resp.status = CompletionStatus.PENDING

        pollCounter++

        if (pollCounter >= 3) {
          resp.status = CompletionStatus.COMPLETED
        }

        return {
          status: 200,
          body: resp.toJsonString()
        }
      }

      if (req.url === 'http://example.com/files/upload') {
        return {
          status: 200
        }
      }
    })
  })

  it('successfully uploads a file', async () => {
    // fill with a known pattern
    const data = new Uint8Array(1024)
    for (let i = 0; i < data.length; i++) {
      data[i] = i % 256
    }

    const file = new File([data], 'test.bin')

    await client.upload('/test.bin', file)

    expect(fetch).toHaveBeenCalledTimes(9)

    const calls = (fetch as jest.Mock).mock.calls as unknown as Array<[string, RequestInit]>

    // First call should be to begin the upload.
    let [url, { method, body }] = calls[0]

    expect(url).toBe('http://example.com/api/bucketeer.upload.v1alpha1.Upload/New')
    expect(method).toBe('POST')
    expect(new TextDecoder().decode(body as Uint8Array)).toBe(
      '{"path":"/test.bin","size":"1024","checksum":"xxh64:6f3914f18fe4df57"}')

    // Next four calls should be to upload the chunks.
    for (let i = 1; i < 5; i++) {
      const [url, { method, headers }] = calls[i]

      expect(url).toBe('http://example.com/files/upload')
      expect(method).toBe('PATCH')
      expect(headers).toEqual(expect.objectContaining({
        'Content-Range': `bytes ${256 * (i - 1)}-${256 * i - 1}/1024`
      }))
    }

    // Next call should be to begin completion.
    [url, { method, body }] = calls[5]

    expect(url).toBe('http://example.com/api/bucketeer.upload.v1alpha1.Upload/Complete')
    expect(method).toBe('POST')
    expect(new TextDecoder().decode(body as Uint8Array)).toBe(
      '"ba700fa9-0ea5-4071-9b3a-42f55597c12b"')

    // Should poll for completion 3 times.
    for (let i = 6; i < 9; i++) {
      const [url, { method }] = calls[i]

      expect(url).toBe('http://example.com/api/bucketeer.upload.v1alpha1.Upload/PollForCompletion')
      expect(method).toBe('POST')
      expect(new TextDecoder().decode(body as Uint8Array)).toBe(
        '"ba700fa9-0ea5-4071-9b3a-42f55597c12b"')
    }
  })
})
