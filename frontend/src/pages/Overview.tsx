import { Activity, AlertCircle, Server, Clock } from 'lucide-react'

const Overview = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Overview</h1>
        <p className="mt-2 text-gray-600">Application health and telemetry summary</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Total Requests</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
            </div>
            <Activity className="h-8 w-8 text-primary-600" />
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Error Rate</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
            </div>
            <AlertCircle className="h-8 w-8 text-red-600" />
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">P95 Latency</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
            </div>
            <Clock className="h-8 w-8 text-yellow-600" />
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Active Services</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
            </div>
            <Server className="h-8 w-8 text-green-600" />
          </div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Getting Started</h2>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700 mb-4">
            SarosObserv is ready to receive OpenTelemetry telemetry. Configure your application to export traces:
          </p>
          <pre className="bg-gray-800 text-gray-100 p-4 rounded-lg overflow-x-auto">
            <code>OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317</code>
          </pre>
          <p className="mt-4 text-sm text-gray-600">
            See the documentation for language-specific configuration examples.
          </p>
        </div>
      </div>
    </div>
  )
}

export default Overview
