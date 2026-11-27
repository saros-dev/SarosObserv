import { useEffect, useState } from 'react'
import { api, TraceLog } from '../api/client'

export default function TraceLogs({ traceId }: { traceId: string }) {
  const [logs, setLogs] = useState<TraceLog[]>()
  const [error, setError] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    api.getTraceLogs(traceId, controller.signal)
      .then(data => setLogs(data.logs))
      .catch((err: Error) => {
        if (!controller.signal.aborted) setError(err.message)
      })
    return () => controller.abort()
  }, [traceId])

  return (
    <section className="card">
      <h2 className="mb-3 text-lg font-semibold">Related logs</h2>
      {error && <p role="alert" className="text-sm text-red-700">{error}</p>}
      {!error && !logs && <p role="status" className="text-sm text-gray-600">Loading logs…</p>}
      {logs?.length === 0 && <p className="text-sm text-gray-600">No logs with this trace ID.</p>}
      {logs && logs.length > 0 && (
        <ul className="divide-y divide-gray-200">
          {logs.map(log => (
            <li key={log.id} className="py-3">
              <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-gray-600">
                <time dateTime={log.timestamp}>{new Date(log.timestamp).toLocaleString()}</time>
                <span className="font-semibold">{log.severity || 'Unspecified'}</span>
                <span>{log.service_name}</span>
                {log.span_id && <span className="font-mono">Span {log.span_id}</span>}
              </div>
              <p className="mt-2 whitespace-pre-wrap break-words font-mono text-sm">{log.message}</p>
              {Object.keys(log.attributes).length > 0 && (
                <details className="mt-2 text-sm text-gray-600">
                  <summary className="cursor-pointer">Attributes</summary>
                  <dl className="mt-2 break-words">
                    {Object.entries(log.attributes).map(([key, value]) => (
                      <div key={key}><dt className="inline font-medium">{key}</dt>{' '}<dd className="inline font-mono">{value}</dd></div>
                    ))}
                  </dl>
                </details>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
