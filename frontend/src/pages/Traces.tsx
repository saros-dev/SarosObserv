import { useEffect, useState } from 'react'
import { GitBranch, Activity, Clock } from 'lucide-react'
import { api, TraceSummary } from '../api/client'
import { Link } from 'react-router-dom'

const Traces = () => {
  const [traces, setTraces] = useState<TraceSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchTraces = async () => {
      try {
        const data = await api.getTraces(50)
        setTraces(data.traces)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load traces')
      } finally {
        setLoading(false)
      }
    }

    fetchTraces()
  }, [])

  const formatDuration = (ns: number) => {
    if (ns < 1000) return `${ns}ns`
    if (ns < 1000000) return `${(ns / 1000).toFixed(2)}µs`
    if (ns < 1000000000) return `${(ns / 1000000).toFixed(2)}ms`
    return `${(ns / 1000000000).toFixed(2)}s`
  }

  const formatTime = (time: string) => {
    return new Date(time).toLocaleString()
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Traces</h1>
        <p className="mt-2 text-gray-600">Distributed trace explorer</p>
      </div>

      {loading ? (
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Activity className="h-12 w-12 text-gray-400 mx-auto mb-4 animate-pulse" />
              <p className="text-gray-500">Loading traces...</p>
            </div>
          </div>
        </div>
      ) : error ? (
        <div className="card bg-red-50 border-red-200">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <GitBranch className="h-12 w-12 text-red-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-red-900">Error loading traces</h3>
              <p className="mt-2 text-red-600">{error}</p>
            </div>
          </div>
        </div>
      ) : traces.length === 0 ? (
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <GitBranch className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900">No traces available</h3>
              <p className="mt-2 text-gray-500">
                Traces will appear here when applications send telemetry to SarosObserv.
              </p>
            </div>
          </div>
        </div>
      ) : (
        <div className="card">
          <table className="min-w-full divide-y divide-gray-200">
            <thead>
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Trace ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Service
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Operation
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Start Time
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Duration
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Spans
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {traces.map((trace) => (
                <tr key={trace.trace_id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      to={`/traces/${trace.trace_id}`}
                      className="text-sm font-medium text-blue-600 hover:text-blue-900 font-mono"
                    >
                      {trace.trace_id.slice(0, 16)}...
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {trace.service_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {trace.operation_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <div className="flex items-center">
                      <Clock className="h-4 w-4 text-gray-400 mr-2" />
                      {formatTime(trace.start_time)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
                    {formatDuration(trace.duration_ns)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
                    {trace.span_count}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

export default Traces
