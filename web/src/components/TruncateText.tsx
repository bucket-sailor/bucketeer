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

interface TruncateTextProps {
  text: string
  containerEl: HTMLElement | null
  containerHeight: number
  containerWidth: number
}

const TruncateText: React.FC<TruncateTextProps> = ({ text, containerEl, containerHeight, containerWidth }) => {
  if (containerEl === null) {
    return ''
  }

  const style = window.getComputedStyle(containerEl)
  const fontSize = parseFloat(style.fontSize)
  const fontFamily = style.fontFamily

  let lineHeight = parseFloat(style.lineHeight)
  if (isNaN(lineHeight)) {
    // Fallback if lineHeight is 'normal'
    lineHeight = fontSize * 1.2
  }
  const charWidth = estimateCharWidth(fontSize, fontFamily)

  const maxCharsPerLine = Math.floor(containerWidth / charWidth)
  const maxLines = Math.floor(containerHeight / lineHeight)
  const totalMaxChars = maxCharsPerLine * maxLines

  if (text.length <= totalMaxChars) {
    return text
  } else {
    const ellipsis = '...'
    const totalMaxCharsWithEllipsis = totalMaxChars - ellipsis.length
    const partLength = Math.floor(totalMaxCharsWithEllipsis / 2)
    const leftPart = text.substring(0, partLength)
    const rightPart = text.substring(text.length - partLength)

    return leftPart + ellipsis + rightPart
  }
}

const charWidthCache: Record<string, number> = {}

const estimateCharWidth = (fontSize: number, fontFamily: string): number => {
  const cacheKey = `${fontFamily}-${fontSize}`
  if (charWidthCache[cacheKey] !== undefined) {
    return charWidthCache[cacheKey]
  }

  const span = document.createElement('span')
  span.style.fontFamily = fontFamily
  span.style.fontSize = `${fontSize}px`
  span.style.visibility = 'hidden'
  span.textContent = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'

  document.body.appendChild(span)
  const width = (span.offsetWidth / span.textContent.length) * 1.05
  document.body.removeChild(span)

  charWidthCache[cacheKey] = width
  return width
}

export default TruncateText
