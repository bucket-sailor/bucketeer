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

import UploadClient from 'windlass'

class Client {
  private readonly url: string
  private readonly uploadClient: UploadClient

  constructor (url: string) {
    this.url = url
    this.uploadClient = new UploadClient(url)
  }

  public download (path: string): void {
    const a = document.createElement('a')
    a.href = `${this.url}/${path}`
    a.setAttribute('download', '')
    a.style.display = 'none'

    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  public async upload (directory: string): Promise<File> {
    return await new Promise<File>((resolve, reject) => {
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

        this.uploadClient.upload(`${directory}/${file.name}`, file).then(() => {
          resolve(file)
        }).catch((error) => {
          reject(error)
        }).finally(() => {
          document.body.removeChild(input)
        })
      }

      document.body.appendChild(input)
      input.click()
    })
  }
}

export default Client
