import { useEffect, useState } from 'react'
import { Server, Activity } from 'lucide-react'
import { api, Service } from '../api/client'

const Services = () => {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchServices = async () => {
      try {
        const data = await api.getServices()
        setServices(data.services)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load services')
      } finally {
        setLoading(false)
      }
    }

    fetchServices()
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Services</h1>
        <p className="mt-2 text-gray-600">Discovered services and their health</p>
      </div>

      {loading ? (
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Activity className="h-12 w-12 text-gray-400 mx-auto mb-4 animate-pulse" />
              <p className="text-gray-500">Loading services...</p>
            </div>
          </div>
        </div>
      ) : error ? (
        <div className="card bg-red-50 border-red-200">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Server className="h-12 w-12 text-red-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-red-900">Error loading services</h3>
              <p className="mt-2 text-red-600">{error}</p>
            </div>
          </div>
        </div>
      ) : services.length === 0 ? (
        <div className="card">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900">No services discovered yet</h3>
              <p className="mt-2 text-gray-500">
                Services will be automatically discovered when telemetry is received.
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
                  Service Name
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Span Count
                </th>
              </tr>.
            </thead>
            <tbody className="divide-y divide-gray-200">
              {services.map((service) => (
                <tr key={service.name} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <Server className="h-5 w-5 text-gray-400 mr-3" />
                      <span className="text-sm font-medium text-gray-900">{service.name}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
                    {service.span_count.toLocaleString()}
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

export default Services
