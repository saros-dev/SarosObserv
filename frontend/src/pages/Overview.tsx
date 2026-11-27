import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { api, RequestOverview } from '../api/client'

const formatTime = (time: string) =>
  new Date(time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

const Overview = () => {
  const [data, setData] = useState<RequestOverview>()
  const [error, setError] = useState('')
  const [refresh, setRefresh] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setError('')
    api.getRequestOverview(controller.signal)
      .then(setData)
      .catch((err: Error) => {
        if (!controller.signal.aborted) setError(err.message)
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [refresh])

  const totals = data?.buckets.reduce(
    (sum, bucket) => ({ requests: sum.requests + bucket.requests, errors: sum.errors + bucket.errors }),
    { requests: 0, errors: 0 },
  )
  const maxRequests = Math.max(1, ...(data?.buckets.map(bucket => bucket.requests) ?? []))

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Overview</h1>
          <p className="mt-2 text-sm text-gray-600">OpenTelemetry server spans · Last 60 minute buckets</p>
        </div>
        <button type="button" onClick={() => setRefresh(value => value + 1)} disabled={loading}
          className="inline-flex items-center gap-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50">
          <RefreshCw size={16} aria-hidden="true" />
          Refresh
        </button>
      </header>
      {loading && !data && <p role="status">Loading request activity…</p>}
      {error && <p role="alert" className="card text-red-700">{error}</p>}
      {data && totals && (
        <>
          <dl className="grid gap-4 sm:grid-cols-3">
            <div className="card"><dt className="text-sm text-gray-600">Requests</dt><dd className="mt-2 text-3xl font-semibold">{totals.requests.toLocaleString()}</dd></div>
            <div className="card"><dt className="text-sm text-gray-600">Failed requests</dt><dd className="mt-2 text-3xl font-semibold text-red-700">{totals.errors.toLocaleString()}</dd></div>
            <div className="card"><dt className="text-sm text-gray-600">Error rate</dt><dd className="mt-2 text-3xl font-semibold">{totals.requests ? `${(100 * totals.errors / totals.requests).toFixed(1)}%` : '—'}</dd></div>
          </dl>
          <section className="card space-y-3" aria-labelledby="requests-heading">
            <div className="flex flex-wrap justify-between gap-2">
              <h2 id="requests-heading" className="text-lg font-semibold">Requests per minute</h2>
              <div className="flex gap-4 text-sm text-gray-600">
                <span><span className="mr-1 text-blue-600" aria-hidden="true">■</span>Other requests</span>
                <span><span className="mr-1 text-red-600" aria-hidden="true">■</span>Errors</span>
              </div>
            </div>
            {totals.requests === 0 ? <p className="text-sm text-gray-600">No server spans in this window.</p> : (
              <figure>
                <div className="flex h-40 items-end gap-px border-b border-gray-300 bg-gray-50 px-1">
                  {data.buckets.map(bucket => {
                    const label = `${formatTime(bucket.start)} · ${bucket.requests} requests · ${bucket.errors} errors`
                    return (
                      <div key={bucket.start} title={label} aria-label={label} role="img"
                        tabIndex={bucket.requests ? 0 : undefined}
                        className="flex h-full min-w-0 flex-1 flex-col justify-end focus:bg-blue-100">
                        <div className="bg-blue-500" style={{ height: `${100 * (bucket.requests - bucket.errors) / maxRequests}%` }} />
                        <div className="bg-red-600" style={{ height: `${100 * bucket.errors / maxRequests}%` }} />
                      </div>
                    )
                  })}
                </div>
                <figcaption className="mt-2 flex justify-between text-xs text-gray-500">
                  <span>{formatTime(data.buckets[0].start)}</span>
                  <span>{formatTime(data.buckets[data.buckets.length - 1].start)}</span>
                </figcaption>
              </figure>
            )}
          </section>
          <section className="card overflow-x-auto">
            <h2 className="mb-3 text-lg font-semibold">Recent failed requests</h2>
            {data.failed.length === 0 ? <p className="text-sm text-gray-600">No failed requests in this window.</p> : (
              <table className="w-full text-left text-sm">
                <thead className="text-gray-500">
                  <tr><th scope="col" className="py-2">Time</th><th scope="col">Service</th><th scope="col">Operation</th><th scope="col">Trace</th></tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {data.failed.map(request => (
                    <tr key={`${request.trace_id}-${request.span_id}`}>
                      <td className="whitespace-nowrap py-3 pr-4">{formatTime(request.start)}</td>
                      <td className="pr-4">{request.service_name}</td>
                      <td className="pr-4">{request.operation}</td>
                      <td><Link to={`/traces/${request.trace_id}`} aria-label={`Open trace ${request.trace_id} for ${request.operation}`}
                        className="font-mono font-medium text-primary-700 underline">{request.trace_id.slice(0, 12)}</Link></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </>
      )}
    </div>
  )
}

export default Overview
