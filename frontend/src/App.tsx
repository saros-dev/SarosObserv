import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Overview from './pages/Overview'
import Services from './pages/Services'
import Traces from './pages/Traces'
import TraceDetail from './pages/TraceDetail'

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/services" element={<Services />} />
          <Route path="/traces" element={<Traces />} />
          <Route path="/traces/:traceId" element={<TraceDetail />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
