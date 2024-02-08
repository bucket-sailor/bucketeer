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

import { useCallback } from 'react'
import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { parse } from 'stacktrace-parser'
import { StackFrame, type TelemetryEvent, TelemetryEventKind } from '../gen/telemetry/v1alpha1/telemetry_pb'
import { Telemetry } from '../gen/telemetry/v1alpha1/telemetry_connect'
import { type PartialMessage, Timestamp } from '@bufbuild/protobuf'

export interface UseTelemetryProps {
  baseURL: string
  sessionID: string
}

export type Destructor = () => void

export interface UseTelemetry {
  registerErrorHandlers: () => Destructor
  reportEvent: (event: PartialMessage<TelemetryEvent>) => Promise<void>
  reportError: (error: Error) => Promise<void>
}

export const useTelemetry = ({ baseURL, sessionID }: UseTelemetryProps): UseTelemetry => {
  const telemetryClient = createPromiseClient(Telemetry, createConnectTransport({ baseUrl: baseURL + '/api' }))

  const reportEvent = useCallback(async (event: PartialMessage<TelemetryEvent>): Promise<void> => {
    event.sessionId = sessionID
    if (event.timestamp === undefined) {
      event.timestamp = Timestamp.now()
    }

    if (event.kind === undefined) {
      event.kind = TelemetryEventKind.INFO
    }

    event.tags = (event.tags ?? []).concat(['web'])

    await telemetryClient.report(event).catch((e) => {
      console.error('Error reporting event:', e)
    })
  }, [sessionID])

  const reportError = useCallback(async (error: Error): Promise<void> => {
    const stackTrace: StackFrame[] = parse(error.stack ?? '').map((frame) => {
      const stackFrame = new StackFrame()
      stackFrame.file = frame.file ?? ''
      stackFrame.function = frame.methodName ?? ''
      stackFrame.line = frame.lineNumber ?? 0
      stackFrame.column = frame.column ?? 0

      return stackFrame
    })

    await telemetryClient.report({
      sessionId: sessionID,
      timestamp: Timestamp.now(),
      kind: TelemetryEventKind.ERROR,
      name: 'Error',
      message: error.message,
      stackTrace,
      tags: ['web']
    }).catch((e) => {
      console.error('Error reporting error:', e)
    })
  }, [sessionID])

  const registerErrorHandlers = useCallback((): Destructor => {
    const errorHandler = (_event: Event | string, _source?: string, _lineno?: number, _colno?: number, error?: Error): void => {
      if (error !== undefined) {
        reportError(error).then(() => {}).catch(() => {})
      }
    }

    const unhandledRejectionHandler = (event: PromiseRejectionEvent): void => {
      if (event.reason instanceof Error) {
        errorHandler(event, undefined, undefined, undefined, event.reason)
      }
    }

    window.onerror = errorHandler
    window.addEventListener('error', errorHandler)

    window.onunhandledrejection = unhandledRejectionHandler
    window.addEventListener('unhandledrejection', unhandledRejectionHandler)

    return () => {
      window.onerror = errorHandler
      window.removeEventListener('error', errorHandler)

      window.onunhandledrejection = unhandledRejectionHandler
      window.removeEventListener('unhandledrejection', unhandledRejectionHandler)
    }
  }, [reportError])

  return {
    registerErrorHandlers,
    reportEvent,
    reportError
  }
}
