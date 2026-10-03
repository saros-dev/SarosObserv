import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { GitBranch, Activity, Clock, ChevronRight, ChevronDown } from 'lucide-react'
import { api, Span } from '../api/client'

const TraceDetail = () => {
  const { traceId } = useParams<{ traceId: string }>()
  const [spans, setSpans] = useState<Span[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedSpans, setExpandedSpans] = useState<Set<string>>(new Set())

  useEffect(() => {
    const fetchTrace = async () => {
      if (!traceId) return
      try {
        const data = await api.getTrace(traceId)
        setSpans(data.spans)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load trace')
      } finally {
        setLoading(false)
      }
    }

    fetchTrace()
  }, [traceId])

  const toggleSpan = (spanId: string) => {
    const newExpanded = new Set(expandedSpans)
    if (newExpanded.has(spanId)) {
      newExpanded.delete(spanId)
    } else {
      newExpanded.add(spanId)
    }
    setExpandedSpans(newExpanded)
  }

  const formatDuration = (ns: number) => {
    if (ns < 1000) return `${ns}ns`
    if (ns < 1000000) return `${(ns / 1000).toFixed(2)}µs`
    if (ns < 1000000000) return `${(ns / 1000000).toFixed(2)}ms`
    return `${(ns / 1000000000).toFixed(2)}s`
  }

  const formatTime = (time: string) => {
    return new Date(time).toLocaleString()
  }

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'ok':
        return 'bg-green-100 text-green-800'
      case 'error':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Trace Detail</h1>
        </div>
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Activity className="h-12 w-12 text-gray-400 mx-auto mb-4 animate-pulse" />
              <p className="text-gray-500">Loading trace...</p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Trace Detail</h1>
        </div>
        <div className="card bg-red-50 border-red-200">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <GitBranch className="h-12 w-12 text-red-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-red-900">Error loading trace</h3>
              <p className="mt-2 text-red-600">{error}</p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Trace Detail</h1>
        <p className="mt-2 text-gray-600 font-mono text-sm">{traceId}</p>
      </div>

      {spans.length === 0 ? (
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <GitBranch className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900">No spans found</h3>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {spans.map((span) => (
            <div key={span.SpanID} className="card">
              <div
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                onClick={() => toggleSpan(span.SpanID)}
              >
                <div className="flex items-center space-x-4 flex-1">
                  {expandedSpans.has(span.SpanID) ? (
                    <ChevronDown className="h-5 w-5 text-gray-400" />
                  ) : (
                    <ChevronRight className="h-5 w-5 text-gray-400" />
                  )}
                  <div className="flex-1">
                    <div className="flex items-center space-x-3">
                      <span className="font-medium text-gray-900">{span.ServiceName}</span>
                      <span className="text-gray-500">{span.OperationName}</span>
                      <span className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor(span.Status)}`}>
                        {span.Status}
                      </span>
                    </div>
                    <div className="flex items-center space-x-4 mt-2 text-sm text-gray-500">
                      <div className="flex items-center">
                        <Clock className="h-4 w-4 mr-1" />
                        {formatTime(span.StartTime)}
                      </div>
                      <div>Duration: {formatDuration(span.DurationNs)}</div>
                      {span.ParentSpanID && (
                        <div className="font-mono text-xs">
                          Parent: {span.ParentSpanID.slice(0, 8)}...
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
              {expandedSpans.has(span.SpanID) && (
                <div className="border-t border-gray-200 p-4 bg-gray-50">
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="font-medium text-gray-700">Span ID:</span>
                      <span className="ml-2 font-mono text-gray-600">{span.SpanID}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-700">Trace ID:</span>
                      <span className="ml-2 font-mono text-gray-600">{span.TraceID.slice(0, 16)}...</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-700">Start Time:</span>
                      <span className="ml-2 text-gray-600">{formatTime(span.StartTime)}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-700">End Time:</span>
                      <span className="ml-2 text-gray-600">{formatTime(span.EndTime)}</span>
                    </div>
                  </div>
                  {span.StatusMessage && (
                    <div className="mt-4">
                      <span className="font-medium text-gray-700">Status Message:</span>
                      <span className="ml-2 text-gray-600">{span.StatusMessage}</span>
                    </div>
                  )}
                  {Object.keys(span.Attributes).length > 0 && (
                    <div className="mt-4">
                      <span className="font-medium text-gray-700 block mb-2">Attributes:</span>
                      <div className="bg-white border border-gray-200 rounded p-3 max-h-48 overflow-y-auto">
                        <table className="min-w-full">
                          <tbody>
                            {Object.entries(span.Attributes).map(([key, value]) => (
                              <tr key={key} className="border-b border-gray-100 last:border-0">
                                <td className="py-1 pr-4 font-mono text-xs text-gray-500">{key}</td>
                                <td className="py-1 text-sm text-gray-900">{value}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default TraceDetail
